.PHONY: build proto testdata

build:
	go build -o keel cmd/keel/main.go

proto:
	@protoc -I . \
		--go_out=. \
		--go_opt=paths=source_relative \
		proto/schema.proto

testdata:
	@cd ./schema && go run ./tools/generate_testdata.go ./testdata

