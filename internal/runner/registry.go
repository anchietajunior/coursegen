package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// builtinSpecs are the default runner configs. Resolution order:
//  1. project override: coursegen/runners/<name>.json
//  2. these built-ins
func builtinSpec(name string) (Spec, bool) {
	specs := map[string]Spec{
		"claude": {
			Name: "claude", Bin: "claude", Healthcheck: "claude --version",
			Args: []string{"-p", "--output-format", "text"}, KillSignal: "TERM",
		},
		"codex": {
			Name: "codex", Bin: "codex", Healthcheck: "codex --version",
			Args: []string{"exec"}, KillSignal: "TERM",
		},
		"gemini": {
			Name: "gemini", Bin: "gemini", Healthcheck: "gemini --version",
			KillSignal: "TERM",
		},
		"cursor": {
			Name: "cursor", Bin: "cursor-agent", Healthcheck: "cursor-agent --version",
			KillSignal: "TERM",
		},
		"opencode": {
			Name: "opencode", Bin: "opencode", Healthcheck: "opencode --version",
			Args: []string{"run"}, KillSignal: "TERM",
		},
		"mock": {Name: "mock", Type: "mock"},
	}
	// Per-runner prompt strategy.
	if s, ok := specs[name]; ok {
		switch name {
		case "claude", "opencode", "codex":
			s.Prompt.Via = ifEmpty(s.Prompt.Via, viaFor(name))
		case "gemini", "cursor":
			s.Prompt.Via = "arg"
			s.Prompt.Flag = "-p"
		}
		return s, true
	}
	return Spec{}, false
}

func viaFor(name string) string {
	switch name {
	case "codex", "opencode":
		return "arg"
	default: // claude
		return "stdin"
	}
}

// Names returns the known built-in runner names.
func Names() []string {
	return []string{"claude", "codex", "gemini", "cursor", "opencode", "mock"}
}

// Resolve returns a Runner for name, honoring a project override if present.
func Resolve(name, root string) (Runner, error) {
	spec, err := resolveSpec(name, root)
	if err != nil {
		return nil, err
	}
	if spec.Type == "mock" || name == "mock" {
		return NewMockRunner(name), nil
	}
	return NewCliRunner(spec), nil
}

func resolveSpec(name, root string) (Spec, error) {
	overridePath := filepath.Join(root, "coursegen", "runners", name+".json")
	if data, err := os.ReadFile(overridePath); err == nil {
		var spec Spec
		if err := json.Unmarshal(data, &spec); err != nil {
			return Spec{}, fmt.Errorf("runner override inválido (%s): %w", overridePath, err)
		}
		if spec.Name == "" {
			spec.Name = name
		}
		return spec, nil
	}
	if spec, ok := builtinSpec(name); ok {
		return spec, nil
	}
	return Spec{}, fmt.Errorf("runner desconhecido: '%s'. Disponíveis: %v", name, Names())
}

func ifEmpty(s, def string) string {
	if s == "" {
		return def
	}
	return s
}
