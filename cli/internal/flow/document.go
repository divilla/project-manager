package flow

import (
	"fmt"
	"strconv"
	"strings"
	"unicode"
)

// TypeOption is one ordered backend Change type option.
type TypeOption struct {
	Slug string
}

// EpicOption is one current-project Epic option.
type EpicOption struct {
	ID    int
	Title string
}

// DocumentOptions supplies current API-backed metadata without exposing transport DTOs.
type DocumentOptions interface {
	ChangeTypes() ([]TypeOption, error)
	Epics() ([]EpicOption, error)
}

// ValidationError identifies artifact content that an Editor may reopen in place.
type ValidationError struct {
	Message string
}

func (e ValidationError) Error() string { return e.Message }

// CanonicalDocument is validated artifact-local content ready for comparison or persistence.
type CanonicalDocument struct {
	Bytes []byte
	Title string
}

// CanonicalizeDocument normalizes and validates an Idea, Spec, or PR submission.
func CanonicalizeDocument(content []byte, options DocumentOptions) (CanonicalDocument, error) {
	normalized := strings.ReplaceAll(string(content), "\r\n", "\n")
	lines := strings.Split(normalized, "\n")
	if len(lines) < 2 || !strings.HasPrefix(lines[0], "# ") || strings.HasPrefix(lines[0], "## ") || strings.TrimSpace(strings.TrimPrefix(lines[0], "# ")) == "" {
		return CanonicalDocument{}, ValidationError{Message: "# Title parsing failed"}
	}
	if lines[1] != "" {
		return CanonicalDocument{}, ValidationError{Message: "title must be followed by exactly one blank line"}
	}

	title := strings.TrimPrefix(lines[0], "# ")
	index := 2
	var typeValues []string
	var epicID *int
	seenTypes := false
	seenEpic := false
	for index < len(lines) {
		line := lines[index]
		if line == "" {
			if index == len(lines)-1 {
				break
			}
			return CanonicalDocument{}, ValidationError{Message: "metadata fields must be followed by exactly one blank line"}
		}
		switch {
		case strings.HasPrefix(line, "Types:"):
			if seenTypes {
				return CanonicalDocument{}, ValidationError{Message: "duplicate Types field"}
			}
			values, err := parseTypes(line)
			if err != nil {
				return CanonicalDocument{}, err
			}
			seenTypes = true
			typeValues = values
		case strings.HasPrefix(line, "Epic:"):
			if seenEpic {
				return CanonicalDocument{}, ValidationError{Message: "duplicate Epic field"}
			}
			id, err := parseEpic(line)
			if err != nil {
				return CanonicalDocument{}, err
			}
			seenEpic = true
			epicID = &id
		default:
			goto body
		}
		if index+1 >= len(lines) || lines[index+1] != "" {
			return CanonicalDocument{}, ValidationError{Message: "metadata fields must be followed by exactly one blank line"}
		}
		index += 2
	}

body:
	body := ""
	if index < len(lines) {
		body = strings.Join(lines[index:], "\n")
	}
	if strings.TrimSpace(body) == "" {
		return CanonicalDocument{}, ValidationError{Message: "document body must contain at least one non-whitespace character"}
	}
	if options == nil && (seenTypes || seenEpic) {
		return CanonicalDocument{}, fmt.Errorf("document metadata options boundary is required")
	}
	canonicalTypes := []string(nil)
	if seenTypes {
		available, err := options.ChangeTypes()
		if err != nil {
			return CanonicalDocument{}, fmt.Errorf("load change type options: %w", err)
		}
		selected := make(map[string]struct{}, len(typeValues))
		for _, value := range typeValues {
			selected[value] = struct{}{}
		}
		known := make(map[string]struct{}, len(available))
		emitted := make(map[string]struct{}, len(available))
		for _, option := range available {
			slug := strings.TrimSpace(option.Slug)
			if slug == "" {
				continue
			}
			known[slug] = struct{}{}
			if _, ok := selected[slug]; ok {
				if _, duplicate := emitted[slug]; duplicate {
					continue
				}
				canonicalTypes = append(canonicalTypes, slug)
				emitted[slug] = struct{}{}
			}
		}
		for _, value := range typeValues {
			if _, ok := known[value]; !ok {
				return CanonicalDocument{}, ValidationError{Message: fmt.Sprintf("unknown type slug %q", value)}
			}
		}
	}

	canonicalEpic := ""
	if seenEpic {
		available, err := options.Epics()
		if err != nil {
			return CanonicalDocument{}, fmt.Errorf("load Epic options: %w", err)
		}
		found := false
		for _, option := range available {
			if option.ID != *epicID {
				continue
			}
			title := strings.TrimSpace(option.Title)
			if title == "" || strings.ContainsAny(title, "\r\n") {
				return CanonicalDocument{}, fmt.Errorf("epic %d has an invalid canonical title", option.ID)
			}
			canonicalEpic = fmt.Sprintf("Epic: %s #%d", title, option.ID)
			found = true
			break
		}
		if !found {
			return CanonicalDocument{}, ValidationError{Message: fmt.Sprintf("unknown Epic ID %d", *epicID)}
		}
	}

	var output strings.Builder
	output.WriteString(lines[0])
	output.WriteString("\n\n")
	if seenTypes {
		output.WriteString("Types: ")
		output.WriteString(strings.Join(canonicalTypes, "|"))
		output.WriteString("\n\n")
	}
	if seenEpic {
		output.WriteString(canonicalEpic)
		output.WriteString("\n\n")
	}
	output.WriteString(body)
	return CanonicalDocument{Bytes: []byte(output.String()), Title: title}, nil
}

func parseTypes(line string) ([]string, error) {
	raw := strings.TrimPrefix(line, "Types:")
	parts := strings.Split(raw, "|")
	seen := make(map[string]struct{}, len(parts))
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		value := strings.TrimSpace(part)
		if value == "" {
			return nil, ValidationError{Message: "Types contains an empty value"}
		}
		if _, duplicate := seen[value]; duplicate {
			return nil, ValidationError{Message: fmt.Sprintf("duplicate type slug %q", value)}
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	return values, nil
}

func parseEpic(line string) (int, error) {
	value := strings.TrimSpace(strings.TrimPrefix(line, "Epic:"))
	hash := strings.LastIndex(value, "#")
	if hash <= 0 || !unicode.IsSpace(rune(value[hash-1])) {
		return 0, ValidationError{Message: fmt.Sprintf("malformed Epic value %q", value)}
	}
	title := strings.TrimSpace(value[:hash])
	idText := strings.TrimSpace(value[hash+1:])
	if title == "" || idText == "" || strings.ContainsAny(idText, " \t") {
		return 0, ValidationError{Message: fmt.Sprintf("malformed Epic value %q", value)}
	}
	id64, err := strconv.ParseUint(idText, 10, 63)
	if err != nil || id64 == 0 {
		return 0, ValidationError{Message: fmt.Sprintf("invalid Epic ID %q", idText)}
	}
	return int(id64), nil
}
