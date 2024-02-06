#!/usr/bin/env node

"use strict";

const { https } = require("follow-redirects"),
  path = require("path"),
  tar = require("tar"),
  zlib = require("zlib"),
  mkdirp = require("mkdirp"),
  fs = require("fs"),
  exec = require("child_process").exec;

// Mapping from Node's `process.arch` to Golang's `$GOARCH`
const ARCH_MAPPING = {
  ia32: "386",
  x64: "amd64",
  arm: "arm",
  arm64: "arm64",
};

// Mapping between Node's `process.platform` to Golang's
const PLATFORM_MAPPING = {
  darwin: "darwin",
  linux: "linux",
  win32: "windows",
  freebsd: "freebsd",
};

function getInstallationPath(callback) {
  // `npm prefix -g` will output the path where the `bin` folder is available
  exec("npm prefix -g", function (err, stdout, stderr) {
    let dir = null;
    if (err || stderr || !stdout || stdout.length === 0) {
      // We couldn't infer path from `npm prefix`. Let's try to get it from
      // Environment variables set by NPM when it runs.
      // npm_config_prefix points to NPM's installation directory where `bin` folder is available
      // Ex: /Users/foo/.nvm/versions/node/v4.3.0
      let env = process.env;
      if (env && env.npm_config_prefix) {
        dir = path.join(env.npm_config_prefix, "bin");
      }
    } else {
      dir = path.join(stdout.trim(), "bin");
    }

    mkdirp.sync(dir);

    callback(null, dir);
  });
}

function verifyAndPlaceBinary(binName, binPath, callback) {
  if (!fs.existsSync(path.join(binPath, binName))) {
    return callback(
      `Downloaded binary does not contain the binary specified in configuration - ${binName}`
    );
  }

  getInstallationPath(function (err, installationPath) {
    if (err) {
      return callback(
        "Error getting binary installation path from `npm root -g`"
      );
    }

    // Move the binary file
    fs.renameSync(
      path.join(binPath, binName),
      path.join(installationPath, binName)
    );

    callback(null);
  });
}

function validateConfiguration(packageJson) {
  if (!packageJson.version) {
    return "'version' property must be specified";
  }

  if (!packageJson.goBinary || typeof packageJson.goBinary !== "object") {
    return "'goBinary' property must be defined and be an object";
  }

  if (!packageJson.goBinary.name) {
    return "'name' property is necessary";
  }

  if (!packageJson.goBinary.path) {
    return "'path' property is necessary";
  }

  if (!packageJson.goBinary.url) {
    return "'url' property is required";
  }
}

function parsePackageJson() {
  if (!(process.arch in ARCH_MAPPING)) {
    console.error(
      "Installation is not supported for this architecture: " + process.arch
    );
    return;
  }

  if (!(process.platform in PLATFORM_MAPPING)) {
    console.error(
      "Installation is not supported for this platform: " + process.platform
    );
    return;
  }

  const packageJsonPath = path.join(".", "package.json");
  if (!fs.existsSync(packageJsonPath)) {
    console.error(
      "Unable to find package.json. " +
        "Please run this script at root of the package you want to be installed"
    );
    return;
  }

  let packageJson = JSON.parse(fs.readFileSync(packageJsonPath));
  let error = validateConfiguration(packageJson);
  if (error && error.length > 0) {
    console.error("Invalid package.json: " + error);
    return;
  }

  // We have validated the config. It exists in all its glory
  let binName = packageJson.goBinary.name;
  let binPath = packageJson.goBinary.path;
  let url = packageJson.goBinary.url;
  let version = packageJson.version;
  if (version[0] === "v") version = version.substr(1); // strip the 'v' if necessary v0.0.1 => 0.0.1

  // Binary name on Windows has .exe suffix
  if (process.platform === "win32") {
    binName += ".exe";
  }

  // Interpolate variables in URL, if necessary
  url = url.replace(/{{arch}}/g, ARCH_MAPPING[process.arch]);
  url = url.replace(/{{platform}}/g, PLATFORM_MAPPING[process.platform]);
  url = url.replace(/{{version}}/g, version);
  url = url.replace(/{{bin_name}}/g, binName);

  return {
    binName: binName,
    binPath: binPath,
    url: url,
    version: version,
  };
}

/**
 * Reads the configuration from application's package.json,
 * validates properties, downloads the binary, untars, and stores at
 * ./bin in the package's root. NPM already has support to install binary files
 * specific locations when invoked with "npm install -g"
 *
 *  See: https://docs.npmjs.com/files/package.json#bin
 */
function install(callback) {
  let opts = parsePackageJson();
  if (!opts) return callback("Invalid inputs");

  console.log(opts);

  mkdirp.sync(opts.binPath);
  let ungz = zlib.createGunzip();
  let untar = tar.extract({ cwd: opts.binPath });

  ungz.on("error", callback);
  untar.on("error", callback);

  // First we will Un-GZip, then we will untar. So once untar is completed,
  // binary is downloaded into `binPath`. Verify the binary and call it good
  untar.on(
    "end",
    verifyAndPlaceBinary.bind(null, opts.binName, opts.binPath, callback)
  );

  console.log("Downloading from URL: " + opts.url);

  const request = https
    .get(opts.url, function (response) {
      if (response.statusCode !== 200) {
        return callback(
          "Error downloading binary. HTTP Status Code: " + response.statusCode
        );
      }

      response.pipe(ungz).pipe(untar);
    })
    .on(
      "error",
      callback.bind(null, "Error downloading from URL: " + opts.url)
    );
}

function uninstall(callback) {
  let opts = parsePackageJson();
  getInstallationPath(function (err, installationPath) {
    if (err) callback("Error finding binary installation directory");

    try {
      fs.unlinkSync(path.join(installationPath, opts.binName));
    } catch (ex) {
      // Ignore errors when deleting the file.
    }

    return callback(null);
  });
}

// Parse command line arguments and call the right method
let actions = {
  install: install,
  uninstall: uninstall,
};

let argv = process.argv;
if (argv && argv.length > 2) {
  let cmd = process.argv[2];
  if (!actions[cmd]) {
    console.log("`install` and `uninstall` are the only supported commands");
    process.exit(1);
  }

  actions[cmd](function (err) {
    if (err) {
      console.error(err);
      process.exit(1);
    } else {
      process.exit(0);
    }
  });
}
