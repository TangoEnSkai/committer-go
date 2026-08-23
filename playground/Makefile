.PHONY: build

# Go 1.24 moved wasm_exec.js from $GOROOT/misc/wasm to $GOROOT/lib/wasm.
# Prefer the new location and fall back to the old one for older toolchains.
GOROOT := $(shell go env GOROOT)
WASM_EXEC := $(firstword $(wildcard $(GOROOT)/lib/wasm/wasm_exec.js $(GOROOT)/misc/wasm/wasm_exec.js))

build:
ifeq ($(WASM_EXEC),)
	$(error wasm_exec.js not found under $(GOROOT)/lib/wasm or $(GOROOT)/misc/wasm)
endif
	mkdir -p static
	GOOS=js GOARCH=wasm go build -o static/app.wasm ../playground/
	cp "$(WASM_EXEC)" static/
	cp index.html static/
