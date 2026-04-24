.PHONY: build dev deploy test wasm-build db-init db-init-remote clean

build:
	go run github.com/syumai/workers/cmd/workers-assets-gen -mode=go
	GOOS=js GOARCH=wasm go build -o ./build/app.wasm .

wasm-build:
	GOOS=js GOARCH=wasm go build -o ./build/app.wasm .

dev:
	npx wrangler dev

deploy:
	npx wrangler deploy

test:
	go test ./...

smoke-test:
	./scripts/test_rules.sh

db-init:
	npx wrangler d1 execute cf-flag --file=./schema.sql

db-init-remote:
	npx wrangler d1 execute cf-flag --remote --file=./schema.sql

clean:
	rm -rf ./build
