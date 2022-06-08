.PHONY: build proto testdata

build:
	go build -o keel cmd/keel/main.go

proto:
	@protoc -I . \
		--go_out=. \
		--go_opt=paths=source_relative \
		proto/schema.proto

testdata:
	@go run schema/tools/initialize-expected-proto-files.go $$(pwd)/schema/testdata

