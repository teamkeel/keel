.PHONY: build proto testdata wasm test testpretty

build:
	go build -o keel cmd/keel/main.go

proto:
	@protoc -I . \
		--go_out=. \
		--go_opt=paths=source_relative \
		proto/schema.proto

testdata:
	@cd ./schema && go run ./tools/generate_testdata.go ./testdata

test:
	go test ./... -count=1

testpretty:
	go test ./... -count=1 -json | gotestpretty

testintegration:
	go test ./integration -count=1 -v

wasm:
	mkdir -p ./packages/wasm/dist
	GOOS=js GOARCH=wasm go build -o ./packages/wasm/dist/keel.wasm ./packages/wasm/lib/main.go
	node ./packages/wasm/encodeWasm.js

prettier:
	npx prettier --write './integration/**/*.{ts,json}'