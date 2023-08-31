# Contributing

This repo is setup so that all contributors follow [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/) conventions when writing commit messages.

__You will not be able to merge your PR without following these conventions.__

#### Important:
- Please ensure that you have read and understood the Contributing guidelines before creating a PR.
- This project is currently focused on bug fixes, nit fixes, documentation, and tests.
- **We are NOT accepting pull requests for new features or refactor contributions at this time.**

### A short primer on conventional commit messages

Prefix your commit messages with:

- For minor version bumps: `feat: some commit message`
- For patches / fixes: `fix: some commit message`
- For major bumps (note the `!`): `fix!: a major / breaking change`

## Note

Keel is currently pre-release (semver < 1). _Do not commit any major releases using conventional commits to bump the version to v1 until we have decided that
this is the case_

---

## Development

### Nix

We use [nix](https://nix.dev/) to manage our development environment. To install
nix on MacOS run:

```sh
$ curl -L https://nixos.org/nix/install | sh
```

Once you have nix installed you can start a nix shell by running `nix-shell` in
the root of the repo. The first time you run this it will install all the
required dependencies so might take a while to run, but subsequently should be
fast. When developing in this repo always start by starting a nix-shell.

### Docker

You also need Docker (for runnning Postgres) which needs to be
[installed separately](https://docs.docker.com/desktop/install/mac-install/).
Once installed the Docker daemon needs to be running, which you can check by
running:

```sh
$ docker ps
```

### Installing project dependencies

Running `make install` will install project level dependencies.

### Setting conventional commits git hook (Optional)

This repo follows
[conventional commits](https://www.conventionalcommits.org/en/v1.0.0/), which
means commits must be written in a certain way. If you want you can run
`make setup-conventional-commits` to install a pre-commit hook which will check
your commit messages as you make them.

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

There are units tests in each of the JS packages in the `./packages` directory.
In each of these directories you can run `pnpm run test` but to run all the
tests from all the packages you can use the make command `make test-js`.

### Other useful make commands

- `make lint` - lint Go code
- `make testdata` - re-generate fixture data (check the diff carefully after
  doing this)
- `make proto` - re-generate Go code in the `./proto` package (run this after
  changing the `.proto` file)

### Using the CLI in development

```bash
go run cmd/keel/main.go [cmd] [args]
```

### Configuring a private key in development

The `-private-key-path` argument on the run command lets you pass in the path
to a file containing a private key in PEM format. This may be useful if you
need to perform development or testing with signed tokens.

```bash
go run cmd/keel/main.go run -private-key-path private-key.pem
```

You can generate a private key in PEM with `openssl genrsa -out private-key.pem 2048`
or using some online [RSA key generator](https://travistidwell.com/jsencrypt/demo/index.html).


### Building from source

You can build the CLI by running:

```bash
make build
```

To then use the built CLI binary you can run:

```bash
./bin/keel validate -f ...
```

---

## Testing

Keel has a built-in testing framework that allows you to test the functionality
of the Keel runtime end-to-end.

Complete documentation and examples for the `@teamkeel/testing` package can be
found [here](/testing/package/README.md).

### Generating new schema test cases

There is a handy helper script to generate the relevant test case files if you'd like to write a new test case for something in the schema / validation rules:

```
sh ./scripts/new-test-case.sh test_case_name
```

This will generate:

- A blank `schema.keel` file
- An `errors.json` file where you can assert what errors you are expecting

### Updating test cases

If you have made a change to error message copy, then you can run `make testdata` and all of the test case JSON files will be updated to the latest copy.

### A note on test cases

Naming convention for test cases should describe the units you are testing in a hierarchical manner. Test cases covering validations should begin with `validation_` for example, and subsequent fragments should describe which validation is being tested.

If a test case is complex, or the name itself doesn't adequately describe what it is testing, then either reconsider the naming, or consider adding a `description.md` file into the directory where the test case lives to provide more colour.

### Get in touch
If you run into any kind of issue, or have any questions, don't hesitate to ask for help. You can get in touch with us via the following channels:

- **Community Discord**
You can join our Discord community by clicking this [link](https://discord.gg/9G7Uv6JGQM)
- **Email**
Get in touch with us via email at [support@keel.so](mailto:support@keel.so)
