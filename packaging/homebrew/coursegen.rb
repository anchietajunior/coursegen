# Homebrew formula para o CourseGen.
#
# Como publicar (resumo):
#   1. Crie um tap: repositório `coursegen/homebrew-tap`.
#   2. Faça um release v0.1.0 (tag) no repo do coursegen.
#   3. Calcule o sha256 do tarball e preencha abaixo.
#   4. Copie este arquivo para o tap como Formula/coursegen.rb.
#   5. Usuário: `brew install coursegen/tap/coursegen`.
#
# Os assets (skills de planejamento) já vêm commitados em internal/assets/skills,
# então `go build` basta — não há download extra em runtime.
class Coursegen < Formula
  desc "Orquestra a produção de aulas de cursos com agentes de IA (uma aula por sessão isolada)"
  homepage "https://github.com/coursegen/coursegen"
  url "https://github.com/coursegen/coursegen/archive/refs/tags/v0.1.0.tar.gz"
  sha256 "REPLACE_WITH_TARBALL_SHA256"
  license "MIT"
  head "https://github.com/coursegen/coursegen.git", branch: "main"

  depends_on "go" => :build

  def install
    ldflags = "-s -w"
    system "go", "build", "-ldflags", ldflags, "-o", bin/"coursegen", "./cmd/coursegen"
  end

  def caveats
    <<~EOS
      Para instalar as skills de planejamento no seu agente (claude, codex,
      gemini, cursor, opencode) e escolher o agente padrão, rode:

        coursegen setup

      Depois, dentro de um projeto de curso aprovado:

        coursegen generate lessons
    EOS
  end

  test do
    assert_match "coursegen 0.1.0", shell_output("#{bin}/coursegen version")
  end
end
