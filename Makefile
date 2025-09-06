APP := tvrn
PKG := github.com/GizzmoShifu/tvrn

.PHONY: all build run test tidy smoke

all: build

build:
\tgo build -o bin/$(APP) ./cmd/tvrn

run: build
\t./bin/$(APP)

tidy:
\tgo mod tidy

test:
\tgo test ./...
