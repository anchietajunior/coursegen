package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/anchietajunior/coursegen/internal/assets"
	"github.com/anchietajunior/coursegen/internal/config"
)

// knownAgents is the ordered list of agents the user can target. These are the
// same tools that act as runners later.
var knownAgents = []string{"claude", "codex", "gemini", "cursor", "opencode"}

// agentSkillDir maps an agent to its skills directory (relative to HOME) for
// agents we can link into with confidence. Agents NOT listed are still served
// by the canonical store (~/.agents/skills); we just print guidance for them.
var agentSkillDir = map[string]string{
	"claude": ".claude/skills",
	"cursor": ".cursor/skills-cursor",
}

// setupState is persisted at ~/.config/coursegen/state.json.
type setupState struct {
	Agent           string   `json:"agent"`
	InstalledSkills []string `json:"installed_skills"`
	Scope           string   `json:"scope"`
	InstalledAt     string   `json:"installed_at"`
}

// cmdSetup installs the CourseGen planning skills into the chosen agent.
//
// Mirrors `compozy setup`: a canonical agnostic store at ~/.agents/skills/<name>
// plus a symlink (or copy) into the agent's own skills dir, recording the choice.
func cmdSetup(args []string) int {
	var (
		agentFlag  string
		copyMode   bool
		listOnly   bool
		yes        bool
		projectDir bool
	)
	rest := args
	// tiny hand-parse so we can read flags in any order
	var positional []string
	for i := 0; i < len(rest); i++ {
		switch rest[i] {
		case "--agent":
			if i+1 < len(rest) {
				agentFlag = rest[i+1]
				i++
			}
		case "--copy":
			copyMode = true
		case "--list":
			listOnly = true
		case "--yes", "-y":
			yes = true
		case "--project":
			projectDir = true
		default:
			if strings.HasPrefix(rest[i], "--agent=") {
				agentFlag = strings.TrimPrefix(rest[i], "--agent=")
			} else {
				positional = append(positional, rest[i])
			}
		}
	}
	_ = positional

	skills, err := embeddedSkillNames()
	if err != nil {
		return fail(err)
	}

	if listOnly {
		fmt.Println("Skills de planejamento embarcadas:")
		for _, s := range skills {
			fmt.Printf("  • %s\n", s)
		}
		return 0
	}

	agent, err := resolveAgent(agentFlag, yes)
	if err != nil {
		return fail(err)
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return fail(err)
	}

	// Base for the canonical agnostic store.
	base := home
	scope := "global"
	if projectDir {
		base, _ = os.Getwd()
		scope = "project"
	}
	canonical := filepath.Join(base, ".agents", "skills")

	fmt.Printf("Instalando %d skills para o agente '%s' (%s)…\n", len(skills), agent, scope)

	for _, name := range skills {
		if err := writeSkill(canonical, name); err != nil {
			return fail(fmt.Errorf("falha ao escrever skill %s: %w", name, err))
		}
	}
	fmt.Printf("✓ store canônico: %s\n", tilde(canonical, home))

	// Link/copy into the agent's own skills dir, when known.
	if rel, ok := agentSkillDir[agent]; ok {
		agentDir := filepath.Join(home, rel)
		if err := os.MkdirAll(agentDir, 0o755); err != nil {
			return fail(err)
		}
		for _, name := range skills {
			src := filepath.Join(canonical, name)
			dst := filepath.Join(agentDir, name)
			if err := linkOrCopy(src, dst, name, copyMode); err != nil {
				return fail(err)
			}
		}
		verb := "symlinkadas"
		if copyMode {
			verb = "copiadas"
		}
		fmt.Printf("✓ %s em %s\n", verb, tilde(agentDir, home))
	} else {
		fmt.Printf("ℹ %s não tem diretório de skills conhecido; use o store agnóstico acima\n", agent)
		fmt.Printf("  (aponte o %s para %s)\n", agent, tilde(canonical, home))
	}

	if err := saveSetupState(agent, skills, scope); err != nil {
		// non-fatal
		fmt.Printf("⚠ não foi possível gravar o estado: %v\n", err)
	}

	fmt.Printf("\n✓ Pronto. Agente padrão: %s.\n", agent)
	fmt.Println("  Comece o planejamento no seu agente com /course-discovery,")
	fmt.Println("  e depois produza com: coursegen generate lessons")
	return 0
}

// resolveAgent picks the agent from the flag, an interactive prompt, or a
// project default — in that order.
func resolveAgent(flagVal string, yes bool) (string, error) {
	if flagVal != "" {
		if !isKnownAgent(flagVal) {
			return "", fmt.Errorf("agente desconhecido: '%s'. Opções: %s", flagVal, strings.Join(knownAgents, ", "))
		}
		return flagVal, nil
	}

	// Project default, if any.
	def := ""
	if cfg, err := config.Load(mustGetwd()); err == nil {
		def = cfg.Runners.Default
	}

	if yes || !stdinIsTerminal() {
		if def != "" && isKnownAgent(def) {
			return def, nil
		}
		return "claude", nil
	}

	// Interactive menu.
	fmt.Println("Qual agente você usa?")
	for i, a := range knownAgents {
		marker := " "
		if a == def {
			marker = "*"
		}
		fmt.Printf("  %s[%d] %s\n", marker, i+1, a)
	}
	prompt := "Escolha [1]: "
	if def != "" {
		prompt = fmt.Sprintf("Escolha (padrão: %s): ", def)
	}
	fmt.Print(prompt)

	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if def != "" {
			return def, nil
		}
		return "claude", nil
	}
	choice := strings.TrimSpace(scanner.Text())
	switch {
	case choice == "" && def != "":
		return def, nil
	case choice == "":
		return knownAgents[0], nil
	case isKnownAgent(choice):
		return choice, nil
	default:
		if n, err := strconv.Atoi(choice); err == nil && n >= 1 && n <= len(knownAgents) {
			return knownAgents[n-1], nil
		}
		return "", fmt.Errorf("escolha inválida: %q", choice)
	}
}

// --- file ops ---------------------------------------------------------------

func embeddedSkillNames() ([]string, error) {
	entries, err := fs.ReadDir(assets.SkillsFS(), "skills")
	if err != nil {
		return nil, err
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)
	return names, nil
}

// writeSkill materializes the embedded skills/<name> tree under dstBase/<name>.
func writeSkill(dstBase, name string) error {
	root := "skills/" + name
	return fs.WalkDir(assets.SkillsFS(), root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath := strings.TrimPrefix(p, "skills/")
		target := filepath.Join(dstBase, relPath)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := fs.ReadFile(assets.SkillsFS(), p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

// linkOrCopy points dst at src (symlink) or copies the skill tree (copyMode).
func linkOrCopy(src, dst, name string, copyMode bool) error {
	_ = os.RemoveAll(dst)
	if copyMode {
		return copyTree(src, dst)
	}
	if err := os.Symlink(src, dst); err != nil {
		// Fall back to copy if symlinks are unavailable (e.g. some Windows).
		return copyTree(src, dst)
	}
	return nil
}

func copyTree(src, dst string) error {
	return filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		target := filepath.Join(dst, rel)
		if info.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
}

func saveSetupState(agent string, skills []string, scope string) error {
	home, err := os.UserHomeDir()
	if err != nil {
		return err
	}
	dir := filepath.Join(home, ".config", "coursegen")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	st := setupState{
		Agent: agent, InstalledSkills: skills, Scope: scope,
		InstalledAt: time.Now().Format(time.RFC3339),
	}
	data, err := json.MarshalIndent(st, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	return os.WriteFile(filepath.Join(dir, "state.json"), data, 0o644)
}

// --- helpers ----------------------------------------------------------------

func isKnownAgent(a string) bool {
	for _, k := range knownAgents {
		if k == a {
			return true
		}
	}
	return false
}

func stdinIsTerminal() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

func mustGetwd() string {
	wd, _ := os.Getwd()
	return wd
}

func tilde(path, home string) string {
	if strings.HasPrefix(path, home) {
		return "~" + strings.TrimPrefix(path, home)
	}
	return path
}
