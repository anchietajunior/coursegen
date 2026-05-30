// Package validator performs a lightweight, heuristic acceptance check on a
// generated lesson. It is intentionally lenient (case-insensitive section
// presence) so a slightly differently phrased heading doesn't discard otherwise
// good output. The executor decides whether issues fail or merely warn.
package validator

import (
	"fmt"
	"strings"
)

const minBytes = 400

var requiredSections = []string{
	"Título", "Objetivo", "Contexto", "Motivação",
	"Explicação conceitual", "Explicação técnica", "Exemplo prático",
	"Boas práticas", "Erros comuns", "Checklist de aprendizado",
	"Exercício da aula", "Resumo final",
}

// CheckLesson returns the list of issues found (empty == valid).
func CheckLesson(content string) []string {
	var issues []string
	if len(content) < minBytes {
		issues = append(issues, fmt.Sprintf("conteúdo muito curto (%d bytes < %d)", len(content), minBytes))
	}

	lower := strings.ToLower(content)
	var missing []string
	for _, s := range requiredSections {
		if !strings.Contains(lower, strings.ToLower(s)) {
			missing = append(missing, s)
		}
	}
	if len(missing) > 0 {
		issues = append(issues, "seções ausentes: "+strings.Join(missing, ", "))
	}
	return issues
}
