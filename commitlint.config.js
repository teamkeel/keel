module.exports = {
  extends: ['@commitlint/config-conventional'],
  rules: {
    'body-max-length': [2, 'always', 300],
    'body-case': [1, "always", "sentence-case"],
    'subject-case': [2, "always", ["sentence-case", "lower-case"]]
  }
}
