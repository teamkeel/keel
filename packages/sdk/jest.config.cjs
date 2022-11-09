const path = require("path");

module.exports = {
  preset: "ts-jest",
  testEnvironment: "node",
  verbose: true,
  moduleNameMapper: {
    chalk: require.resolve("chalk"),
    "#ansi-styles": path.join(
      require.resolve("chalk").split("chalk")[0],
      "chalk/source/vendor/ansi-styles/index.js"
    ),
    "#supports-color": path.join(
      require.resolve("chalk").split("chalk")[0],
      "chalk/source/vendor/supports-color/index.js"
    ),
  },
};
