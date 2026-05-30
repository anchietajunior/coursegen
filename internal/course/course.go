// Package course locates the course's input artifacts on disk: lesson specs,
// module specs and shared docs. Read-only with respect to docs/.
package course

import (
	"fmt"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/coursegen/coursegen/internal/config"
)

var lessonRe = regexp.MustCompile(`lesson-(\d{2})-(\d{2})`)

// Lesson is a single unit of work.
type Lesson struct {
	ModuleID string
	LessonID string
	SpecPath string
}

// Unit returns e.g. "lesson-01-01".
func (l Lesson) Unit() string { return fmt.Sprintf("lesson-%s-%s", l.ModuleID, l.LessonID) }

// ModuleDir returns e.g. "module-01".
func (l Lesson) ModuleDir() string { return "module-" + l.ModuleID }

// Course wraps a config to find artifacts.
type Course struct{ cfg *config.Config }

func New(cfg *config.Config) *Course { return &Course{cfg: cfg} }

// Lessons discovers all lessons, sorted, optionally filtered by module/lesson.
func (c *Course) Lessons(moduleFilter, lessonFilter string) ([]Lesson, error) {
	pattern := filepath.Join(c.cfg.DocsPath(), "05-lesson-specs", "module-*", "lesson-*-*.md")
	matches, err := filepath.Glob(pattern)
	if err != nil {
		return nil, err
	}

	var lessons []Lesson
	for _, p := range matches {
		m := lessonRe.FindStringSubmatch(filepath.Base(p))
		if m == nil {
			continue
		}
		lessons = append(lessons, Lesson{ModuleID: m[1], LessonID: m[2], SpecPath: p})
	}

	sort.Slice(lessons, func(i, j int) bool {
		if lessons[i].ModuleID != lessons[j].ModuleID {
			return lessons[i].ModuleID < lessons[j].ModuleID
		}
		return lessons[i].LessonID < lessons[j].LessonID
	})

	if moduleFilter != "" {
		id := normalizeModule(moduleFilter)
		lessons = filter(lessons, func(l Lesson) bool { return l.ModuleID == id })
	}
	if lessonFilter != "" {
		unit := normalizeLesson(lessonFilter)
		lessons = filter(lessons, func(l Lesson) bool { return l.Unit() == unit })
	}
	return lessons, nil
}

// ModuleSpecPath returns the path to a module's spec.
func (c *Course) ModuleSpecPath(moduleID string) string {
	return filepath.Join(c.cfg.DocsPath(), "04-module-specs", "module-"+moduleID+".md")
}

// LessonSpecPath reconstructs a lesson spec path from ids (used by retry).
func (c *Course) LessonSpecPath(moduleID, lessonID string) string {
	return filepath.Join(c.cfg.DocsPath(), "05-lesson-specs", "module-"+moduleID,
		fmt.Sprintf("lesson-%s-%s.md", moduleID, lessonID))
}

func normalizeModule(s string) string {
	s = strings.TrimPrefix(s, "module-")
	if len(s) == 1 {
		s = "0" + s
	}
	return s
}

func normalizeLesson(s string) string {
	if strings.HasPrefix(s, "lesson-") {
		return s
	}
	return "lesson-" + s
}

func filter(in []Lesson, keep func(Lesson) bool) []Lesson {
	out := in[:0]
	for _, l := range in {
		if keep(l) {
			out = append(out, l)
		}
	}
	return out
}
