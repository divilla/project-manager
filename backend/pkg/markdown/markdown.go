package markdown

import (
	"bytes"

	"github.com/microcosm-cc/bluemonday"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
)

type Parser interface {
	Parse(source string) string
}

type Sanitizer interface {
	Parse(source string) string
}

type GoldmarkParser struct {
	parser goldmark.Markdown
}

func NewGoldmarkParser() *GoldmarkParser {
	return &GoldmarkParser{
		parser: goldmark.New(
			goldmark.WithExtensions(extension.GFM),
			goldmark.WithRendererOptions(html.WithUnsafe()),
		),
	}
}

func (p *GoldmarkParser) Parse(source string) string {
	var out bytes.Buffer
	if err := p.parser.Convert([]byte(source), &out); err != nil {
		return ""
	}
	return out.String()
}

type BluemondaySanitizer struct {
	policy *bluemonday.Policy
}

func NewBluemondaySanitizer() *BluemondaySanitizer {
	policy := bluemonday.UGCPolicy()
	policy.AllowAttrs("type", "checked", "disabled").OnElements("input")
	return &BluemondaySanitizer{policy: policy}
}

func (s *BluemondaySanitizer) Parse(source string) string {
	return s.policy.Sanitize(source)
}
