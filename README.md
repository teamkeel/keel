<p align="center">
  <a href="https://keel.so/">
    <img alt="Keel" src="https://raw.githubusercontent.com/teamkeel/keel/main/docs/images/keel-light.svg" width="300" />
  </a>
</p>

  <p align="center">Your production-grade backend from one file<p>

<p align="center">
   <a href="https://docs.keel.so">Documentation</a> | <a href="https://keel.so/discord">Join our Discord</a>
</p>

---

# Keel

Keel is the all-in-one platform that solves the backend puzzle by combining all the essential technology and tools your team needs to build a business that thrives.

🔥 A fully-managed relational database  
🔨 Scalable infrastructure  
🦄 A great local development experience  
🤝 Flexible and customisable API's  
👻 Authentication & Permissions  
🌍 Multiple environments  
⚡️ Configuration and secret management  
🧐 Observability  
🚀 Git-based deployments  
🎟️ Pub/Sub  
💼 Cron jobs 

Keel is currently in private beta, [join the waiting list](https://keel.so/) for access to the platform. In the meantime you can run your Keel projects locally without an account using the CLI.

## Getting started

The best way to get started with Keel is with the CLI. 

### Dependencies

First, you will need to install [Node.js](https://nodejs.org/) and [Docker](https://docker.com/) as they are required to run Keel. 

### Installing the CLI

Install the CLI globally:

```bash
$ npm install --global keel
```

If the CLI has installed correctly, then `keel --version` will output the latest version number:

```bash
$ keel --version
```

### Creating a new project

Creating a new project with the CLI is easy! You can create a new Keel project locally by running:

```bash
$ mkdir my-app && cd my-app
$ keel init
```

This will create a skeleton project containing a `keel.schema` file, a `.gitignore` and a `keelconfig.yaml` file.

_Run `keel init --help` to see all of the available options for `init`._

### Setting up your editor

Keel has created a handy [VSCode extension](https://marketplace.visualstudio.com/items?itemName=teamkeel.vscode-keel) that accelerates your development experience by providing syntax highlighting, autocompletions and schema validation.

### Your very first Keel schema

To create your first Keel schema, follow our [Basic Tutorial](https://docs.keel.so/tutorial#the-keel-schema).

### Running your Keel app

Now that you've created your first Keel schema, you can start using your APIs by starting the Keel development server:

```bash
$ keel run
```

_Run `keel run --help` to see all of the available options for `run`._

If everything has been successful, you should see something like this in your terminal:

![Keel CLI](docs/images/cli-run.png)

You can now interact with your APIs at the URLs displayed. For more next steps, take a look at our [documentation](https://docs.keel.so/).

Have fun with your new Keel backend!

### CLI Reference

This is an overview of the commands available to you once you are set up.

| Command    | Description                                                         |
| ---------- | ------------------------------------------------------------------- |
| init       | Initialise a new Keel project                                       |
| [run](https://docs.keel.so/cli#running-your-project)        | Run the application locally                                         |
| [validate](https://docs.keel.so/cli#validating)   | Validate a Keel schema                                            |
| [generate](https://docs.keel.so/cli#generating-code)   | Generates dynamic JavaScript code from your Keel schema          |
| [secrets](https://docs.keel.so/secrets#secrets-in-development)    | Add, remove & list secrets                                          |
| help       | Gives extra information about Keel CLI commands                                       |

If you want to know more about any of the CLI commands, you can run `keel [command] --help`. 

## Contributing

If you would like to build and run the Keel Runtime locally, you can find out how to get set up for development from our [Contributors](CONTRIBUTING.md) page. 

## Reporting a problem

If you need help, have questions, or want to contact us for any reason, you can do so by emailing us at [help@keel.so](mailto:help@keel.so) or by visiting our [Discord community](https://keel.so/discord).

If you find a bug and want to raise an issue, you can use [GitHub Issues](https://github.com/teamkeel/keel/issues) to do so. Please try to include as much information as you can, including steps to reproduce the issue, your expected outcome, the actual outcome and any details that could help such as envirnment details, traces or logs.

We hold our values very dearly at Keel, and want our communities to be respectful, welcoming places to engage in. As such, we have a [Code Of Conduct](CODE_OF_CONDUCT.md) governing this project that we take very seriously. If you feel that you have had an experience while engaging with this project that violates this code of conduct and you want to report it, please do so by emailing us at [help@keel.so](mailto:help@keel.so). Please include as much detail as you feel able to so that we can look after you as well as we can.