DB_NAME ?= cf-flag

.PHONY: build dev deploy test wasm-build db-create db-init db-init-remote db-clean db-clean-remote clean

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

db-create:
	npx wrangler d1 create $(DB_NAME)

db-init:
	npx wrangler d1 execute $(DB_NAME) --file=./schema.sql

db-init-remote:
	npx wrangler d1 execute $(DB_NAME) --remote --file=./schema.sql

db-clean:
	npx wrangler d1 execute $(DB_NAME) --file=./reset.sql

db-clean-remote:
	npx wrangler d1 execute $(DB_NAME) --remote --file=./reset.sql

clean:
	rm -rf ./build
