BINARY := coursegen
PKG    := ./cmd/coursegen
VERSION := $(shell sed -n 's/.*Version = "\(.*\)"/\1/p' internal/cli/cli.go)

# Plataformas para release (binário único por OS/arch).
PLATFORMS := darwin/arm64 darwin/amd64 linux/amd64 linux/arm64 windows/amd64

.PHONY: build install vet test fmt clean release run sync-skills

# Copia as skills (fonte da verdade em skills/*.md) para o local embarcável.
# Necessário antes do build; o resultado é commitado para que `go install` funcione.
sync-skills:
	@rm -rf internal/assets/skills
	@for f in skills/*.md; do \
		name=$$(basename "$$f" .md); \
		[ "$$name" = "README" ] && continue; \
		mkdir -p internal/assets/skills/$$name; \
		cp "$$f" internal/assets/skills/$$name/SKILL.md; \
	done
	@echo "→ skills sincronizadas para internal/assets/skills/"

build: sync-skills ## Compila o binário em ./bin/coursegen
	@mkdir -p bin
	go build -o bin/$(BINARY) $(PKG)
	@echo "→ bin/$(BINARY) ($(VERSION))"

install: ## Instala o binário no GOBIN/PATH
	go install $(PKG)

vet:
	go vet ./...

test:
	go test ./...

fmt:
	gofmt -w .

clean:
	rm -rf bin dist

run: build ## Ex.: make run ARGS="generate lessons --runner mock"
	./bin/$(BINARY) $(ARGS)

release: sync-skills ## Cross-compila binários estáticos para todas as plataformas em ./dist
	@mkdir -p dist
	@for p in $(PLATFORMS); do \
		os=$${p%/*}; arch=$${p#*/}; \
		ext=""; [ "$$os" = "windows" ] && ext=".exe"; \
		echo "build $$os/$$arch"; \
		CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch \
		  go build -ldflags "-s -w" -o dist/$(BINARY)-$$os-$$arch$$ext $(PKG); \
	done
	@echo "→ binários em ./dist (sem runtime, sem CGO)"
