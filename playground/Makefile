.PHONY: build

build:
	mkdir -p static
	GOOS=js GOARCH=wasm go build -o static/app.wasm ../playground/
	cp $$(go env GOROOT)/misc/wasm/wasm_exec.js static/
	cp index.html static/
