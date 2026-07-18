package flow

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeDocumentOptions struct {
	types     []TypeOption
	epics     []EpicOption
	typeCalls int
	epicCalls int
	typeErr   error
	epicErr   error
}

func (f *fakeDocumentOptions) ChangeTypes() ([]TypeOption, error) {
	f.typeCalls++
	return f.types, f.typeErr
}

func (f *fakeDocumentOptions) Epics() ([]EpicOption, error) {
	f.epicCalls++
	return f.epics, f.epicErr
}

func TestCanonicalizeDocumentNormalizesAndCanonicalizesMetadata(t *testing.T) {
	options := &fakeDocumentOptions{
		types: []TypeOption{{Slug: "feature"}, {Slug: "test"}, {Slug: "ci"}},
		epics: []EpicOption{{ID: 11, Title: " Canonical Epic "}},
	}
	document, err := CanonicalizeDocument([]byte("# Title\r\n\r\nEpic: stale title #0011\r\n\r\nTypes: ci | feature | test\r\n\r\nBody\r\n"), options)
	require.NoError(t, err)
	assert.Equal(t, "Title", document.Title)
	assert.Equal(t, "# Title\n\nTypes: feature|test|ci\n\nEpic: Canonical Epic #11\n\nBody\n", string(document.Bytes))
	assert.Equal(t, 1, options.typeCalls)
	assert.Equal(t, 1, options.epicCalls)
}

func TestCanonicalizeDocumentLoadsOnlyPresentOptions(t *testing.T) {
	options := &fakeDocumentOptions{types: []TypeOption{{Slug: "feature"}}}
	document, err := CanonicalizeDocument([]byte("# Title\n\nBody"), options)
	require.NoError(t, err)
	assert.Equal(t, "# Title\n\nBody", string(document.Bytes))
	assert.Zero(t, options.typeCalls)
	assert.Zero(t, options.epicCalls)
}

func TestCanonicalizeDocumentEmitsSelectedTypeOnceWhenOptionsRepeat(t *testing.T) {
	options := &fakeDocumentOptions{types: []TypeOption{{Slug: "feature"}, {Slug: "feature"}}}
	document, err := CanonicalizeDocument([]byte("# Title\n\nTypes: feature\n\nBody"), options)
	require.NoError(t, err)
	assert.Equal(t, "# Title\n\nTypes: feature\n\nBody", string(document.Bytes))
}

func TestCanonicalizeDocumentAcceptsEveryHeaderOrderAndPreservesBodyMetadataText(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
		types   int
		epics   int
	}{
		{name: "title", content: "# Title\n\nBody", want: "# Title\n\nBody"},
		{name: "types", content: "# Title\n\nTypes: feature\n\nBody", want: "# Title\n\nTypes: feature\n\nBody", types: 1},
		{name: "epic", content: "# Title\n\nEpic: stale #01\n\nBody", want: "# Title\n\nEpic: Epic #1\n\nBody", epics: 1},
		{name: "types then epic", content: "# Title\n\nTypes: feature\n\nEpic: stale #1\n\nBody", want: "# Title\n\nTypes: feature\n\nEpic: Epic #1\n\nBody", types: 1, epics: 1},
		{name: "epic then types", content: "# Title\n\nEpic: stale #1\n\nTypes: feature\n\nBody", want: "# Title\n\nTypes: feature\n\nEpic: Epic #1\n\nBody", types: 1, epics: 1},
		{name: "body metadata", content: "# Title\n\nBody\nTypes: untouched\nEpic: untouched #999", want: "# Title\n\nBody\nTypes: untouched\nEpic: untouched #999"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := &fakeDocumentOptions{types: []TypeOption{{Slug: "feature"}}, epics: []EpicOption{{ID: 1, Title: "Epic"}}}
			document, err := CanonicalizeDocument([]byte(test.content), options)
			require.NoError(t, err)
			assert.Equal(t, test.want, string(document.Bytes))
			assert.Equal(t, test.types, options.typeCalls)
			assert.Equal(t, test.epics, options.epicCalls)
		})
	}
}

func TestCanonicalizeDocumentRequiresNonWhitespaceBodyBeforeLoadingOptions(t *testing.T) {
	tests := []struct {
		name    string
		content string
	}{
		{name: "title only", content: "# Title\n\n"},
		{name: "whitespace body", content: "# Title\n\n \t\n"},
		{name: "Types only", content: "# Title\n\nTypes: feature\n\n"},
		{name: "Epic only", content: "# Title\n\nEpic: stale #1\n\n"},
		{name: "all metadata only", content: "# Title\n\nTypes: feature\n\nEpic: stale #1\n\n"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			options := &fakeDocumentOptions{
				types: []TypeOption{{Slug: "feature"}},
				epics: []EpicOption{{ID: 1, Title: "Epic"}},
			}
			_, err := CanonicalizeDocument([]byte(test.content), options)
			require.Error(t, err)
			var validation ValidationError
			assert.True(t, errors.As(err, &validation))
			assert.Equal(t, "document body must contain at least one non-whitespace character", err.Error())
			assert.Zero(t, options.typeCalls)
			assert.Zero(t, options.epicCalls)
		})
	}
}

func TestCanonicalizeDocumentRejectsInvalidContentAndOptionFailures(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{name: "title", content: "Title\n\n", want: "# Title parsing failed"},
		{name: "separator", content: "# Title\nBody", want: "exactly one blank line"},
		{name: "duplicate type", content: "# Title\n\nTypes: feature|feature\n\nBody", want: "duplicate type slug"},
		{name: "duplicate field", content: "# Title\n\nTypes: feature\n\nTypes: feature\n\nBody", want: "duplicate Types field"},
		{name: "empty type", content: "# Title\n\nTypes: feature||test\n\nBody", want: "empty value"},
		{name: "unknown type", content: "# Title\n\nTypes: unknown\n\nBody", want: "unknown type slug"},
		{name: "epic", content: "# Title\n\nEpic: title #nope\n\nBody", want: "invalid Epic ID"},
		{name: "epic blank title", content: "# Title\n\nEpic: #1\n\nBody", want: "malformed Epic"},
		{name: "extra separator", content: "# Title\n\nTypes: feature\n\n\nBody", want: "exactly one blank line"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := CanonicalizeDocument([]byte(test.content), &fakeDocumentOptions{types: []TypeOption{{Slug: "feature"}}})
			require.Error(t, err)
			var validation ValidationError
			assert.True(t, errors.As(err, &validation))
			assert.Contains(t, err.Error(), test.want)
		})
	}

	_, err := CanonicalizeDocument([]byte("# Title\n\nTypes: feature\n\nBody"), &fakeDocumentOptions{typeErr: errors.New("backend unavailable")})
	require.Error(t, err)
	var validation ValidationError
	assert.False(t, errors.As(err, &validation))
}
