package course

import (
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/anchietajunior/coursegen/internal/config"
)

// ReadinessResult is the outcome of parsing the readiness checklist.
type ReadinessResult struct {
	Exists   bool
	Approved bool
	Verdict  string
	Blockers *int
	Warnings *int
	Source   string
}

var (
	verdictSectionRe = regexp.MustCompile(`(?ims)^#+\s*Veredito\b(.*?)(?:^#+\s|\z)`)
	blockersRe       = regexp.MustCompile(`(?i)Bloqueadores:\s*(\d+)`)
	warningsRe       = regexp.MustCompile(`(?i)Avisos:\s*(\d+)`)
)

// CheckReadiness parses docs/06-course-readiness-checklist.md and decides
// whether the course is approved for production. This is the gate that
// materializes "o agente define, a CLI escala".
func CheckReadiness(cfg *config.Config) ReadinessResult {
	source := cfg.ReadinessSource()
	data, err := os.ReadFile(source)
	if err != nil {
		return ReadinessResult{Exists: false, Approved: false, Verdict: "AUSENTE", Source: source}
	}

	text := string(data)
	region := verdictRegion(text)
	marker := cfg.Readiness.ApprovedMarker

	hasOK := regexp.MustCompile(`\b` + regexp.QuoteMeta(marker) + `\b`).MatchString(region)
	hasBad := strings.Contains(region, "REPROVADO")

	var blockers, warnings *int
	if m := blockersRe.FindStringSubmatch(text); m != nil {
		n, _ := strconv.Atoi(m[1])
		blockers = &n
	}
	if m := warningsRe.FindStringSubmatch(text); m != nil {
		n, _ := strconv.Atoi(m[1])
		warnings = &n
	}

	verdict := "INDEFINIDO"
	switch {
	case hasOK && hasBad:
		verdict = "AMBÍGUO"
	case hasOK:
		verdict = "APROVADO"
	case hasBad:
		verdict = "REPROVADO"
	}

	approved := verdict == "APROVADO" && (blockers == nil || *blockers == 0)

	return ReadinessResult{
		Exists: true, Approved: approved, Verdict: verdict,
		Blockers: blockers, Warnings: warnings, Source: source,
	}
}

func verdictRegion(text string) string {
	if m := verdictSectionRe.FindStringSubmatch(text); m != nil {
		return m[1]
	}
	lines := strings.SplitN(text, "\n", 16)
	if len(lines) > 15 {
		lines = lines[:15]
	}
	return strings.Join(lines, "\n")
}
