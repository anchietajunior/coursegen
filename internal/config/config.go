// Package config loads and resolves coursegen.yml. All paths are resolved
// relative to the project root (the directory containing coursegen.yml).
package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config is the parsed coursegen.yml plus the resolved project root.
type Config struct {
	Root    string `yaml:"-"`
	Version int    `yaml:"version"`

	Course struct {
		Name     string `yaml:"name"`
		Language string `yaml:"language"`
	} `yaml:"course"`

	Paths struct {
		Docs   string `yaml:"docs"`
		Output string `yaml:"output"`
		State  string `yaml:"state"`
		Logs   string `yaml:"logs"`
		Runs   string `yaml:"runs"`
	} `yaml:"paths"`

	Readiness struct {
		Required       *bool  `yaml:"required"`
		Source         string `yaml:"source"`
		ApprovedMarker string `yaml:"approved_marker"`
	} `yaml:"readiness"`

	Runners struct {
		Default string `yaml:"default"`
	} `yaml:"runners"`

	Execution struct {
		TimeoutSeconds      int    `yaml:"timeout_seconds"`
		OnValidationFailure string `yaml:"on_validation_failure"`
	} `yaml:"execution"`

	Context struct {
		Shared            []string `yaml:"shared"`
		MaxTokensEstimate int      `yaml:"max_tokens_estimate"`
	} `yaml:"context"`
}

// Load reads coursegen.yml from root and applies defaults.
func Load(root string) (*Config, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	path := filepath.Join(abs, "coursegen.yml")
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("coursegen.yml não encontrado em %s. Rode `coursegen init` primeiro", abs)
	} else if err != nil {
		return nil, err
	}

	cfg := &Config{Root: abs}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("coursegen.yml inválido: %w", err)
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
