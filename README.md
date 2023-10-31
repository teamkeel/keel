<p align="center">
  <a href="https://keel.so/">
    <img alt="Keel" src="docs/keel.svg" width="300" />
  </a>
</p>

  <p align="center">Your production-grade backend from one file<p>

<p align="center">
   <a href="https://docs.keel.so">Documentation</a> | <a href="https://keel.so/discord">Join our Discord</a>
</p>

---


# Keel 

Keel is the all-in-one backend platform that gives you everything you need to build your product.  

🔥 A fully-managed relational database  
🔨 Scalable infrastructure  
🦄 A great local development experience  
🤝 Flexible and customisable API's  
👻 Authentication & Permissions  
🌍 Multiple environments  
⚡️ Configuration and secret management  
🧐 Observability  
🚀 Git-based deployments  

Keel is currently in private beta, [join the waiting list](https://keel.so/) for access to the platform. In the meantime you can run your Keel projects locally without an account using the CLI

## Getting Started

To install and use Keel locally, let's start by building it. Follow these steps to get set up:

1. Clone this repo
2. Ensure you've got the required tools. You need [**Go (Golang)**](https://go.dev) installed locally.
3. Run `make` to build the Keel CLI

The Keel CLI is now available in your `./bin` directory. You can add this to your `PATH` if you want to use it globally.

You will also need [Node.js](https://nodejs.org/) and [Docker](https://docker.com/) to run Keel. Once you grab those, you've now got the full power of Keel locally.

## Creating a New Project

Creating a new project with the CLI is easy! Assuming you've added `./bin/keel` to your `PATH`, you can create a new Keel project locally by running:

```bash
$ mkdir my-keel-project && cd my-keel-project
$ keel init
```

This will create a skeleton project containing a `keel.schema` file, a `.gitignore` and a `.keelconfig.yaml` file.

_Run `keel init --help` to see all of the available options for `init`._

## Setting up Your Editor

Keel has created a handy [VSCode extension](https://marketplace.visualstudio.com/items?itemName=teamkeel.vscode-keel) that accelerates your development experience by providing syntax highlighting, autocompletions and validation of schemas.

## Your Very First Keel Schema

To create your first Keel schema, follow our [Quickstart Tutorial](https://docs.keel.so/get-started/tutorial#the-keel-schema).

### Running your Keel App

Now that you've created your first Keel schema, you can start using your APIs by starting the Keel development server:

```bash
$ keel run
```

_Run `keel run --help` to see all of the available options for `run`._

If everything has been successful, you should see something like this in your terminal:

![Keel CLI](https://i.imgur.com/gwkUdpU.png)

You can now interact with your APIs at the URLs displayed. For more next steps, take a look at our [documentation](https://docs.keel.so/).


Have fun with your new Keel backend!

### CLI Reference

This is an overview of the commands available to you once you are set up.

| Command    | Description                                                         |
| ---------- | ------------------------------------------------------------------- |
| init       | Initialise a new Keel project                                       |
| [run](https://docs.keel.so/docs/cli#validating)        | Run the application locally                                         |
| [validate](https://docs.keel.so/docs/cli#validating)   | Validate a Keel schema                                            |
| [generate](https://docs.keel.so/docs/cli#generating-code)   | Generates dynamic JavaScript code from your Keel schema          |
| [secrets](https://docs.keel.so/docs/secrets#secrets-in-development)    | Add, remove & list secrets                                          |
| help       | Gives extra information about Keel CLI commands                                       |

If you want to know more about any of the CLI commands, you can run `keel [command] --help`. 

## Contributing

We are not currently looking for active contributors at this time as we are still pre-release, but we are happy to accept bug fixes, tests or docs changes. If you would like to build and run the Keel Runtime locally, you can find out how to get set up for development from our [Contributors](CONTRIBUTING.md) page. 

## Reporting a problem

If you need help, have questions, or want to contact us for any reason, you can do so by emailing us at [help@keel.so](mailto:help@keel.so).

If you find a bug and want to raise an issue, you can use [GitHub Issues](https://github.com/teamkeel/keel/issues) to do so. Please try to include as much information as you can, including steps to reproduce the issue, your expected outcome, the actual outcome and any details that could help such as envirnment details, traces or logs.

We hold our values very dearly at Keel, and want our communities to be respectful, welcoming places to engage in. As such, we have a [Code Of Conduct](CODE_OF_CONDUCT.md) governing this project that we take very seriously. If you feel that you have had an experience while engaging with this project that violates this code of conduct and you want to report it, please do so by emailing us at [help@keel.so](mailto:help@keel.so). Please include as much detail as you feel able to so that we can look after you as well as we can.


