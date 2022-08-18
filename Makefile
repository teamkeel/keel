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

wasm:
	GOOS=js GOARCH=wasm go build -o ./wasm/keel.wasm ./wasm/main.go