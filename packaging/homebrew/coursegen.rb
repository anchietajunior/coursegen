# Homebrew formula para o CourseGen.
#
# Como publicar (resumo):
#   1. Crie um tap: repositório `anchietajunior/homebrew-tap`.
#   2. Faça um release vX.Y.Z (tag) no repo do coursegen.
#   3. Calcule o sha256 do tarball e preencha abaixo.
#   4. Copie este arquivo para o tap como Formula/coursegen.rb.
#   5. Usuário: `brew install anchietajunior/tap/coursegen`.
#
# A CLI tem ZERO dependências externas (100% stdlib Go), então não há `vendor/`
# nem download em runtime — `go build` compila offline na sandbox do Homebrew.
# As skills de planejamento já vêm embarcadas em internal/assets/skills.
class Coursegen < Formula
  desc "Orquestra a producao de aulas de cursos com agentes de IA (uma aula por sessao isolada)"
  homepage "https://github.com/anchietajunior/coursegen"
  url "https://github.com/anchietajunior/coursegen/archive/refs/tags/v0.1.2.tar.gz"
  sha256 "REPLACE_WITH_TARBALL_SHA256"
  license "MIT"
  head "https://github.com/anchietajunior/coursegen.git", branch: "main"

  depends_on "go" => :build

  def install
    system "go", "build", "-trimpath", "-ldflags", "-s -w",
           "-o", bin/"coursegen", "./cmd/coursegen"
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
    assert_match "coursegen 0.1.2", shell_output("#{bin}/coursegen version")
  end
end
