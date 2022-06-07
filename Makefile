.PHONY: build

build:
	go build -o keel cmd/keel/main.go

proto:
	@protoc -I . \
		--go_out=. \
		--go_opt=paths=source_relative \
		proto/schema.proto

