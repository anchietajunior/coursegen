// Package contextpack builds the MINIMAL set of context handed to the agent for
// ONE lesson. This is the heart of token economy: shared docs (PRD, market
// research, architecture) + the single module spec + the single lesson spec.
// Never all lessons.
package contextpack

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/coursegen/coursegen/internal/config"
	"github.com/coursegen/coursegen/internal/course"
	"github.com/coursegen/coursegen/internal/tokens"
)

// SharedDoc is a labeled shared document included in the pack.
type SharedDoc struct {
	Title   string
	Content string
}

// Pack holds everything the prompt template needs (exported fields/methods so
// text/template can access them).
type Pack struct {
	Language   string
	ModuleID   string
	LessonID   string
	Shared     []SharedDoc
	ModuleSpec string
	LessonSpec string
}

var sharedTitles = map[string]string{
	"01-course-prd.md":            "COURSE PRD",
	"02-market-research.md":       "MARKET RESEARCH",
	"03-learning-architecture.md": "LEARNING ARCHITECTURE",
}

// Build assembles the minimal pack for one lesson, failing clearly if any
// required input is missing.
func Build(cfg *config.Config, c *course.Course, lesson course.Lesson) (*Pack, error) {
	var shared []SharedDoc
	for _, rel := range cfg.Context.Shared {
		content, err := readRequired(cfg.Abs(rel), "doc compartilhado")
		if err != nil {
			return nil, err
		}
		shared = append(shared, SharedDoc{Title: titleFor(rel), Content: content})
	}

	moduleSpec, err := readRequired(c.ModuleSpecPath(lesson.ModuleID),
		"module spec do módulo "+lesson.ModuleID)
	if err != nil {
		return nil, err
	}
	lessonSpec, err := readRequired(lesson.SpecPath, "lesson spec "+lesson.Unit())
	if err != nil {
		return nil, err
	}

	return &Pack{
		Language:   cfg.Course.Language,
		ModuleID:   lesson.ModuleID,
		LessonID:   lesson.LessonID,
		Shared:     shared,
		ModuleSpec: moduleSpec,
		LessonSpec: lessonSpec,
	}, nil
}

// AllContent concatenates everything, for token estimates and fingerprints.
func (p *Pack) AllContent() string {
	var b strings.Builder
	for _, s := range p.Shared {
		b.WriteString(s.Content)
	}
	b.WriteString(p.ModuleSpec)
	b.WriteString(p.LessonSpec)
	return b.String()
}

// InputTokens estimates the context-pack input size.
func (p *Pack) InputTokens() int { return tokens.Estimate(p.AllContent()) }

// Fingerprint is a stable hash of the inputs; it drives idempotent skipping.
func (p *Pack) Fingerprint(templateFingerprint string) string {
	h := sha256.Sum256([]byte(strings.Join(
		[]string{templateFingerprint, p.Language, p.ModuleID, p.LessonID, p.AllContent()}, " ")))
	return hex.EncodeToString(h[:])
}

func titleFor(rel string) string {
	base := filepath.Base(rel)
	if t, ok := sharedTitles[base]; ok {
		return t
	}
	return strings.ToUpper(strings.TrimSuffix(base, ".md"))
}

func readRequired(path, label string) (string, error) {
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return "", fmt.Errorf("%s não encontrado: %s", label, path)
	} else if err != nil {
		return "", err
	}
	return string(data), nil
}
