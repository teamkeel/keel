# keel-cli

keel-cli is a tool to build and deploy services

## Usage

| Command    | Description                                                         |
| ---------- | ------------------------------------------------------------------- |
| build      | Build the application ready for production deployment               |
| completion | Generate the autocompletion script for the specified shell          |
| diff       | Read DB migrations directory, construct the schema and diff the two |
| help       | Help about any command                                              |
| run        | Run the application locally                                         |
| validate   | Validate the Keel schema                                            |
| generate   | Generates requisite types and runtime code for custom functions     |

## Testing

Keel has a built-in testing framework that allows you to test the functionality of the Keel runtime end-to-end.

Complete documentation and examples for the `@teamkeel/testing` package can be found [here](/testing/package/README.md).

## Development

You need the following installed:

- Go `brew install go`
- Node - first install [`pnpm`](https://pnpm.io/installation) then run `pnpm env use --global lts`
- [Docker](https://docs.docker.com/desktop/install/mac-install/)

A working setup will look something like this (paths will vary):

```sh
~/code/keel main $ which go
/usr/local/go/bin/go
~/code/keel main $ which node
/Users/jonbretman/.nvm/versions/node/v16.16.0/bin/node
~/code/keel main $ which docker
/usr/local/bin/docker
```

### Setting up conventional commits

Run the following setup command:

```bash
sh ./scripts/setup.sh
```

### Using the CLI in development

```bash
go run cmd/keel/main.go [cmd] [args]
```

## Building from source

You can build the CLI executable by running:

```bash
make
```

And to interact with the executable version of the CLI, simply run:

```bash
./keel validate -f ...
```

## Contributing

Please read the [Contribution guidelines](/CONTRIBUTING.md) before lodging a PR.
