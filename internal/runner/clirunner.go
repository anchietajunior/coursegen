package runner

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Spec is the YAML-driven configuration of a CLI runner. The exact flags of
// each agent are version-dependent and live here ON PURPOSE, so a tool update
// is a config edit, not a code change.
type Spec struct {
	Name        string   `yaml:"name"`
	Type        string   `yaml:"type"`
	Bin         string   `yaml:"bin"`
	Healthcheck string   `yaml:"healthcheck"`
	Args        []string `yaml:"args"`
	KillSignal  string   `yaml:"kill_signal"`

	Prompt struct {
		Via  string `yaml:"via"`  // stdin | arg | file
		Flag string `yaml:"flag"` // optional flag preceding the prompt
	} `yaml:"prompt"`

	Output struct {
		StripCodeFences bool `yaml:"strip_code_fences"`
	} `yaml:"output"`

	Env map[string]string `yaml:"env"`
}

var envVarRe = regexp.MustCompile(`\$\{(\w+)\}`)

// CliRunner shells out to an external agent CLI.
type CliRunner struct{ spec Spec }

func NewCliRunner(spec Spec) *CliRunner { return &CliRunner{spec: spec} }

func (r *CliRunner) Name() string { return r.spec.Name }

func (r *CliRunner) Available() bool {
	if r.spec.Healthcheck == "" {
		_, err := exec.LookPath(r.spec.Bin)
		return err == nil
	}
	parts := strings.Fields(r.spec.Healthcheck)
	cmd := exec.Command(parts[0], parts[1:]...)
	return cmd.Run() == nil
}

func (r *CliRunner) Version() string {
	out, err := exec.Command(r.spec.Bin, "--version").CombinedOutput()
	if err != nil {
		return ""
	}
	line := strings.SplitN(strings.TrimSpace(string(out)), "\n", 2)[0]
	return line
}

func (r *CliRunner) Run(inv Invocation) Result {
	argv, stdin := r.buildCommand(inv)

	ctx, cancel := context.WithTimeout(context.Background(), inv.Timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	cmd.Dir = inv.Workdir
	cmd.Env = r.buildEnv(inv)
	if stdin != nil {
		cmd.Stdin = strings.NewReader(*stdin)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	start := time.Now()
	err := cmd.Run()
	dur := time.Since(start)

	if ctx.Err() == context.DeadlineExceeded {
		return Result{Status: StatusTimeout, Stdout: stdout.String(), Stderr: stderr.String(),
			Duration: dur, Err: "timeout após " + inv.Timeout.String()}
	}

	var execErr *exec.Error
	if errors.As(err, &execErr) {
		return Result{Status: StatusFailed, Duration: dur,
			Err: "binário não encontrado para o runner '" + r.spec.Name + "': " + r.spec.Bin + ". Rode `coursegen doctor`."}
	}

	artifact := r.postProcess(stdout.String())
	exitCode := cmd.ProcessState.ExitCode()
	ok := err == nil && strings.TrimSpace(artifact) != ""

	res := Result{
		Artifact: artifact, Stdout: stdout.String(), Stderr: stderr.String(),
		ExitCode: exitCode, Duration: dur,
	}
	if ok {
		res.Status = StatusOK
	} else {
		res.Status = StatusFailed
		res.Err = failureReason(err, stderr.String(), exitCode)
	}
	return res
}

func (r *CliRunner) buildCommand(inv Invocation) (argv []string, stdin *string) {
	via := r.spec.Prompt.Via
	if via == "" {
		via = "stdin"
	}
	args := append([]string{}, r.spec.Args...)
	promptFile := filepath.Join(inv.Workdir, "PROMPT.md")

	switch via {
	case "stdin":
		stdin = &inv.Prompt
	case "arg":
		if r.spec.Prompt.Flag != "" {
			args = append(args, r.spec.Prompt.Flag)
		}
		args = append(args, "{prompt}")
	case "file":
		if r.spec.Prompt.Flag != "" {
			args = append(args, r.spec.Prompt.Flag)
		}
		args = append(args, "{prompt_file}")
	}

	argv = append([]string{r.spec.Bin}, expandTokens(args, inv, promptFile)...)
	return argv, stdin
}

func expandTokens(args []string, inv Invocation, promptFile string) []string {
	repl := strings.NewReplacer(
		"{prompt}", inv.Prompt,
		"{prompt_file}", promptFile,
		"{workdir}", inv.Workdir,
		"{output_path}", inv.OutputPath,
	)
	out := make([]string, len(args))
	for i, a := range args {
		out[i] = repl.Replace(a)
	}
	return out
}

func (r *CliRunner) buildEnv(inv Invocation) []string {
	env := os.Environ()
	for k, v := range r.spec.Env {
		env = append(env, k+"="+expandEnvValue(v))
	}
	for k, v := range inv.Env {
		env = append(env, k+"="+v)
	}
	return env
}

func expandEnvValue(v string) string {
	return envVarRe.ReplaceAllStringFunc(v, func(m string) string {
		name := envVarRe.FindStringSubmatch(m)[1]
		return os.Getenv(name)
	})
}

func (r *CliRunner) postProcess(stdout string) string {
	text := strings.TrimSpace(stdout)
	if r.spec.Output.StripCodeFences {
		text = stripCodeFences(text)
	}
	return text
}

func stripCodeFences(text string) string {
	if strings.HasPrefix(text, "```") && strings.HasSuffix(text, "```") {
		if i := strings.IndexByte(text, '\n'); i >= 0 {
			text = text[i+1:]
		}
		text = strings.TrimSuffix(text, "```")
		text = strings.TrimRight(text, "\n")
	}
	return text
}

func failureReason(err error, stderr string, exitCode int) string {
	msg := strings.TrimSpace(stderr)
	if msg != "" {
		if len(msg) > 500 {
			msg = msg[:500]
		}
		return msg
	}
	if err != nil {
		return "exit code " + strconv.Itoa(exitCode)
	}
	return "saída vazia"
}
