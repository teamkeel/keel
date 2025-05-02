.PHONY: build proto testdata wasm test testpretty

# Supply PACKAGES arg to only run tests for one page e.g. PACKAGES=./traing
PACKAGES?=./...

# Supply RUN to only run some tests e.g. `RUN=TestMyFunction make test`
RUNARG=
ifdef RUN
# If running only some tests add -v for more verbose output
RUNARG=-run $(RUN) -v
endif

build:
	export CGO_ENABLED=0 && go build -o ./bin/keel cmd/keel/main.go

proto:
	nix-shell --command "protoc -I . \
		--go_out=. \
		--go_opt=paths=source_relative \
		proto/schema.proto"

proto-tools:
	PROTO_SRC_PATH=./ \
	IMPORT_MAPPING="rpcutil/empty.proto=github.com/example/rpcutil" \
	protoc \
		--proto_path=tools/proto/ \
		--proto_path=proto/ \
		--go_opt=Mschema.proto=github.com/teamkeel/keel/proto \
		--go_opt=Mtools.proto=tools/proto \
		--twirp_out=Mtools.proto=tools/proto:tools/proto \
		--go_out=./ \
		tools.proto

test:
	TZ=UTC go test $(PACKAGES) -count=1 $(RUNARG)

test-js:
	cd ./packages/functions-runtime && pnpm run test
	cd ./packages/testing-runtime && pnpm run test
	cd ./packages/wasm && pnpm run test

lint:
	export CGO_ENABLED=0 && golangci-lint run  -c .golangci.yml

wasm:
	mkdir -p ./packages/wasm/dist
	GOOS=js GOARCH=wasm go build -o ./packages/wasm/dist/keel.wasm ./packages/wasm/lib/main.go
	node ./packages/wasm/encodeWasm.js

prettier:
	npx prettier@3.0.0 --write './integration/**/*.{ts,json,yaml}'
	npx prettier@3.0.0 --write './packages/**/*.{ts,js,mjs,tsx}'
	npx prettier@3.0.0 --write './node/templates/**/*.{ts,js,mjs}'
	npx prettier@3.0.0 --write './packages/**/package.json'
	npx prettier@3.0.0 --write './schema/testdata/proto/**/*.json'
	npx prettier@3.0.0 --write './runtime/jsonschema/testdata/**/*.json'
	npx prettier@3.0.0 --write './runtime/openapi/testdata/**/*.json'
	npx prettier@3.0.0 --write './tools/testdata/**/*.json'

install:
	go mod download
	cd ./packages/functions-runtime && pnpm install
	cd ./packages/testing-runtime && pnpm install
	cd ./packages/wasm && pnpm install
	cd ./packages/client-react && pnpm install
	cd ./packages/client-react-query && pnpm install

setup-conventional-commits:
	brew install pre-commit -q
	pre-commit install --hook-type commit-msg

goreleaser:
	rm -rf dist
	goreleaser release --snapshot

# Generate the defualt private key used by the run and test commands
gen-pk:
	go run ./util/privatekey/pkgen.go ./testing/testing/default.pem
	
rpc-api:
	PROTO_SRC_PATH=./ \
	IMPORT_MAPPING="rpcutil/empty.proto=github.com/example/rpcutil" \
	protoc \
		--proto_path=rpc/ \
		--proto_path=proto/ \
		--proto_path=tools/proto/ \
		--go_opt=Mschema.proto=github.com/teamkeel/keel/proto \
		--go_opt=Mrpc.proto=rpc/rpc \
		--go_opt=Mtools.proto=github.com/teamkeel/keel/tools/proto \
		--twirp_out=Mrpc.proto=rpc/rpc:rpc/rpc \
		--go_out=./ \
		rpc.proto


proto-tests:
	nix-shell --command "cd ./schema && go run ./testdata/generate_testdata.go ./testdata/proto"
proto-tools-tests:
	cd ./tools && go run ./testdata/generate_testdata.go ./testdata

