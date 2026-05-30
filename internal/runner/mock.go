package runner

import (
	"fmt"
	"regexp"
	"time"
)

var idsRe = regexp.MustCompile(`lesson-\d{2}-\d{2}`)

// MockRunner is a deterministic, zero-cost runner that exercises the whole
// pipeline WITHOUT calling a real agent or spending tokens. It synthesizes a
// valid lesson (all required sections) from the output path.
type MockRunner struct{ name string }

func NewMockRunner(name string) *MockRunner { return &MockRunner{name: name} }

func (m *MockRunner) Name() string    { return m.name }
func (m *MockRunner) Available() bool { return true }
func (m *MockRunner) Version() string { return "mock" }

func (m *MockRunner) Run(inv Invocation) Result {
	ids := idsRe.FindString(inv.OutputPath)
	if ids == "" {
		ids = "lesson"
	}
	body := synthesize(ids)
	time.Sleep(50 * time.Millisecond) // make durations non-zero

	return Result{
		Status: StatusOK, Artifact: body, Stdout: body,
		ExitCode: 0, Duration: 50 * time.Millisecond,
	}
}

func synthesize(ids string) string {
	return fmt.Sprintf(`# Título
Aula %s (mock)

## Objetivo
Conteúdo gerado pelo runner `+"`mock`"+` para validar o pipeline (%s).

## Contexto
Texto de exemplo.

## Motivação
Texto de exemplo.

## Explicação conceitual
Texto de exemplo.

## Explicação técnica
Texto de exemplo.

## Exemplo prático
Texto de exemplo.

## Exemplo de código
`+"```go"+`
fmt.Println("exemplo")
`+"```"+`

## Boas práticas
- Exemplo.

## Erros comuns
- Exemplo.

## Checklist de aprendizado
- [ ] Exemplo.

## Exercício da aula
Exemplo.

## Resumo final
Texto de exemplo.
`, ids, ids)
}
