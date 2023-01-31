# keel-cli

keel-cli is a tool to build and deploy services.

# Note

Keel is currently pre-release (semver < 1). _Do not commit any major releases using conventional commits to bump the version to v1 until we have decided that this is the case_

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

### Nix

We use [nix](https://nix.dev/) to manage our development environment. To install nix on MacOS run:

```sh
$ curl -L https://nixos.org/nix/install | sh
```

Once you have nix installed you can start a nix shell by running `nix-shell` in the root of the repo. The first time you run this it will install all the required dependencies so might take a while to run, but subsequently should be fast. When developing in this repo always start by starting a nix-shell.

### Docker

You also need Docker (for runnning Postgres) which needs to be [installed separately](https://docs.docker.com/desktop/install/mac-install/). Once installed the Docker daemon needs to be running, which you can check by running:

```sh
$ docker ps
```

### Installing project dependencies

Running `make install` will install project level dependencies.

### Setting conventional commits git hook (Optional)

This repo follows [conventional commits](https://www.conventionalcommits.org/en/v1.0.0/), which means commits must be written in a certain way. If you want you can run `make setup-conventional-commits` to install a pre-commit hook which will check your commit messages as you make them.

### Running Go tests

To run tests first make sure you have running Postgres using Docker:

```sh
docker compose up -d
```

Then use `make test` to run the tests.

```sh
# run all tests
$ make test

# run all test in the schema package (and all sub-packages)
$ make test PACKAGES=./schema/...

# A specific test from the integration package
$ make test PACKAGES=./integration RUN=TestIntegration/built_in_actions
```

### Running JS tests

There are units tests in each of the JS packages in the `./packages` directory. In each of these directories you can run `pnpm run test` but to run all the tests from all the packages you can use the make command `make test-js`.

### Other useful make commands

- `make lint` - lint Go code
- `make testdata` - re-generate fixture data (check the diff carefully after doing this)
- `make proto` - re-generate Go code in the `./proto` package (run this after changing the `.proto` file)

### Using the CLI in development

```bash
go run cmd/keel/main.go [cmd] [args]
```

## Building from source

You can build the CLI by running:

```bash
make build
```

To then use the built CLI binary you can run:

```bash
./bin/keel validate -f ...
```

## Contributing

Please read the [Contribution guidelines](/CONTRIBUTING.md) before lodging a PR.
