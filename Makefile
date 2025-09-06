APP := tvdb
SMOKE := smoke-tvdb
PKG := github.com/GizzmoShifu/tvrn

.PHONY: all build run test tidy smoke

all: build

build:
\tgo build -o bin/$(APP) ./cmd/tvdb
\tgo build -o bin/$(SMOKE) ./cmd/smoke-tvdb

run: build
\t./bin/$(APP)

tidy:
\tgo mod tidy

test:
\tgo test ./...
