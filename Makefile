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
	npx prettier --write './integration/**/*.{ts,json,yaml}'
	npx prettier --write './packages/**/*.{ts,js,mjs}'

install:
	go mod download
	npm install
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
	