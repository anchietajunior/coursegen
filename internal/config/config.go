// Package config loads and resolves coursegen.json. All paths are resolved
// relative to the project root (the directory containing coursegen.json).
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Config is the parsed coursegen.json plus the resolved project root.
type Config struct {
	Root    string `json:"-"`
	Version int    `json:"version"`

	Course struct {
		Name     string `json:"name"`
		Language string `json:"language"`
	} `json:"course"`

	Paths struct {
		Docs   string `json:"docs"`
		Output string `json:"output"`
		State  string `json:"state"`
		Logs   string `json:"logs"`
		Runs   string `json:"runs"`
	} `json:"paths"`

	Readiness struct {
		Required       *bool  `json:"required"`
		Source         string `json:"source"`
		ApprovedMarker string `json:"approved_marker"`
	} `json:"readiness"`

	Runners struct {
		Default string `json:"default"`
	} `json:"runners"`

	Execution struct {
		TimeoutSeconds      int    `json:"timeout_seconds"`
		OnValidationFailure string `json:"on_validation_failure"`
	} `json:"execution"`

	Context struct {
		Shared            []string `json:"shared"`
		MaxTokensEstimate int      `json:"max_tokens_estimate"`
	} `json:"context"`
}

// Load reads coursegen.json from root and applies defaults.
func Load(root string) (*Config, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(abs, "coursegen.json")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("coursegen.json não encontrado em %s. Rode `coursegen init` primeiro", abs)
	} else if err != nil {
		return nil, err
	}

	cfg := &Config{Root: abs}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("coursegen.json inválido: %w", err)
	}
	cfg.applyDefaults()
	return cfg, nil
}

func (c *Config) applyDefaults() {
	setIfEmpty(&c.Course.Language, "pt-BR")
	setIfEmpty(&c.Paths.Docs, "docs")
	setIfEmpty(&c.Paths.Output, "output")
	setIfEmpty(&c.Paths.State, ".coursegen/state.json")
	setIfEmpty(&c.Paths.Logs, ".coursegen/logs")
	setIfEmpty(&c.Paths.Runs, ".coursegen/runs")
	setIfEmpty(&c.Readiness.Source, "docs/06-course-readiness-checklist.md")
	setIfEmpty(&c.Readiness.ApprovedMarker, "APROVADO")
	setIfEmpty(&c.Runners.Default, "claude")
	setIfEmpty(&c.Execution.OnValidationFailure, "warn")
	if c.Execution.TimeoutSeconds == 0 {
		c.Execution.TimeoutSeconds = 900
	}
	if c.MaxTokensEstimate() == 0 {
		c.Context.MaxTokensEstimate = 120_000
	}
	if len(c.Context.Shared) == 0 {
		c.Context.Shared = []string{
			"docs/01-course-prd.md",
			"docs/02-market-research.md",
			"docs/03-learning-architecture.md",
		}
	}
}

// Abs resolves a project-relative path to an absolute one.
func (c *Config) Abs(rel string) string { return filepath.Join(c.Root, rel) }

func (c *Config) DocsPath() string        { return c.Abs(c.Paths.Docs) }
func (c *Config) OutputPath() string      { return c.Abs(c.Paths.Output) }
func (c *Config) StatePath() string       { return c.Abs(c.Paths.State) }
func (c *Config) RunsPath() string        { return c.Abs(c.Paths.Runs) }
func (c *Config) ReadinessSource() string { return c.Abs(c.Readiness.Source) }

func (c *Config) ReadinessRequired() bool {
	return c.Readiness.Required == nil || *c.Readiness.Required
}

func (c *Config) MaxTokensEstimate() int { return c.Context.MaxTokensEstimate }

// PromptTemplatePath returns the project override if present, else "" (caller
// falls back to the embedded default).
func (c *Config) PromptTemplatePath() string {
	override := c.Abs("coursegen/prompts/generate-lesson.tmpl")
	if _, err := os.Stat(override); err == nil {
		return override
	}
	return ""
}

func setIfEmpty(s *string, def string) {
	if *s == "" {
		*s = def
	}
}
