// Package prompt renders the lesson prompt from a Go text/template and a
// context pack. The default template is embedded in the binary; a project may
// override it at coursegen/prompts/generate-lesson.tmpl.
package prompt

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"os"
	"strings"
	"text/template"

	"github.com/anchietajunior/coursegen/internal/contextpack"
)

//go:embed templates/generate-lesson.tmpl
var defaultTemplate string

// Builder renders prompts from a fixed template source.
type Builder struct {
	source string
	tmpl   *template.Template
}

// NewBuilder loads the template from path, or the embedded default when path
// is empty.
func NewBuilder(path string) (*Builder, error) {
	source := defaultTemplate
	if path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		source = string(data)
	}
	t, err := template.New("generate-lesson").Parse(source)
	if err != nil {
		return nil, err
	}
	return &Builder{source: source, tmpl: t}, nil
}

// Render produces the final prompt for a pack.
func (b *Builder) Render(pack *contextpack.Pack) (string, error) {
	var sb strings.Builder
	if err := b.tmpl.Execute(&sb, pack); err != nil {
		return "", err
	}
	return sb.String(), nil
}

// Fingerprint of the template source; editing it invalidates the cache.
func (b *Builder) Fingerprint() string {
	h := sha256.Sum256([]byte(b.source))
	return hex.EncodeToString(h[:])
}
