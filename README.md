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
|  generate         | Generates requisite types and runtime code for custom functions                                   |

## Development

Requires Go 1.18.x

### Setting up

Run the following setup command:

```
sh ./scripts/setup.sh
```

### Using the CLI in development

```
go run cmd/keel/main.go [cmd] [args]
```

## Building from source

You can build the CLI executable by running:

```
make
```

And to interact with the executable version of the CLI, simply run:

```
./keel validate -f ...
```

# Contributing

Please read the [Contribution guidelines](/CONTRIBUTING.md) before lodging a PR.