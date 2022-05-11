# keel-cli

keel-cli is a tool to build and deploy services

## Usage

| Command           | Description                                                                                       |
|-------------------|---------------------------------------------------------------------------------------------------|
|  build            | Build the application ready for production deployment                                             |
|  completion       | Generate the autocompletion script for the specified shell                                        |
|  diff             | Read DB migrations directory, construct the schema and diff the two                               |
|  help             | Help about any command                                                                            |
|  run              | Run the application locally                                                                       |
|  validate         | Validate the Keel schema                                                                          |


## Development

Requires Go 1.18.x

## Updating protobuf definition

Make amends to protobuf definition file (e.g. `./proto/schema.proto`) and regenerate corresponding generated go code:

```
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    proto/schema.proto
```