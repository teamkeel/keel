# Version: v0.366.2-prerelease0

* [#1090](https://github.com/teamkeel/keel/pull/1090): fix: only use verified emails for roles
* [#1146](https://github.com/teamkeel/keel/pull/1146): fix: row based permission issue
* [#1163](https://github.com/teamkeel/keel/pull/1163): Update contributing.md
* [#1181](https://github.com/teamkeel/keel/pull/1181): fix: use name of nodes for relationship attribute validation errors
* [#1187](https://github.com/teamkeel/keel/pull/1187): Keel run fix (with Windows support)
* [#1194](https://github.com/teamkeel/keel/pull/1194): [Changelog CI] Add Changelog for Version v0.366.0
* [#1198](https://github.com/teamkeel/keel/pull/1198): fix: an action input should only be nullable if the field being targeted is optional
* [#1200](https://github.com/teamkeel/keel/pull/1200): fix: set MessageName field of MessageField for query input messages
* [#1201](https://github.com/teamkeel/keel/pull/1201): fix: audit trigger migrations for existing models
* [#1202](https://github.com/teamkeel/keel/pull/1202): chore: sending events from audit table, discriminating types
* [#1209](https://github.com/teamkeel/keel/pull/1209): Lazy OIDC


# Version: v0.366.0

* [#1154](https://github.com/teamkeel/keel/pull/1154): fix: subscribers not in transactions, fixed opt on testing types
* [#1160](https://github.com/teamkeel/keel/pull/1160): fix: update auth tracing
* [#1161](https://github.com/teamkeel/keel/pull/1161): build: error response and tracing improvements
* [#1168](https://github.com/teamkeel/keel/pull/1168): chore: passing traceparent to event handler
* [#1169](https://github.com/teamkeel/keel/pull/1169): fix: format issues
* [#1170](https://github.com/teamkeel/keel/pull/1170): chore: refactoring test names
* [#1171](https://github.com/teamkeel/keel/pull/1171): feat: identity and tracing context in audit logs
* [#1172](https://github.com/teamkeel/keel/pull/1172): chore: query database module in integration tests
* [#1174](https://github.com/teamkeel/keel/pull/1174): build(release): use base_ref not event ref
* [#1175](https://github.com/teamkeel/keel/pull/1175): chore: audit log tests
* [#1176](https://github.com/teamkeel/keel/pull/1176): build(release): Set pre release envvar
* [#1177](https://github.com/teamkeel/keel/pull/1177): fix: self referencing relationship on many bug
* [#1178](https://github.com/teamkeel/keel/pull/1178): chore: reverts 1172
* [#1179](https://github.com/teamkeel/keel/pull/1179): fix: self referencing relationship on many bug (#1177)
* [#1182](https://github.com/teamkeel/keel/pull/1182): build(release): Update release workflows
* [#1183](https://github.com/teamkeel/keel/pull/1183): fix: self referencing model validation & associations
* [#1184](https://github.com/teamkeel/keel/pull/1184): fix: self referencing model validation & associations (#1183)
* [#1188](https://github.com/teamkeel/keel/pull/1188): build: reworked auditing of identity and trace id entirely
* [#1189](https://github.com/teamkeel/keel/pull/1189): Auditident branch fix
* [#1191](https://github.com/teamkeel/keel/pull/1191): feat: audit logs


## [0.365.14](https://github.com/teamkeel/keel/compare/v0.365.13...v0.365.14) (2023-08-21)


### Bug Fixes

* use go 1.20 ([#1120](https://github.com/teamkeel/keel/issues/1120)) ([f460929](https://github.com/teamkeel/keel/commit/f460929e37a86fe2ccf6aff24643aecb46c5ed61))

## [0.365.13](https://github.com/teamkeel/keel/compare/v0.365.12...v0.365.13) (2023-08-21)


### Bug Fixes

* add composite unique constraint on identity issuer+email ([#1098](https://github.com/teamkeel/keel/issues/1098)) ([a93f5a1](https://github.com/teamkeel/keel/commit/a93f5a170e8dd60045dff3d513865f63afb4b571))

## [0.365.12](https://github.com/teamkeel/keel/compare/v0.365.11...v0.365.12) (2023-08-18)


### Bug Fixes

* change docs link to new doc site ([#1116](https://github.com/teamkeel/keel/issues/1116)) ([cf7a738](https://github.com/teamkeel/keel/commit/cf7a738b201a070329653f0b97fe00c2dbdef932))

## [0.365.11](https://github.com/teamkeel/keel/compare/v0.365.10...v0.365.11) (2023-08-18)


### Bug Fixes

* export fields and set json tags ([#1115](https://github.com/teamkeel/keel/issues/1115)) ([b1ab881](https://github.com/teamkeel/keel/commit/b1ab881c47ac0ff663724366a744ff3f379433ce))

## [0.365.10](https://github.com/teamkeel/keel/compare/v0.365.9...v0.365.10) (2023-08-18)


### Bug Fixes

* enabling resolveJsonModule in tsconfig ([#1114](https://github.com/teamkeel/keel/issues/1114)) ([5ef5ac4](https://github.com/teamkeel/keel/commit/5ef5ac4fcc04a7a8460cf9393582ee9d4f9c33e7))

## [0.365.9](https://github.com/teamkeel/keel/compare/v0.365.8...v0.365.9) (2023-08-17)


### Bug Fixes

* client packages build ([#1112](https://github.com/teamkeel/keel/issues/1112)) ([864a4e0](https://github.com/teamkeel/keel/commit/864a4e0a57af42fe5f90b89df71a0ea2abb8c038))

## [0.365.8](https://github.com/teamkeel/keel/compare/v0.365.7...v0.365.8) (2023-08-17)


### Bug Fixes

* add line break between test output ([#1111](https://github.com/teamkeel/keel/issues/1111)) ([67f1ddd](https://github.com/teamkeel/keel/commit/67f1ddd74892864313f6a44380212d3e11b38e7c))

## [0.365.7](https://github.com/teamkeel/keel/compare/v0.365.6...v0.365.7) (2023-08-17)


### Bug Fixes

* scope fetch to globalThis ([#1110](https://github.com/teamkeel/keel/issues/1110)) ([4ff9f56](https://github.com/teamkeel/keel/commit/4ff9f5688a26fb546c58bed4234aedacdb0b9055))

## [0.365.6](https://github.com/teamkeel/keel/compare/v0.365.5...v0.365.6) (2023-08-16)


### Bug Fixes

* package.json ([#1109](https://github.com/teamkeel/keel/issues/1109)) ([e0e6031](https://github.com/teamkeel/keel/commit/e0e603143660b9218e792a04d676137a15529f5d))

## [0.365.5](https://github.com/teamkeel/keel/compare/v0.365.4...v0.365.5) (2023-08-16)


### Bug Fixes

* client package building ([#1107](https://github.com/teamkeel/keel/issues/1107)) ([183b2db](https://github.com/teamkeel/keel/commit/183b2db4be66e01d72139904d5af78a57f057297))

## [0.365.4](https://github.com/teamkeel/keel/compare/v0.365.3...v0.365.4) (2023-08-16)


### Bug Fixes

* client building ([#1102](https://github.com/teamkeel/keel/issues/1102)) ([b13e21d](https://github.com/teamkeel/keel/commit/b13e21d81018c3bc60cbf96b2d7c53f273be7b90))

## [0.365.3](https://github.com/teamkeel/keel/compare/v0.365.2...v0.365.3) (2023-08-16)


### Bug Fixes

* permission relationship fixes ([#1104](https://github.com/teamkeel/keel/issues/1104)) ([a2a2333](https://github.com/teamkeel/keel/commit/a2a233359787d96fc7ac556b987a6514facd58f3))

## [0.365.2](https://github.com/teamkeel/keel/compare/v0.365.1...v0.365.2) (2023-08-15)


### Bug Fixes

* foreign key missing error message copy ([#1100](https://github.com/teamkeel/keel/issues/1100)) ([4a17340](https://github.com/teamkeel/keel/commit/4a17340bd80a11730b77e3a474a7aaf715255a2e))

## [0.365.1](https://github.com/teamkeel/keel/compare/v0.365.0...v0.365.1) (2023-08-15)


### Bug Fixes

* implement check for unexpected arguments ([#1087](https://github.com/teamkeel/keel/issues/1087)) ([ff55b32](https://github.com/teamkeel/keel/commit/ff55b32e4cd07a51ddd82d0dca56a60d3f09007b))

# [0.365.0](https://github.com/teamkeel/keel/compare/v0.364.3...v0.365.0) (2023-08-14)


### Features

* client improvements ([8c2840a](https://github.com/teamkeel/keel/commit/8c2840acd44061c24b665577ac53378923841905))
* client namespacing ([3ceca0e](https://github.com/teamkeel/keel/commit/3ceca0e0126f5f2b0fbf9629115d62d13251e70d))

## [0.364.3](https://github.com/teamkeel/keel/compare/v0.364.2...v0.364.3) (2023-08-14)


### Bug Fixes

* issuer suffix ([#1095](https://github.com/teamkeel/keel/issues/1095)) ([88cd7b7](https://github.com/teamkeel/keel/commit/88cd7b7ab4b30ecdb7808fe1989d4fdce8826ecc))

## [0.364.2](https://github.com/teamkeel/keel/compare/v0.364.1...v0.364.2) (2023-08-14)


### Bug Fixes

* manual jobs issuer prefix stripping bug ([#1093](https://github.com/teamkeel/keel/issues/1093)) ([59db419](https://github.com/teamkeel/keel/commit/59db4191ff1d703fc56a44b25a886a2434b9f5eb))

## [0.364.1](https://github.com/teamkeel/keel/compare/v0.364.0...v0.364.1) (2023-08-14)


### Bug Fixes

* revert actions schema changes from [#1077](https://github.com/teamkeel/keel/issues/1077) ([#1094](https://github.com/teamkeel/keel/issues/1094)) ([13b3aff](https://github.com/teamkeel/keel/commit/13b3affc5b498c8f808579d9c6048aa746335611))

# [0.364.0](https://github.com/teamkeel/keel/compare/v0.363.2...v0.364.0) (2023-08-14)


### Features

* replace 'operations' and 'functions with new 'actions' keyword ([#1077](https://github.com/teamkeel/keel/issues/1077)) ([60d348b](https://github.com/teamkeel/keel/commit/60d348be8361b8438d0897fc25b5682d339fbb6f))

## [0.363.2](https://github.com/teamkeel/keel/compare/v0.363.1...v0.363.2) (2023-08-12)


### Bug Fixes

* find public key that matches token kid ([#1089](https://github.com/teamkeel/keel/issues/1089)) ([3758264](https://github.com/teamkeel/keel/commit/3758264d74136885e3b80970a80e4f2a1cdfcb6e))

## [0.363.1](https://github.com/teamkeel/keel/compare/v0.363.0...v0.363.1) (2023-08-11)


### Bug Fixes

* client error catching and docs fixes ([49e8ad9](https://github.com/teamkeel/keel/commit/49e8ad95b889361997f3979647c5574ba1d88af3))

# [0.363.0](https://github.com/teamkeel/keel/compare/v0.362.1...v0.363.0) (2023-08-10)


### Features

* scheduled job permission ([#1083](https://github.com/teamkeel/keel/issues/1083)) ([d3c424b](https://github.com/teamkeel/keel/commit/d3c424bd6175de3d179db9cf26448e2ce54c2bfb))

## [0.362.1](https://github.com/teamkeel/keel/compare/v0.362.0...v0.362.1) (2023-08-10)


### Bug Fixes

* default to 'test' keel env' ([#1081](https://github.com/teamkeel/keel/issues/1081)) ([ed5f05a](https://github.com/teamkeel/keel/commit/ed5f05a3f5994fc9b32131c52356889ecf00155a))

# [0.362.0](https://github.com/teamkeel/keel/compare/v0.361.0...v0.362.0) (2023-08-09)


### Features

* give schedule jobs permission when scheduled only ([#1080](https://github.com/teamkeel/keel/issues/1080)) ([c7e6bd8](https://github.com/teamkeel/keel/commit/c7e6bd8dd313a00042bae0cb8d4199596de33f4d))

# [0.361.0](https://github.com/teamkeel/keel/compare/v0.360.1...v0.361.0) (2023-08-08)


### Features

* support external jwt issuers ([#1073](https://github.com/teamkeel/keel/issues/1073)) ([00882ab](https://github.com/teamkeel/keel/commit/00882abe97b846fd6ac321af7b4490bf950ba65e))

## [0.360.1](https://github.com/teamkeel/keel/compare/v0.360.0...v0.360.1) (2023-08-08)


### Bug Fixes

* jobs can have schedule and permission ([#1079](https://github.com/teamkeel/keel/issues/1079)) ([ec750be](https://github.com/teamkeel/keel/commit/ec750beddf408cb5cc4671f8e390a8476fa282e8))

# [0.360.0](https://github.com/teamkeel/keel/compare/v0.359.3...v0.360.0) (2023-08-05)


### Features

* support scheduled jobs ([#1076](https://github.com/teamkeel/keel/issues/1076)) ([542d338](https://github.com/teamkeel/keel/commit/542d3388454ddaff4a87a77adb57c7ce8beb628f))

## [0.359.3](https://github.com/teamkeel/keel/compare/v0.359.2...v0.359.3) (2023-08-04)


### Bug Fixes

* correctly convert int64 to int for linux386 arch ([#1075](https://github.com/teamkeel/keel/issues/1075)) ([b2fd5e9](https://github.com/teamkeel/keel/commit/b2fd5e9fa5fe70cbef5915cf2b8b88b9fe8fa5ef))

## [0.359.2](https://github.com/teamkeel/keel/compare/v0.359.1...v0.359.2) (2023-08-02)


### Bug Fixes

* handle fields marked as unique in one-to-one validation check ([#1071](https://github.com/teamkeel/keel/issues/1071)) ([da7e3c4](https://github.com/teamkeel/keel/commit/da7e3c41688c6d84a74212a0d0de41a0294bb387))

## [0.359.1](https://github.com/teamkeel/keel/compare/v0.359.0...v0.359.1) (2023-07-26)


### Bug Fixes

* improved attribute validations & job permissions validation ([#1059](https://github.com/teamkeel/keel/issues/1059)) ([146666c](https://github.com/teamkeel/keel/commit/146666c585f24393f2a1429fdcb9419900df4777))

# [0.359.0](https://github.com/teamkeel/keel/compare/v0.358.2...v0.359.0) (2023-07-25)


### Features

* add jobs to what keel generate scaffolds ([#1061](https://github.com/teamkeel/keel/issues/1061)) ([c05b0bd](https://github.com/teamkeel/keel/commit/c05b0bd01311c6d2e1c17e0de39074c6349856a3))

## [0.358.2](https://github.com/teamkeel/keel/compare/v0.358.1...v0.358.2) (2023-07-25)


### Bug Fixes

* disabling transactions for jobs ([#1060](https://github.com/teamkeel/keel/issues/1060)) ([5c6ccd5](https://github.com/teamkeel/keel/commit/5c6ccd55fbfa4daf010df0ca85f852aaa411d70f))

## [0.358.1](https://github.com/teamkeel/keel/compare/v0.358.0...v0.358.1) (2023-07-24)


### Bug Fixes

* unique constraint on relationship field ([#1058](https://github.com/teamkeel/keel/issues/1058)) ([dd866d1](https://github.com/teamkeel/keel/commit/dd866d1ece03b4cc06c606122b92f54cc9bc8689))

# [0.358.0](https://github.com/teamkeel/keel/compare/v0.357.2...v0.358.0) (2023-07-24)


### Features

* job permissions  ([#1056](https://github.com/teamkeel/keel/issues/1056)) ([8537aa9](https://github.com/teamkeel/keel/commit/8537aa9de2c080cba2228f6f77ced41f4f0608af))

## [0.357.2](https://github.com/teamkeel/keel/compare/v0.357.1...v0.357.2) (2023-07-21)


### Bug Fixes

* fixed action errors & early auth on functions ([#1054](https://github.com/teamkeel/keel/issues/1054)) ([4620f78](https://github.com/teamkeel/keel/commit/4620f78048f5a1fb8a16c81f6a0cd5c3cb1aeb12))

## [0.357.1](https://github.com/teamkeel/keel/compare/v0.357.0...v0.357.1) (2023-07-20)


### Bug Fixes

* trigger npm ([570ed6d](https://github.com/teamkeel/keel/commit/570ed6dae875ee7ad80a2c6afeea11e30ccb9f81))

# [0.357.0](https://github.com/teamkeel/keel/compare/v0.356.1...v0.357.0) (2023-07-20)


### Bug Fixes

* bump pnpm-locks to v6 ([c81e25f](https://github.com/teamkeel/keel/commit/c81e25fa5d155effa102b2125f4db73b5a2355eb))
* correct pnpm lock files ([4aa5eba](https://github.com/teamkeel/keel/commit/4aa5ebac23e5492495ed373b6564e15a3a763231))


### Features

* client-react and client-react-query packages ([53a55a5](https://github.com/teamkeel/keel/commit/53a55a501b3d155d2d6405017f34e02ea60f1397))

## [0.356.1](https://github.com/teamkeel/keel/compare/v0.356.0...v0.356.1) (2023-07-18)


### Bug Fixes

* create operation nested hasMany optional ([921eca6](https://github.com/teamkeel/keel/commit/921eca63026e82c8ada97ec49c3fc25477b8cdc0))

# [0.356.0](https://github.com/teamkeel/keel/compare/v0.355.0...v0.356.0) (2023-07-18)


### Features

* client auto parse ISO8601 dates ([c21a49a](https://github.com/teamkeel/keel/commit/c21a49aa6732a1bf9cad9fd6448a4101bca0507c))

# [0.355.0](https://github.com/teamkeel/keel/compare/v0.354.1...v0.355.0) (2023-07-14)


### Features

* auto gen default API in schema if none provided ([#1047](https://github.com/teamkeel/keel/issues/1047)) ([e2c8585](https://github.com/teamkeel/keel/commit/e2c858542c9ab3b60e8fd2d51b647f9715b97420))

## [0.354.1](https://github.com/teamkeel/keel/compare/v0.354.0...v0.354.1) (2023-07-12)


### Bug Fixes

* provide completions for composite unique ([#1024](https://github.com/teamkeel/keel/issues/1024)) ([0f6fa6d](https://github.com/teamkeel/keel/commit/0f6fa6dcd5d78e56d605fc3377335bb36ff9d99e))

# [0.354.0](https://github.com/teamkeel/keel/compare/v0.353.0...v0.354.0) (2023-07-12)


### Features

* validation on duplicate action inputs ([#1045](https://github.com/teamkeel/keel/issues/1045)) ([004a31f](https://github.com/teamkeel/keel/commit/004a31ff06ada38249e647ae5733951b2c70bdd3))

# [0.353.0](https://github.com/teamkeel/keel/compare/v0.352.0...v0.353.0) (2023-07-12)


### Features

* [@permission](https://github.com/permission) conditions & early permission checking ([#1041](https://github.com/teamkeel/keel/issues/1041)) ([22afb57](https://github.com/teamkeel/keel/commit/22afb5725ef877b7bfdabffcd4d4e0235fb02f52))
* jobs in runtime & integration tests ([#1044](https://github.com/teamkeel/keel/issues/1044)) ([1ff2b88](https://github.com/teamkeel/keel/commit/1ff2b888547146faa3690a15f2f97405f8663509))

# [0.352.0](https://github.com/teamkeel/keel/compare/v0.351.0...v0.352.0) (2023-07-11)


### Features

* vscode formatting ([#1043](https://github.com/teamkeel/keel/issues/1043)) ([be7b0c4](https://github.com/teamkeel/keel/commit/be7b0c47d5a97ef0a28979fccb43ce5ecf599d94))

# [0.351.0](https://github.com/teamkeel/keel/compare/v0.350.0...v0.351.0) (2023-07-10)


### Features

* generating typescript clients ([ef9e115](https://github.com/teamkeel/keel/commit/ef9e115f95cafc73d11b7ff68ef06b5da61921b5))

# [0.350.0](https://github.com/teamkeel/keel/compare/v0.349.0...v0.350.0) (2023-07-10)


### Features

* capture console logs in traces ([1c76cbc](https://github.com/teamkeel/keel/commit/1c76cbc44324fbb54e008145b5fc5a6c8ef3a8cf))

# [0.349.0](https://github.com/teamkeel/keel/compare/v0.348.0...v0.349.0) (2023-07-10)


### Features

* job definition syntax highlighting and code completions ([#1042](https://github.com/teamkeel/keel/issues/1042)) ([d7848b9](https://github.com/teamkeel/keel/commit/d7848b922d7d2b694930761f942c6a483ecf4e13))

# [0.348.0](https://github.com/teamkeel/keel/compare/v0.347.0...v0.348.0) (2023-07-07)


### Features

* job typescript definitions ([#1039](https://github.com/teamkeel/keel/issues/1039)) ([f4467e3](https://github.com/teamkeel/keel/commit/f4467e33cf817d7a718b0f6bee432af90bdbbb8b))

# [0.347.0](https://github.com/teamkeel/keel/compare/v0.346.0...v0.347.0) (2023-07-06)


### Features

* job schema validation ([#1036](https://github.com/teamkeel/keel/issues/1036)) ([deb0cc5](https://github.com/teamkeel/keel/commit/deb0cc51cc0421611c3be086f96d90522b9736c3))

# [0.346.0](https://github.com/teamkeel/keel/compare/v0.345.3...v0.346.0) (2023-07-06)


### Features

* code completions & syntax highlighting for [@order](https://github.com/order)By and [@sortable](https://github.com/sortable) ([#1033](https://github.com/teamkeel/keel/issues/1033)) ([173ec7f](https://github.com/teamkeel/keel/commit/173ec7fc7fb68807518a4a3e165cc6141defc205))

## [0.345.3](https://github.com/teamkeel/keel/compare/v0.345.2...v0.345.3) (2023-07-05)


### Bug Fixes

* make database exists check more robust ([#1035](https://github.com/teamkeel/keel/issues/1035)) ([d82800f](https://github.com/teamkeel/keel/commit/d82800f5e492793678bb0482f283dbdfaeede1a4))

## [0.345.2](https://github.com/teamkeel/keel/compare/v0.345.1...v0.345.2) (2023-07-04)


### Bug Fixes

* redesign CLI dockerised DB strategy ([#1027](https://github.com/teamkeel/keel/issues/1027)) ([ef08507](https://github.com/teamkeel/keel/commit/ef08507e510d776b8c1c58871e13bee65e304828))
* reinstate package.json ([#1032](https://github.com/teamkeel/keel/issues/1032)) ([3cb5572](https://github.com/teamkeel/keel/commit/3cb5572493de23870f8eb8cd6b95a7fc0a581625))

## [0.345.1](https://github.com/teamkeel/keel/compare/v0.345.0...v0.345.1) (2023-07-04)


### Bug Fixes

* improving [@order](https://github.com/order)By validation ([6561362](https://github.com/teamkeel/keel/commit/656136285603f4416ef891187f138428acd0740d))

# [0.345.0](https://github.com/teamkeel/keel/compare/v0.344.0...v0.345.0) (2023-07-04)


### Features

* [@sortable](https://github.com/sortable) request ordering ([#1030](https://github.com/teamkeel/keel/issues/1030)) ([2424394](https://github.com/teamkeel/keel/commit/24243940398ac06c9cebe311ae52a42c5a8f59b8))

# [0.344.0](https://github.com/teamkeel/keel/compare/v0.343.0...v0.344.0) (2023-07-03)


### Features

* [@sortable](https://github.com/sortable) parsing and schema validation ([#1029](https://github.com/teamkeel/keel/issues/1029)) ([6b90d5a](https://github.com/teamkeel/keel/commit/6b90d5a98447eb6c2db99239d3e6c78ffa90c945))

# [0.343.0](https://github.com/teamkeel/keel/compare/v0.342.0...v0.343.0) (2023-07-02)


### Features

* [@order](https://github.com/order)By sorting & forward pagination ([#1028](https://github.com/teamkeel/keel/issues/1028)) ([c87a4d7](https://github.com/teamkeel/keel/commit/c87a4d7c6e1f4e9289af8be720d8d4918bf542ac))

# [0.342.0](https://github.com/teamkeel/keel/compare/v0.341.1...v0.342.0) (2023-06-30)


### Features

* orderBy attribute schema validation ([#1025](https://github.com/teamkeel/keel/issues/1025)) ([c9aac22](https://github.com/teamkeel/keel/commit/c9aac22d9d8373a262bb9e41d80c2f62347d850f))

## [0.341.1](https://github.com/teamkeel/keel/compare/v0.341.0...v0.341.1) (2023-06-29)


### Bug Fixes

* validate unique attribute restrictions ([#1022](https://github.com/teamkeel/keel/issues/1022)) ([ff43bb9](https://github.com/teamkeel/keel/commit/ff43bb98d07db994ac36777b6c8ac7e01dd11c3d))

# [0.341.0](https://github.com/teamkeel/keel/compare/v0.340.0...v0.341.0) (2023-06-29)


### Features

* nested input structure for list actions ([#1023](https://github.com/teamkeel/keel/issues/1023)) ([cccb9d8](https://github.com/teamkeel/keel/commit/cccb9d82eccaa11d702eef084c8b361872dbb645))

# [0.340.0](https://github.com/teamkeel/keel/compare/v0.339.1...v0.340.0) (2023-06-26)


### Features

* adds dynamically generated docs to model api instances ([#1020](https://github.com/teamkeel/keel/issues/1020)) ([c72efe0](https://github.com/teamkeel/keel/commit/c72efe07db61b83cd2efe1aa818b060e6d3cdd34))

## [0.339.1](https://github.com/teamkeel/keel/compare/v0.339.0...v0.339.1) (2023-06-23)


### Bug Fixes

* unique composite validation message ([#1018](https://github.com/teamkeel/keel/issues/1018)) ([b13f701](https://github.com/teamkeel/keel/commit/b13f7015a2473485c00c0a090fc768a197207d89))

# [0.339.0](https://github.com/teamkeel/keel/compare/v0.338.0...v0.339.0) (2023-06-20)


### Features

* expose kysely instance from sdk ([#1017](https://github.com/teamkeel/keel/issues/1017)) ([1f93db9](https://github.com/teamkeel/keel/commit/1f93db98008f156114ab90a72297cf3f42c68390))

# [0.338.0](https://github.com/teamkeel/keel/compare/v0.337.0...v0.338.0) (2023-06-19)


### Features

* support composite unique constraints in database migrations ([0f4a491](https://github.com/teamkeel/keel/commit/0f4a4912f46cea8efbaad60399f39922d3d000d6))

# [0.337.0](https://github.com/teamkeel/keel/compare/v0.336.0...v0.337.0) (2023-06-19)


### Features

* instrumented pg for better tracing in functions ([c867f8f](https://github.com/teamkeel/keel/commit/c867f8fc3ba4d9eb4a311e10f0f6e33f157cf74d))

# [0.336.0](https://github.com/teamkeel/keel/compare/v0.335.0...v0.336.0) (2023-06-16)


### Features

* change getDatabase to useDatabase ([#1011](https://github.com/teamkeel/keel/issues/1011)) ([039e8bc](https://github.com/teamkeel/keel/commit/039e8bcb128c56c0c1c0b939b6dfb12e927a61b8))

# [0.335.0](https://github.com/teamkeel/keel/compare/v0.334.0...v0.335.0) (2023-06-16)


### Features

* support offset / limit / orderBy in Model API ([#1009](https://github.com/teamkeel/keel/issues/1009)) ([faedd98](https://github.com/teamkeel/keel/commit/faedd981313a5f159b840d34f7680883c36070fd))

# [0.334.0](https://github.com/teamkeel/keel/compare/v0.333.0...v0.334.0) (2023-06-16)


### Features

* remove prisma ([5867670](https://github.com/teamkeel/keel/commit/5867670a37e35bfd05b6fb92b91d9f263a4ad4c8))

# [0.333.0](https://github.com/teamkeel/keel/compare/v0.332.0...v0.333.0) (2023-06-15)


### Features

* allow passing prisma binary target as an option to codegen ([598c540](https://github.com/teamkeel/keel/commit/598c5405420cdd07aba59df6f83b7b90d982e3a8))

# [0.332.0](https://github.com/teamkeel/keel/compare/v0.331.0...v0.332.0) (2023-06-15)


### Features

* list filtering by null ([#1000](https://github.com/teamkeel/keel/issues/1000)) ([93c4de4](https://github.com/teamkeel/keel/commit/93c4de4583f4b242eff3e2c6abd17d3439c229ae))

# [0.331.0](https://github.com/teamkeel/keel/compare/v0.330.3...v0.331.0) (2023-06-14)


### Features

* configure prisma tracing ([#1007](https://github.com/teamkeel/keel/issues/1007)) ([adae004](https://github.com/teamkeel/keel/commit/adae004170aaa85b1a6b0f9de3bf2b609cc88c9b))

## [0.330.3](https://github.com/teamkeel/keel/compare/v0.330.2...v0.330.3) (2023-06-14)


### Bug Fixes

* specify binaryTargets in prisma schema ([321feb7](https://github.com/teamkeel/keel/commit/321feb72576e13688fed0c5be0eeaef6a9ba801c))

## [0.330.2](https://github.com/teamkeel/keel/compare/v0.330.1...v0.330.2) (2023-06-13)


### Bug Fixes

* empty inputs for arbitrary functions ([#1003](https://github.com/teamkeel/keel/issues/1003)) ([6d05cf0](https://github.com/teamkeel/keel/commit/6d05cf086c90b07556bbf4445c9deb2b617fa926))

## [0.330.1](https://github.com/teamkeel/keel/compare/v0.330.0...v0.330.1) (2023-06-13)


### Bug Fixes

* upgrade bubbletea to fix issue with how tty is handled ([8f57ad9](https://github.com/teamkeel/keel/commit/8f57ad9940be404121bb30a73dbd8d692a2eeecd))

# [0.330.0](https://github.com/teamkeel/keel/compare/v0.329.0...v0.330.0) (2023-06-13)


### Features

* support multiple schema files in wasm  API ([9400463](https://github.com/teamkeel/keel/commit/9400463cb5552d207031abfefa096e2006a50c07))

# [0.329.0](https://github.com/teamkeel/keel/compare/v0.328.0...v0.329.0) (2023-06-13)


### Features

* nullable inputs ([#998](https://github.com/teamkeel/keel/issues/998)) ([e2a2c91](https://github.com/teamkeel/keel/commit/e2a2c91da66f3f8024cef9a711cea52987c62a2f))

# [0.328.0](https://github.com/teamkeel/keel/compare/v0.327.1...v0.328.0) (2023-06-12)


### Features

* generate prisma schema and upgrade user dependencies ([#996](https://github.com/teamkeel/keel/issues/996)) ([0ab5d37](https://github.com/teamkeel/keel/commit/0ab5d37aef3398d373cc57222ceee94345bedb68))

## [0.327.1](https://github.com/teamkeel/keel/compare/v0.327.0...v0.327.1) (2023-06-12)


### Bug Fixes

* move cors to only be used in localhost ([#997](https://github.com/teamkeel/keel/issues/997)) ([5426034](https://github.com/teamkeel/keel/commit/54260344e9a8a67090afa34aab8952ad52b0e3e8))

# [0.327.0](https://github.com/teamkeel/keel/compare/v0.326.1...v0.327.0) (2023-06-11)


### Features

* create a new database instance from a conn string ([#993](https://github.com/teamkeel/keel/issues/993)) ([e0d2326](https://github.com/teamkeel/keel/commit/e0d2326f401d6174c6019e978515422afae01820))

## [0.326.1](https://github.com/teamkeel/keel/compare/v0.326.0...v0.326.1) (2023-06-01)


### Bug Fixes

* using our null-supported graphql-go fork ([#991](https://github.com/teamkeel/keel/issues/991)) ([681c902](https://github.com/teamkeel/keel/commit/681c902faf49be3247abb49341f1ed33fcd27db5))

# [0.326.0](https://github.com/teamkeel/keel/compare/v0.325.0...v0.326.0) (2023-05-30)


### Features

* prisma client creation ([#988](https://github.com/teamkeel/keel/issues/988)) ([e3a4e5a](https://github.com/teamkeel/keel/commit/e3a4e5a2be77e7dd3f109b1672082439abe73a14))

# [0.325.0](https://github.com/teamkeel/keel/compare/v0.324.1...v0.325.0) (2023-05-25)


### Features

* init cmd as standalone tea program ([#986](https://github.com/teamkeel/keel/issues/986)) ([9c837fa](https://github.com/teamkeel/keel/commit/9c837fa1d32ca45a9c17f0e431ea88b4020bfb56))

## [0.324.1](https://github.com/teamkeel/keel/compare/v0.324.0...v0.324.1) (2023-05-25)


### Bug Fixes

* minor generate fixes ([#985](https://github.com/teamkeel/keel/issues/985)) ([af5413b](https://github.com/teamkeel/keel/commit/af5413b06540fd4fe0875d9ffbebeaf4a3e7bd72))

# [0.324.0](https://github.com/teamkeel/keel/compare/v0.323.4...v0.324.0) (2023-05-24)


### Features

* generate cmd ([#981](https://github.com/teamkeel/keel/issues/981)) ([182ef66](https://github.com/teamkeel/keel/commit/182ef66b07288c240070eecab1b45e07bd6e6c9f))

## [0.323.4](https://github.com/teamkeel/keel/compare/v0.323.3...v0.323.4) (2023-05-22)


### Bug Fixes

* model and message cannot have required fields of same type ([#978](https://github.com/teamkeel/keel/issues/978)) ([5a32966](https://github.com/teamkeel/keel/commit/5a32966c587bec1f4cc9e0cd71e62115dbb00627))

## [0.323.3](https://github.com/teamkeel/keel/compare/v0.323.2...v0.323.3) (2023-05-19)


### Bug Fixes

* init cmd minor fixes ([#979](https://github.com/teamkeel/keel/issues/979)) ([ef1bd78](https://github.com/teamkeel/keel/commit/ef1bd782b044a73d8577b8a581b3f509b4d3a54e))

## [0.323.2](https://github.com/teamkeel/keel/compare/v0.323.1...v0.323.2) (2023-05-18)


### Bug Fixes

* validate against bare model as input ([#977](https://github.com/teamkeel/keel/issues/977)) ([09c869f](https://github.com/teamkeel/keel/commit/09c869f505b1a5ccec264d96bd7ade152125e0b8))

## [0.323.1](https://github.com/teamkeel/keel/compare/v0.323.0...v0.323.1) (2023-05-18)


### Bug Fixes

* validate against direct many to many ([#976](https://github.com/teamkeel/keel/issues/976)) ([6748ada](https://github.com/teamkeel/keel/commit/6748adad46fcb2a0c101c808d7140baa5daee508))

# [0.323.0](https://github.com/teamkeel/keel/compare/v0.322.3...v0.323.0) (2023-05-18)


### Features

* init cmd ([#973](https://github.com/teamkeel/keel/issues/973)) ([4c2f929](https://github.com/teamkeel/keel/commit/4c2f9295469117a1ded857734a68a5af83eb77d7))

## [0.322.3](https://github.com/teamkeel/keel/compare/v0.322.2...v0.322.3) (2023-05-18)


### Bug Fixes

* generated database names ([#974](https://github.com/teamkeel/keel/issues/974)) ([2ffed0f](https://github.com/teamkeel/keel/commit/2ffed0fdb6cdeb94413cd5c0a7110c5a131c2c62))

## [0.322.2](https://github.com/teamkeel/keel/compare/v0.322.1...v0.322.2) (2023-05-17)


### Bug Fixes

* test lhs as well as rhs for repeated fields. #BLD-541 ([#969](https://github.com/teamkeel/keel/issues/969)) ([12fee30](https://github.com/teamkeel/keel/commit/12fee30c47ba65b9322c5f3fc7224f5d75a31c97)), closes [#BLD-541](https://github.com/teamkeel/keel/issues/BLD-541)

## [0.322.1](https://github.com/teamkeel/keel/compare/v0.322.0...v0.322.1) (2023-05-17)


### Bug Fixes

* use version number from ldflags for npm packages ([#975](https://github.com/teamkeel/keel/issues/975)) ([d70e632](https://github.com/teamkeel/keel/commit/d70e632282cee21a2144dc85f83f25e39ac30ae1))

# [0.322.0](https://github.com/teamkeel/keel/compare/v0.321.2...v0.322.0) (2023-05-17)


### Features

* add headers to ctx expressions auto complete ([#972](https://github.com/teamkeel/keel/issues/972)) ([43f3aea](https://github.com/teamkeel/keel/commit/43f3aea5cbc2568a5b29574f540f737a1d81084e))

## [0.321.2](https://github.com/teamkeel/keel/compare/v0.321.1...v0.321.2) (2023-05-17)


### Bug Fixes

* functions input argument order ([#967](https://github.com/teamkeel/keel/issues/967)) ([ba1db01](https://github.com/teamkeel/keel/commit/ba1db0195880b0dc50674b07de8db73552fef968))

## [0.321.1](https://github.com/teamkeel/keel/compare/v0.321.0...v0.321.1) (2023-05-15)


### Bug Fixes

* write generated files in scaffold ([e512b36](https://github.com/teamkeel/keel/commit/e512b369fe3aca29f16e58a13f97ce07a42d9937))

# [0.321.0](https://github.com/teamkeel/keel/compare/v0.320.0...v0.321.0) (2023-05-15)


### Features

* permissions for GQL nested queries ([#959](https://github.com/teamkeel/keel/issues/959)) ([a4ff78c](https://github.com/teamkeel/keel/commit/a4ff78cff420f169df6bd300c42a66a3fc25173f))

# [0.320.0](https://github.com/teamkeel/keel/compare/v0.319.2...v0.320.0) (2023-05-13)


### Features

* change API of node.Generate and GeneratedFiles.Write ([20c4a1a](https://github.com/teamkeel/keel/commit/20c4a1a98e9ece5387426cead604bad86451d362))

## [0.319.2](https://github.com/teamkeel/keel/compare/v0.319.1...v0.319.2) (2023-05-11)


### Bug Fixes

* more tests for enum/message clashes ([#965](https://github.com/teamkeel/keel/issues/965)) ([5d666ad](https://github.com/teamkeel/keel/commit/5d666ad95b86db6a5539fdea0348fd221b40329c))

## [0.319.1](https://github.com/teamkeel/keel/compare/v0.319.0...v0.319.1) (2023-05-11)


### Bug Fixes

* return Response from patched fetch in tracing ([589e824](https://github.com/teamkeel/keel/commit/589e82444f299b05f0b2e2c7c7583ac635566519))

# [0.319.0](https://github.com/teamkeel/keel/compare/v0.318.2...v0.319.0) (2023-05-11)


### Features

* pg_stat_statements extension ([#899](https://github.com/teamkeel/keel/issues/899)) ([a77e509](https://github.com/teamkeel/keel/commit/a77e509b4ee00464f954afec6548d00c8430ea2b))

## [0.318.2](https://github.com/teamkeel/keel/compare/v0.318.1...v0.318.2) (2023-05-11)


### Bug Fixes

* validate name clashes  ([#962](https://github.com/teamkeel/keel/issues/962)) ([77cdf05](https://github.com/teamkeel/keel/commit/77cdf05802e6b680841183dfae24335bead6b6e1))

## [0.318.1](https://github.com/teamkeel/keel/compare/v0.318.0...v0.318.1) (2023-05-10)


### Bug Fixes

* prettier type names for sdk, graphql etc ([#950](https://github.com/teamkeel/keel/issues/950)) ([b6478a1](https://github.com/teamkeel/keel/commit/b6478a14ef8717a9575d9e5748dfd2ce28e5cca4))

# [0.318.0](https://github.com/teamkeel/keel/compare/v0.317.0...v0.318.0) (2023-05-10)


### Features

* instrument fetch with tracing ([4c3dfa0](https://github.com/teamkeel/keel/commit/4c3dfa082101e377b859f57a208e152c0fa843b2))

# [0.317.0](https://github.com/teamkeel/keel/compare/v0.316.0...v0.317.0) (2023-05-05)


### Features

* drop api.fetch ([3a1fb06](https://github.com/teamkeel/keel/commit/3a1fb062ff4c5e22b49f3e9a0f926ee021b43b2f))

# [0.316.0](https://github.com/teamkeel/keel/compare/v0.315.0...v0.316.0) (2023-05-05)


### Features

* fix withSpan tracer ([2c1dc1d](https://github.com/teamkeel/keel/commit/2c1dc1df6fa0d7b62fad5846fb179c156107373f))

# [0.315.0](https://github.com/teamkeel/keel/compare/v0.314.1...v0.315.0) (2023-05-04)


### Features

* add tracing to handleRequest ([2f454f2](https://github.com/teamkeel/keel/commit/2f454f264a030c8df201c40d6320f67e61780177))
* change trace name ([ecdee8c](https://github.com/teamkeel/keel/commit/ecdee8c03fdd92eba071e4a0219633f9c6949ac3))

## [0.314.1](https://github.com/teamkeel/keel/compare/v0.314.0...v0.314.1) (2023-05-04)


### Bug Fixes

* use .Type.Value ([#949](https://github.com/teamkeel/keel/issues/949)) ([fe71b38](https://github.com/teamkeel/keel/commit/fe71b38e38e00acb032aac233c8ca7537f7a8516))

# [0.314.0](https://github.com/teamkeel/keel/compare/v0.313.7...v0.314.0) (2023-05-04)


### Features

* add tracing to functions-runtime ([cced24a](https://github.com/teamkeel/keel/commit/cced24ad07556962ac89705475874fa81bc0c5db))
* address PR feedback ([0fe1271](https://github.com/teamkeel/keel/commit/0fe12717dcc02bac6f532ba57b2c65fcbc608fa5))

## [0.313.7](https://github.com/teamkeel/keel/compare/v0.313.6...v0.313.7) (2023-05-03)


### Bug Fixes

* disallow inputs in permission rules ([#946](https://github.com/teamkeel/keel/issues/946)) ([3c0dbee](https://github.com/teamkeel/keel/commit/3c0dbee32cf3763416cafa26482f56a0fcb29009))

## [0.313.6](https://github.com/teamkeel/keel/compare/v0.313.5...v0.313.6) (2023-05-03)


### Bug Fixes

* modify parser field type to attach node info ([#944](https://github.com/teamkeel/keel/issues/944)) ([8e39297](https://github.com/teamkeel/keel/commit/8e39297b5d80988e90a27496961bd2b35a56ba81))

## [0.313.5](https://github.com/teamkeel/keel/compare/v0.313.4...v0.313.5) (2023-05-02)


### Bug Fixes

* validate against repeated scalar fields defined on models ([#941](https://github.com/teamkeel/keel/issues/941)) ([81c686e](https://github.com/teamkeel/keel/commit/81c686eadaacd127cb16c708f4d182f6d9cb7722))

## [0.313.4](https://github.com/teamkeel/keel/compare/v0.313.3...v0.313.4) (2023-05-02)


### Bug Fixes

* create casing utility functions ([#939](https://github.com/teamkeel/keel/issues/939)) ([6dc4724](https://github.com/teamkeel/keel/commit/6dc47244de3c7ccc1fb17d6468949e2064979981))

## [0.313.3](https://github.com/teamkeel/keel/compare/v0.313.2...v0.313.3) (2023-04-28)


### Bug Fixes

* validation rules for identity fields in create operations ([938ad17](https://github.com/teamkeel/keel/commit/938ad176cce99c6911848cade5f39f7b3905d1aa))

## [0.313.2](https://github.com/teamkeel/keel/compare/v0.313.1...v0.313.2) (2023-04-26)


### Bug Fixes

* fix multline function logs from being garbled ([#938](https://github.com/teamkeel/keel/issues/938)) ([333bc0c](https://github.com/teamkeel/keel/commit/333bc0c5cff6ecb2b9eb7236dafb31351a760722))

## [0.313.1](https://github.com/teamkeel/keel/compare/v0.313.0...v0.313.1) (2023-04-26)


### Bug Fixes

* Ensure correct use of with in autocompletions. ([#929](https://github.com/teamkeel/keel/issues/929)) ([adf7344](https://github.com/teamkeel/keel/commit/adf734448a09de0982fd60f2a66dac848b8ccf85))

# [0.313.0](https://github.com/teamkeel/keel/compare/v0.312.3...v0.313.0) (2023-04-26)


### Features

* private key run argument ([#935](https://github.com/teamkeel/keel/issues/935)) ([bbebd2e](https://github.com/teamkeel/keel/commit/bbebd2e721a6565b87125f3cbcbc255a75bc54cf))

## [0.312.3](https://github.com/teamkeel/keel/compare/v0.312.2...v0.312.3) (2023-04-26)


### Bug Fixes

* support notEquals operator ([#934](https://github.com/teamkeel/keel/issues/934)) ([d6e60e2](https://github.com/teamkeel/keel/commit/d6e60e28cbde4aebb91c316107edbf433f130df1))

## [0.312.2](https://github.com/teamkeel/keel/compare/v0.312.1...v0.312.2) (2023-04-26)


### Bug Fixes

* hasNextPage bug ([#936](https://github.com/teamkeel/keel/issues/936)) ([5090415](https://github.com/teamkeel/keel/commit/509041596d10bdce8f56f0c7bdf294986685077e))

## [0.312.1](https://github.com/teamkeel/keel/compare/v0.312.0...v0.312.1) (2023-04-26)


### Bug Fixes

* oneOf operator for Number type ([#933](https://github.com/teamkeel/keel/issues/933)) ([6493060](https://github.com/teamkeel/keel/commit/6493060fe14e6fadbdfbe4e01b9252117fbea033))
* use MaxWidth and MaxHeight to constrain content based on term dimensions ([#932](https://github.com/teamkeel/keel/issues/932)) ([a100b88](https://github.com/teamkeel/keel/commit/a100b88066707167107f4d9e08fd4415b4047c71))
* use never for result of permissions.deny() ([#931](https://github.com/teamkeel/keel/issues/931)) ([4763447](https://github.com/teamkeel/keel/commit/47634479baf952e1a10f80a49424431c55f642ec))

# [0.312.0](https://github.com/teamkeel/keel/compare/v0.311.0...v0.312.0) (2023-04-25)


### Features

* special treatment for Identity in nested create (validation) ([#930](https://github.com/teamkeel/keel/issues/930)) ([e84cc79](https://github.com/teamkeel/keel/commit/e84cc79a7812931455597352b741af480009e5af))

# [0.311.0](https://github.com/teamkeel/keel/compare/v0.310.0...v0.311.0) (2023-04-24)


### Features

* supporting third party tokens ([#928](https://github.com/teamkeel/keel/issues/928)) ([52617d5](https://github.com/teamkeel/keel/commit/52617d551eb010d2bf01a9d596dea1ffa9c10a77))

# [0.310.0](https://github.com/teamkeel/keel/compare/v0.309.0...v0.310.0) (2023-04-21)


### Features

* restrict use of named action inputs ([#927](https://github.com/teamkeel/keel/issues/927)) ([d08edfa](https://github.com/teamkeel/keel/commit/d08edfa36dac5b6be2a7b32dfa003e280b06352c))

# [0.309.0](https://github.com/teamkeel/keel/compare/v0.308.0...v0.309.0) (2023-04-20)


### Features

* revert ([#926](https://github.com/teamkeel/keel/issues/926)) ([299d58d](https://github.com/teamkeel/keel/commit/299d58d85f8ca7432a29cb12d6dd934955fd9aa6))

# [0.308.0](https://github.com/teamkeel/keel/compare/v0.307.1...v0.308.0) (2023-04-20)


### Features

* remove src ([#925](https://github.com/teamkeel/keel/issues/925)) ([713d502](https://github.com/teamkeel/keel/commit/713d502aa3d795c1b0949ee56237a8c406c110b3))

## [0.307.1](https://github.com/teamkeel/keel/compare/v0.307.0...v0.307.1) (2023-04-19)


### Bug Fixes

* fix codegen nil pointer ([#924](https://github.com/teamkeel/keel/issues/924)) ([811beba](https://github.com/teamkeel/keel/commit/811bebaa664f1b3c09d2b34a9d35b78bebc050b5))

# [0.307.0](https://github.com/teamkeel/keel/compare/v0.306.0...v0.307.0) (2023-04-19)


### Features

* cli fix ([#923](https://github.com/teamkeel/keel/issues/923)) ([feddbca](https://github.com/teamkeel/keel/commit/feddbca4d986722b4889d9be1e0b02e307e572e5))

# [0.306.0](https://github.com/teamkeel/keel/compare/v0.305.0...v0.306.0) (2023-04-19)


### Features

* homebrew tap changes ([#922](https://github.com/teamkeel/keel/issues/922)) ([5358d00](https://github.com/teamkeel/keel/commit/5358d0088be348ebccb5d517f7c0a9dcc9e659cb))

# [0.305.0](https://github.com/teamkeel/keel/compare/v0.304.0...v0.305.0) (2023-04-19)


### Features

* change path ([#921](https://github.com/teamkeel/keel/issues/921)) ([95d710b](https://github.com/teamkeel/keel/commit/95d710b192eda3eb1e19f2f50d117461672b7f54))

# [0.304.0](https://github.com/teamkeel/keel/compare/v0.303.0...v0.304.0) (2023-04-19)


### Features

* custom gh download strategy ([#920](https://github.com/teamkeel/keel/issues/920)) ([1a1c078](https://github.com/teamkeel/keel/commit/1a1c0784bc8e85c9dc1fccca2b89aa9550381a13))

# [0.303.0](https://github.com/teamkeel/keel/compare/v0.302.0...v0.303.0) (2023-04-18)


### Features

* use gh_token ([#919](https://github.com/teamkeel/keel/issues/919)) ([35d6082](https://github.com/teamkeel/keel/commit/35d60826beaac3b7b9b32131794300f377a97988))

# [0.302.0](https://github.com/teamkeel/keel/compare/v0.301.0...v0.302.0) (2023-04-18)


### Features

* homebrew tap ([#918](https://github.com/teamkeel/keel/issues/918)) ([f22ee2e](https://github.com/teamkeel/keel/commit/f22ee2e584f3cf0d39b51392861051724f13b482))

# [0.301.0](https://github.com/teamkeel/keel/compare/v0.300.3...v0.301.0) (2023-04-17)


### Features

* remove [] from ActionInputs in parser ([#915](https://github.com/teamkeel/keel/issues/915)) ([2cf0e1b](https://github.com/teamkeel/keel/commit/2cf0e1bd558df270f774bb5d53989633115b1a2f))

## [0.300.3](https://github.com/teamkeel/keel/compare/v0.300.2...v0.300.3) (2023-04-14)


### Bug Fixes

* run permissions checks for delete/update ([#913](https://github.com/teamkeel/keel/issues/913)) ([69840cb](https://github.com/teamkeel/keel/commit/69840cb6b41139c4bd8b1e649dcff2453623e96b))

## [0.300.2](https://github.com/teamkeel/keel/compare/v0.300.1...v0.300.2) (2023-04-13)


### Bug Fixes

* wrap terminal text ([#912](https://github.com/teamkeel/keel/issues/912)) ([c91bb06](https://github.com/teamkeel/keel/commit/c91bb06e77370417f555e793ffa1a2cb6f1556be))

## [0.300.1](https://github.com/teamkeel/keel/compare/v0.300.0...v0.300.1) (2023-04-13)


### Bug Fixes

* show errors emitted from node process failing to start ([#911](https://github.com/teamkeel/keel/issues/911)) ([f993d4f](https://github.com/teamkeel/keel/commit/f993d4f58682edef2bccb8125cc40d48a8026c8f))

# [0.300.0](https://github.com/teamkeel/keel/compare/v0.299.0...v0.300.0) (2023-04-13)


### Features

* role based permissions in custom functions ([#906](https://github.com/teamkeel/keel/issues/906)) ([0391ff5](https://github.com/teamkeel/keel/commit/0391ff559fecb84ceac563e1ec2709f3ffdf35f4))

# [0.299.0](https://github.com/teamkeel/keel/compare/v0.298.0...v0.299.0) (2023-04-12)


### Features

* schema validation new nested create ambiguity rule ([#907](https://github.com/teamkeel/keel/issues/907)) ([5fb4bbb](https://github.com/teamkeel/keel/commit/5fb4bbb76be0f1171423809a0e50c3c26746666c))

# [0.298.0](https://github.com/teamkeel/keel/compare/v0.297.1...v0.298.0) (2023-04-12)


### Features

* requestPasswordReset and passwordReset actions ([#901](https://github.com/teamkeel/keel/issues/901)) ([04ea252](https://github.com/teamkeel/keel/commit/04ea252910cf2ca12472707dc4604b99da51f2c3))

## [0.297.1](https://github.com/teamkeel/keel/compare/v0.297.0...v0.297.1) (2023-04-12)


### Bug Fixes

* missing defer on span ending in runtime ([ee4a106](https://github.com/teamkeel/keel/commit/ee4a106f76e1c45e38d3ade79e274a6f181c53d3))

# [0.297.0](https://github.com/teamkeel/keel/compare/v0.296.2...v0.297.0) (2023-04-05)


### Features

* [@set](https://github.com/set) in nested related data on create op ([#896](https://github.com/teamkeel/keel/issues/896)) ([b43a7df](https://github.com/teamkeel/keel/commit/b43a7dfa43c116f3d447fb64670ffd0bfebaf069))

## [0.296.2](https://github.com/teamkeel/keel/compare/v0.296.1...v0.296.2) (2023-04-04)


### Bug Fixes

* message field autoformatting ([#890](https://github.com/teamkeel/keel/issues/890)) ([11fdbf0](https://github.com/teamkeel/keel/commit/11fdbf03ff5eeb638131935ae8ed6ea4021f90a5))

## [0.296.1](https://github.com/teamkeel/keel/compare/v0.296.0...v0.296.1) (2023-04-04)


### Bug Fixes

* runtime version key ([f098670](https://github.com/teamkeel/keel/commit/f098670e6732c4792cdd7253da631e949927a973))

# [0.296.0](https://github.com/teamkeel/keel/compare/v0.295.2...v0.296.0) (2023-04-03)


### Features

* add version to traces ([b1de1ec](https://github.com/teamkeel/keel/commit/b1de1ec8816bac5e668d91fa18e570047cb77bab))

## [0.295.2](https://github.com/teamkeel/keel/compare/v0.295.1...v0.295.2) (2023-04-03)


### Bug Fixes

* return 400 for json validation errors ([33c6e8d](https://github.com/teamkeel/keel/commit/33c6e8d7beafdbdb48589d78f349005696b951e5))

## [0.295.1](https://github.com/teamkeel/keel/compare/v0.295.0...v0.295.1) (2023-04-03)


### Bug Fixes

* omitting optional model association ([#894](https://github.com/teamkeel/keel/issues/894)) ([0a55877](https://github.com/teamkeel/keel/commit/0a558777cb0754382a3a1adc942bcdcb95bde5fc))

# [0.295.0](https://github.com/teamkeel/keel/compare/v0.294.3...v0.295.0) (2023-03-31)


### Features

* upgrade validation of Create ops for nested creation ([#888](https://github.com/teamkeel/keel/issues/888)) ([3a7ef1c](https://github.com/teamkeel/keel/commit/3a7ef1c4cc24bbf054638e464739128f0ac091bf))

## [0.294.3](https://github.com/teamkeel/keel/compare/v0.294.2...v0.294.3) (2023-03-31)


### Bug Fixes

* action level permission rules take precedence ([#889](https://github.com/teamkeel/keel/issues/889)) ([a4b07a5](https://github.com/teamkeel/keel/commit/a4b07a55bc50c0e792ca0655a5489185e3b70f72))

## [0.294.2](https://github.com/teamkeel/keel/compare/v0.294.1...v0.294.2) (2023-03-31)


### Bug Fixes

* fixing assumption in runtime test ([#892](https://github.com/teamkeel/keel/issues/892)) ([ad20f80](https://github.com/teamkeel/keel/commit/ad20f80430dcfada604f71ec57aa52707f266500))

## [0.294.1](https://github.com/teamkeel/keel/compare/v0.294.0...v0.294.1) (2023-03-31)


### Bug Fixes

* rpc and json api support for creating relationships  ([#891](https://github.com/teamkeel/keel/issues/891)) ([1f02d92](https://github.com/teamkeel/keel/commit/1f02d92712ea3a8d6cd008c00519c55d0b348c5c))

# [0.294.0](https://github.com/teamkeel/keel/compare/v0.293.4...v0.294.0) (2023-03-31)


### Features

* creating relationships ([#882](https://github.com/teamkeel/keel/issues/882)) ([f494e00](https://github.com/teamkeel/keel/commit/f494e00d1287554e6f5fa932be33aa3211a0b1ae))

## [0.293.4](https://github.com/teamkeel/keel/compare/v0.293.3...v0.293.4) (2023-03-30)


### Bug Fixes

* add ctx.isAuthenticated to completions ([#877](https://github.com/teamkeel/keel/issues/877)) ([9b74915](https://github.com/teamkeel/keel/commit/9b74915108f5f5cc244db6fbfc19db2e23b89c77))

## [0.293.3](https://github.com/teamkeel/keel/compare/v0.293.2...v0.293.3) (2023-03-30)


### Bug Fixes

* validate 'with' keyword usage ([#887](https://github.com/teamkeel/keel/issues/887)) ([338a2bb](https://github.com/teamkeel/keel/commit/338a2bb3853f3cdc750d0a3156499f210901c3aa))

## [0.293.2](https://github.com/teamkeel/keel/compare/v0.293.1...v0.293.2) (2023-03-30)


### Bug Fixes

* guard against empty list inputs ([#886](https://github.com/teamkeel/keel/issues/886)) ([672d545](https://github.com/teamkeel/keel/commit/672d545703c7eb79aae21293ce8cb80ccd41c3a4))

## [0.293.1](https://github.com/teamkeel/keel/compare/v0.293.0...v0.293.1) (2023-03-29)


### Bug Fixes

* validate arbitrary function return types ([#881](https://github.com/teamkeel/keel/issues/881)) ([04548b3](https://github.com/teamkeel/keel/commit/04548b3f9cdb7fde0b4ee70a7c0f0f800d1db97b))

# [0.293.0](https://github.com/teamkeel/keel/compare/v0.292.0...v0.293.0) (2023-03-29)


### Features

* scaffold command ([#883](https://github.com/teamkeel/keel/issues/883)) ([a270a7d](https://github.com/teamkeel/keel/commit/a270a7da8a67070cea13b3dab82139a85464a1b0))

# [0.292.0](https://github.com/teamkeel/keel/compare/v0.291.0...v0.292.0) (2023-03-29)


### Features

* support totalCount ([#876](https://github.com/teamkeel/keel/issues/876)) ([331e4ab](https://github.com/teamkeel/keel/commit/331e4ab484f25200d6801a31958366a92cc85354))

# [0.291.0](https://github.com/teamkeel/keel/compare/v0.290.1...v0.291.0) (2023-03-29)


### Bug Fixes

* change makefile to use nix shell on some commands ([c7d16cc](https://github.com/teamkeel/keel/commit/c7d16cc9c46948f57438b4ed2454a7b1ae29f456))
* check code generation on ci ([3e1f122](https://github.com/teamkeel/keel/commit/3e1f122ac33ebf3dc85ebddf5b1b173dba176e60))
* drop useless bit on yaml ([98b4224](https://github.com/teamkeel/keel/commit/98b42241c497237658afb2d968a15573e6d55382))
* pb schema version ([99e0a03](https://github.com/teamkeel/keel/commit/99e0a037c726648e3c526f8a05841eb0f84e32c7))
* try a different approach ([d3825f6](https://github.com/teamkeel/keel/commit/d3825f695114765ed1d56b2e7b1a76fbd67ac85f))
* update go.mod ([ebae9ea](https://github.com/teamkeel/keel/commit/ebae9ea60e01b2c97afd280b2c0a6737466e8f43))


### Features

* update go to 1.20 ([005ac83](https://github.com/teamkeel/keel/commit/005ac835557c20ccfe1167425764006c901c4c15))

## [0.290.1](https://github.com/teamkeel/keel/compare/v0.290.0...v0.290.1) (2023-03-29)


### Bug Fixes

* disable cgo ([0f24d19](https://github.com/teamkeel/keel/commit/0f24d19e31546b4eaeb6bca1e3482ba1f549d0cb))
* update Makefile ([b4a430a](https://github.com/teamkeel/keel/commit/b4a430a466c670af76be481166240acd77f49bbc))

# [0.290.0](https://github.com/teamkeel/keel/compare/v0.289.2...v0.290.0) (2023-03-28)


### Features

* built-in permissions in custom functions ([#861](https://github.com/teamkeel/keel/issues/861)) ([9387a38](https://github.com/teamkeel/keel/commit/9387a38d2b55d354963fceb4590971f8c8e9d741))

## [0.289.2](https://github.com/teamkeel/keel/compare/v0.289.1...v0.289.2) (2023-03-27)


### Bug Fixes

* add seconds back to timestamp gql object (#BLD-344) ([f52ead7](https://github.com/teamkeel/keel/commit/f52ead7407a7bf238401ac02382637cf59df0180)), closes [#BLD-344](https://github.com/teamkeel/keel/issues/BLD-344)
* update tests to check for seconds ([8ce9ba5](https://github.com/teamkeel/keel/commit/8ce9ba55fa95a587e909afb709ca6e928b94ea50))
* update tests to include seconds field in timestamp object (#BLD-344) ([7ea8139](https://github.com/teamkeel/keel/commit/7ea81390dc8d25ad2a31211a633cd9be5903474c)), closes [#BLD-344](https://github.com/teamkeel/keel/issues/BLD-344)

## [0.289.1](https://github.com/teamkeel/keel/compare/v0.289.0...v0.289.1) (2023-03-27)


### Bug Fixes

* graphql support model message field type ([#873](https://github.com/teamkeel/keel/issues/873)) ([99c7010](https://github.com/teamkeel/keel/commit/99c701080a72d78d6d3e9f5619325e0b7f8a6bd5))

# [0.289.0](https://github.com/teamkeel/keel/compare/v0.288.2...v0.289.0) (2023-03-25)


### Features

* temporarily use tea.Println to expose logs ([#867](https://github.com/teamkeel/keel/issues/867)) ([6fa521f](https://github.com/teamkeel/keel/commit/6fa521f00f94d31e6f9b6fb1b233d7eb774872f0))

## [0.288.2](https://github.com/teamkeel/keel/compare/v0.288.1...v0.288.2) (2023-03-24)


### Bug Fixes

* dont format ID type as Id ([#870](https://github.com/teamkeel/keel/issues/870)) ([f2f907f](https://github.com/teamkeel/keel/commit/f2f907f97f17f19cf0b14e7380f2ce0d60f10291))

## [0.288.1](https://github.com/teamkeel/keel/compare/v0.288.0...v0.288.1) (2023-03-23)


### Bug Fixes

* use sub for JWT id ([59dbc8e](https://github.com/teamkeel/keel/commit/59dbc8e3f946347931781b4f06ff0feacbc73687))

# [0.288.0](https://github.com/teamkeel/keel/compare/v0.287.0...v0.288.0) (2023-03-22)


### Features

* add operationId to openAPI definitions ([81745ac](https://github.com/teamkeel/keel/commit/81745ac811af74bda9b27af7dcb0a00d0675b367))

# [0.287.0](https://github.com/teamkeel/keel/compare/v0.286.0...v0.287.0) (2023-03-21)


### Features

* case insensitive urls ([86c4357](https://github.com/teamkeel/keel/commit/86c43570f754d87aeb6a6a1c363a13fb3c6030c2))

# [0.286.0](https://github.com/teamkeel/keel/compare/v0.285.0...v0.286.0) (2023-03-16)


### Features

* permissions sdk ([#860](https://github.com/teamkeel/keel/issues/860)) ([fcc7d21](https://github.com/teamkeel/keel/commit/fcc7d21ea0ba7225d32d8d83329f9a713339a201))

# [0.285.0](https://github.com/teamkeel/keel/compare/v0.284.1...v0.285.0) (2023-03-15)


### Features

* use timestamptz for storing dates/timestamps ([#858](https://github.com/teamkeel/keel/issues/858)) ([8e4ad22](https://github.com/teamkeel/keel/commit/8e4ad2212bd117c26ca149951b5b3ccf601bab85))

## [0.284.1](https://github.com/teamkeel/keel/compare/v0.284.0...v0.284.1) (2023-03-15)


### Bug Fixes

* handle bad iso8601 formats ([#857](https://github.com/teamkeel/keel/issues/857)) ([a3a0fd3](https://github.com/teamkeel/keel/commit/a3a0fd31f9f7ec89ad6cbfe00690ac47981f88b5))

# [0.284.0](https://github.com/teamkeel/keel/compare/v0.283.0...v0.284.0) (2023-03-15)


### Features

* automatic iso8601 parsing ([#854](https://github.com/teamkeel/keel/issues/854)) ([7e220f5](https://github.com/teamkeel/keel/commit/7e220f5f07f6fbb108aabab4ad382e3cc2a56110))

# [0.283.0](https://github.com/teamkeel/keel/compare/v0.282.0...v0.283.0) (2023-03-14)


### Features

* unify date/timestamp types in graphql with jsonrpc api ([#850](https://github.com/teamkeel/keel/issues/850)) ([6986f2f](https://github.com/teamkeel/keel/commit/6986f2f8932ebff83f78b143f362fc5cd48216fa))

# [0.282.0](https://github.com/teamkeel/keel/compare/v0.281.0...v0.282.0) (2023-03-14)


### Features

* support [@relation](https://github.com/relation) in code completions ([#853](https://github.com/teamkeel/keel/issues/853)) ([73a16df](https://github.com/teamkeel/keel/commit/73a16df683e6417190485c0cd3d19db8b923257d))

# [0.281.0](https://github.com/teamkeel/keel/compare/v0.280.0...v0.281.0) (2023-03-14)


### Features

* supporting multiple conditions, AND, OR, and parenthesis ([#847](https://github.com/teamkeel/keel/issues/847)) ([eb3d037](https://github.com/teamkeel/keel/commit/eb3d0373af76efcea4c8e5a2582ad3b8193546ce))

# [0.280.0](https://github.com/teamkeel/keel/compare/v0.279.0...v0.280.0) (2023-03-13)


### Features

* completion support for any inputs ([#848](https://github.com/teamkeel/keel/issues/848)) ([d2a4bcc](https://github.com/teamkeel/keel/commit/d2a4bcc12e6b0501098492715b1e5e13e8c3d036))

# [0.279.0](https://github.com/teamkeel/keel/compare/v0.278.0...v0.279.0) (2023-03-13)


### Features

* support 'Any' type  ([#840](https://github.com/teamkeel/keel/issues/840)) ([f00e56d](https://github.com/teamkeel/keel/commit/f00e56d3fca73a7a90978dfaa02452450b2c9c33))

# [0.278.0](https://github.com/teamkeel/keel/compare/v0.277.1...v0.278.0) (2023-03-02)


### Features

* hide foreign key field names from the user ([#842](https://github.com/teamkeel/keel/issues/842)) ([1e9a5d7](https://github.com/teamkeel/keel/commit/1e9a5d78f6afd23536baa2ca2256c37dcf639146))

## [0.277.1](https://github.com/teamkeel/keel/compare/v0.277.0...v0.277.1) (2023-03-01)


### Bug Fixes

* enums in messages ([#839](https://github.com/teamkeel/keel/issues/839)) ([a7cca01](https://github.com/teamkeel/keel/commit/a7cca016150db8af98b5776e02a06bd8bb825c57))

# [0.277.0](https://github.com/teamkeel/keel/compare/v0.276.0...v0.277.0) (2023-03-01)


### Features

* arrays and model support in messages ([#837](https://github.com/teamkeel/keel/issues/837)) ([77ca8a8](https://github.com/teamkeel/keel/commit/77ca8a86fad897654910eca38aca31e9d2937c59))

# [0.276.0](https://github.com/teamkeel/keel/compare/v0.275.0...v0.276.0) (2023-03-01)


### Features

* nested and reuseable messages ([#836](https://github.com/teamkeel/keel/issues/836)) ([18773b1](https://github.com/teamkeel/keel/commit/18773b19b7de46b7cddc3878ad3f9c0f34a3efea))

# [0.275.0](https://github.com/teamkeel/keel/compare/v0.274.1...v0.275.0) (2023-02-28)


### Features

* arbitrary functions in custom functions ([#834](https://github.com/teamkeel/keel/issues/834)) ([a1e37d8](https://github.com/teamkeel/keel/commit/a1e37d84e8cd246b2291f0d17301bdcbf0b1b5cf))

## [0.274.1](https://github.com/teamkeel/keel/compare/v0.274.0...v0.274.1) (2023-02-28)


### Bug Fixes

* proto schema handles arb funcs with inline inputs ([#833](https://github.com/teamkeel/keel/issues/833)) ([4d744a8](https://github.com/teamkeel/keel/commit/4d744a84e8d376591c2bc3153f7f47b6b30aa502))

# [0.274.0](https://github.com/teamkeel/keel/compare/v0.273.8...v0.274.0) (2023-02-27)


### Features

* provide fully working [@relation](https://github.com/relation) attribute ([#830](https://github.com/teamkeel/keel/issues/830)) ([eb45ae9](https://github.com/teamkeel/keel/commit/eb45ae99857dd21329492fd71eec5e292edadb28))

## [0.273.8](https://github.com/teamkeel/keel/compare/v0.273.7...v0.273.8) (2023-02-27)


### Bug Fixes

* moving list query inputs into proto ([#821](https://github.com/teamkeel/keel/issues/821)) ([cfe2565](https://github.com/teamkeel/keel/commit/cfe2565cbfd7a51fe3c882933425a14e5627c9aa))

## [0.273.7](https://github.com/teamkeel/keel/compare/v0.273.6...v0.273.7) (2023-02-23)


### Bug Fixes

* go_package option in proto file was wrong ([dead3d3](https://github.com/teamkeel/keel/commit/dead3d31f0a0bf0ba35ceafaa66fa3701d814c28))

## [0.273.6](https://github.com/teamkeel/keel/compare/v0.273.5...v0.273.6) (2023-02-21)


### Bug Fixes

* fix ignoring of tags in semantic release action ([1c2d0be](https://github.com/teamkeel/keel/commit/1c2d0bee6b9602c85d1313f6d8e151dfdd9afec9))

## [0.273.5](https://github.com/teamkeel/keel/compare/v0.273.4...v0.273.5) (2023-02-21)


### Bug Fixes

* run semantic-release on main (no tag) and goreleaser on tags ([2e99063](https://github.com/teamkeel/keel/commit/2e990632f1eb32310f6381f9073bd5ae27ca2f00))

## [0.273.4](https://github.com/teamkeel/keel/compare/v0.273.3...v0.273.4) (2023-02-21)


### Bug Fixes

* run goreleaser as part of semantic-release process ([c58d1fe](https://github.com/teamkeel/keel/commit/c58d1fe018c2bf28bee6ee5548ada35de4dc0dd2))

## [0.273.3](https://github.com/teamkeel/keel/compare/v0.273.2...v0.273.3) (2023-02-21)


### Bug Fixes

* run gorelease in same job as semantic release ([64331a6](https://github.com/teamkeel/keel/commit/64331a6318a084ad4bba122d5fe884e70cc4bf35))

## [0.273.2](https://github.com/teamkeel/keel/compare/v0.273.1...v0.273.2) (2023-02-21)


### Bug Fixes

* testing goreleaser ([82cee83](https://github.com/teamkeel/keel/commit/82cee8308a20fc4a5d7fa1b99e8abaa90aa36fce))

# [0.273.0](https://github.com/teamkeel/keel/compare/v0.272.0...v0.273.0) (2023-02-21)


### Features

* secrets set & remove commands ([#809](https://github.com/teamkeel/keel/issues/809)) ([0980dff](https://github.com/teamkeel/keel/commit/0980dffacccb34ff3692bb31217d4e0815e14ec5))

# [0.272.0](https://github.com/teamkeel/keel/compare/v0.271.0...v0.272.0) (2023-02-21)


### Features

* secrets list command ([#807](https://github.com/teamkeel/keel/issues/807)) ([2303702](https://github.com/teamkeel/keel/commit/2303702effd0dd7eccd51c0d8532638eebe19315))

# [0.271.0](https://github.com/teamkeel/keel/compare/v0.270.0...v0.271.0) (2023-02-21)


### Features

* remove short_message from validation error ([#803](https://github.com/teamkeel/keel/issues/803)) ([8ace621](https://github.com/teamkeel/keel/commit/8ace621538c3f1b530c68e0d77c3d6d596ff5aab))

# [0.270.0](https://github.com/teamkeel/keel/compare/v0.269.1...v0.270.0) (2023-02-21)


### Features

* Better support for colours in CLI ([#801](https://github.com/teamkeel/keel/issues/801)) ([1923ef4](https://github.com/teamkeel/keel/commit/1923ef481edbbde530b802252165ef72cb518c0f))

## [0.269.1](https://github.com/teamkeel/keel/compare/v0.269.0...v0.269.1) (2023-02-20)


### Bug Fixes

* reinstate ToAnnotatedSchema func on errorhandling.ValidationErrors ([042b682](https://github.com/teamkeel/keel/commit/042b682a7ce2125a03b7901f3096bb67ab7176ba))

# [0.269.0](https://github.com/teamkeel/keel/compare/v0.268.0...v0.269.0) (2023-02-20)


### Features

* add secrets to proto ([#788](https://github.com/teamkeel/keel/issues/788)) ([d12555a](https://github.com/teamkeel/keel/commit/d12555a41ffd41031e8eaf7ff88db03344575c9c))

# [0.268.0](https://github.com/teamkeel/keel/compare/v0.267.0...v0.268.0) (2023-02-17)


### Features

* format message types ([#797](https://github.com/teamkeel/keel/issues/797)) ([96eee92](https://github.com/teamkeel/keel/commit/96eee92cfe4069ef852f2d2658a974a0b5e4ffce))

# [0.267.0](https://github.com/teamkeel/keel/compare/v0.266.0...v0.267.0) (2023-02-17)


### Features

* basic autocomplete for message types ([#796](https://github.com/teamkeel/keel/issues/796)) ([949fa44](https://github.com/teamkeel/keel/commit/949fa4424757f04c7d8e5466ba5506f53e7be8c2))

# [0.266.0](https://github.com/teamkeel/keel/compare/v0.265.1...v0.266.0) (2023-02-17)


### Features

* codegen ctx.env declarations ([#794](https://github.com/teamkeel/keel/issues/794)) ([f3f58de](https://github.com/teamkeel/keel/commit/f3f58de20c07ab7dcd6eac4eeeeb5b25eb4df8fe))

## [0.265.1](https://github.com/teamkeel/keel/compare/v0.265.0...v0.265.1) (2023-02-17)


### Bug Fixes

* whitespace issue in testdata json ([15efee3](https://github.com/teamkeel/keel/commit/15efee3bf7f7e1bdb1dbfbd3c02bafe727864ef2))

# [0.265.0](https://github.com/teamkeel/keel/compare/v0.264.1...v0.265.0) (2023-02-17)


### Features

* add nix format check ([d56c318](https://github.com/teamkeel/keel/commit/d56c318a842bae27da35b307f6a12186e5f4dea7))

## [0.264.1](https://github.com/teamkeel/keel/compare/v0.264.0...v0.264.1) (2023-02-16)


### Bug Fixes

* message type for nested messages ([#781](https://github.com/teamkeel/keel/issues/781)) ([c8359b1](https://github.com/teamkeel/keel/commit/c8359b1429863857c976544b04521963791392c2))

# [0.264.0](https://github.com/teamkeel/keel/compare/v0.263.0...v0.264.0) (2023-02-16)


### Bug Fixes

* address review comments ([e92e272](https://github.com/teamkeel/keel/commit/e92e2729800103c19ac3f1e11f934f98cea3667e))


### Features

* add request and response logging ([a5b73f1](https://github.com/teamkeel/keel/commit/a5b73f17d7b03399a85e0e89e680de69990edc3e))

# [0.263.0](https://github.com/teamkeel/keel/compare/v0.262.0...v0.263.0) (2023-02-15)


### Features

* cli config ([#778](https://github.com/teamkeel/keel/issues/778)) ([1ffaa48](https://github.com/teamkeel/keel/commit/1ffaa48f092706e4303dec4e48e90f1f70180769))

# [0.262.0](https://github.com/teamkeel/keel/compare/v0.261.0...v0.262.0) (2023-02-14)


### Features

* message type in proto schema ([#777](https://github.com/teamkeel/keel/issues/777)) ([f038f97](https://github.com/teamkeel/keel/commit/f038f97682bcf74a73a94486b367ecab5bd1ab93))

# [0.261.0](https://github.com/teamkeel/keel/compare/v0.260.0...v0.261.0) (2023-02-13)


### Features

* revert [#774](https://github.com/teamkeel/keel/issues/774) ([#775](https://github.com/teamkeel/keel/issues/775)) ([4b77cb6](https://github.com/teamkeel/keel/commit/4b77cb6b73fab3d14e51d7d4473283f5ab4d7470))

# [0.260.0](https://github.com/teamkeel/keel/compare/v0.259.0...v0.260.0) (2023-02-13)


### Features

* compress binary with upx ([#774](https://github.com/teamkeel/keel/issues/774)) ([2150700](https://github.com/teamkeel/keel/commit/21507002173c902fce1f445be34c9e1f4543d496))

# [0.259.0](https://github.com/teamkeel/keel/compare/v0.258.0...v0.259.0) (2023-02-13)


### Features

* remove chmod ([#773](https://github.com/teamkeel/keel/issues/773)) ([68568bf](https://github.com/teamkeel/keel/commit/68568bf64975267ae4ec1a393c024492ae9b8e09))

# [0.258.0](https://github.com/teamkeel/keel/compare/v0.257.0...v0.258.0) (2023-02-13)


### Features

* chmod binary ([#772](https://github.com/teamkeel/keel/issues/772)) ([d3103d4](https://github.com/teamkeel/keel/commit/d3103d4e8c4d43065fce189bc93bfea1673b0b34))
* rename env vars using keel prefix ® ([#771](https://github.com/teamkeel/keel/issues/771)) ([f871ab1](https://github.com/teamkeel/keel/commit/f871ab147d75093770691a076c2b76affbe2440a))

# [0.257.0](https://github.com/teamkeel/keel/compare/v0.256.0...v0.257.0) (2023-02-13)


### Features

* compile cli for apple silicon ([#770](https://github.com/teamkeel/keel/issues/770)) ([8891b41](https://github.com/teamkeel/keel/commit/8891b411b8add123bd9400834fe8f89d8f9dbb01))

# [0.256.0](https://github.com/teamkeel/keel/compare/v0.255.0...v0.256.0) (2023-02-13)


### Features

* adds semantic release packages ([#769](https://github.com/teamkeel/keel/issues/769)) ([32c161b](https://github.com/teamkeel/keel/commit/32c161b268f8bf91e30eb86b7ced2f0fddf9040e))
* publish keel cli binary as part of github release ([#768](https://github.com/teamkeel/keel/issues/768)) ([4ec5c7c](https://github.com/teamkeel/keel/commit/4ec5c7cd585fd6320d4c93e98039a7132cf3817d))

# [0.255.0](https://github.com/teamkeel/keel/compare/v0.254.0...v0.255.0) (2023-02-13)


### Features

* add missing supported db type ([c0d6d32](https://github.com/teamkeel/keel/commit/c0d6d321ce920aa8fd7d24d57dbb5050726e2361))
* validate database value input types ([6efce54](https://github.com/teamkeel/keel/commit/6efce54ff679064608220c429820a487d42e4e92))

# [0.254.0](https://github.com/teamkeel/keel/compare/v0.253.1...v0.254.0) (2023-02-11)


### Features

* add openapi spec endpoint to http json api ([13a60d8](https://github.com/teamkeel/keel/commit/13a60d8a79503bbbd25ed983cab23515a462ca5f))

## [0.253.1](https://github.com/teamkeel/keel/compare/v0.253.0...v0.253.1) (2023-02-10)


### Bug Fixes

* find one can't look up by a many relationship ([052bae1](https://github.com/teamkeel/keel/commit/052bae11bcc524ed5cfbf3908d7b8c4b61d2ab0f))

# [0.253.0](https://github.com/teamkeel/keel/compare/v0.252.0...v0.253.0) (2023-02-09)


### Features

* headers in expressions ([#764](https://github.com/teamkeel/keel/issues/764)) ([264668b](https://github.com/teamkeel/keel/commit/264668b6c95b96cfbdfbc9302ef3c1bb1b026671))

# [0.252.0](https://github.com/teamkeel/keel/compare/v0.251.1...v0.252.0) (2023-02-08)


### Features

* add fetch() to custom func api ([#745](https://github.com/teamkeel/keel/issues/745)) ([729ff39](https://github.com/teamkeel/keel/commit/729ff3935c41531b79bd904685e3b11e4458e553))

## [0.251.1](https://github.com/teamkeel/keel/compare/v0.251.0...v0.251.1) (2023-02-08)


### Bug Fixes

* generate vitest config into .build so its not in node_modules ([9c3d333](https://github.com/teamkeel/keel/commit/9c3d3337488f68692a6d546ffdd8a8016d5e33fa))

# [0.251.0](https://github.com/teamkeel/keel/compare/v0.250.0...v0.251.0) (2023-02-08)


### Features

* support ctx.headers in parser ([#752](https://github.com/teamkeel/keel/issues/752)) ([4f1b08f](https://github.com/teamkeel/keel/commit/4f1b08fe6518f9aba6434f90919a5e5e407ddd13))

# [0.250.0](https://github.com/teamkeel/keel/compare/v0.249.0...v0.250.0) (2023-02-08)


### Features

* set response headers in functions runtime ([#748](https://github.com/teamkeel/keel/issues/748)) ([81b4ad2](https://github.com/teamkeel/keel/commit/81b4ad2e8d18c6b2ddac0b949b5e6cf4749b3ecf))

# [0.249.0](https://github.com/teamkeel/keel/compare/v0.248.0...v0.249.0) (2023-02-07)


### Features

* add json schema validation to http json api ([b9aca1d](https://github.com/teamkeel/keel/commit/b9aca1d09d6ae61fe05fe28d4ba7c95e3b1d0279))

# [0.248.0](https://github.com/teamkeel/keel/compare/v0.247.3...v0.248.0) (2023-02-07)


### Features

* graceful custom function error handling ([#749](https://github.com/teamkeel/keel/issues/749)) ([adda4a1](https://github.com/teamkeel/keel/commit/adda4a122baa78b8646fae25e9e62ca374e4a9c9))

## [0.247.3](https://github.com/teamkeel/keel/compare/v0.247.2...v0.247.3) (2023-02-06)


### Bug Fixes

* golangci-lint version mismatch ([#746](https://github.com/teamkeel/keel/issues/746)) ([7355501](https://github.com/teamkeel/keel/commit/7355501581004f9645370fd340bf763149ff727e))

## [0.247.2](https://github.com/teamkeel/keel/compare/v0.247.1...v0.247.2) (2023-02-03)


### Bug Fixes

* dont pass ksuid's around, they are just strings ([a6d7adc](https://github.com/teamkeel/keel/commit/a6d7adc1e76feeb58f38f06f4fc7d11d1948b979))

## [0.247.1](https://github.com/teamkeel/keel/compare/v0.247.0...v0.247.1) (2023-02-03)


### Bug Fixes

* redirect custom function output to os.stdout in test cmd ([#733](https://github.com/teamkeel/keel/issues/733)) ([aad2024](https://github.com/teamkeel/keel/commit/aad2024d2adeb82cb383af17f2c235e51fdc6429))

# [0.247.0](https://github.com/teamkeel/keel/compare/v0.246.0...v0.247.0) (2023-02-03)


### Features

* ctx.identity model in custom functions ([#737](https://github.com/teamkeel/keel/issues/737)) ([fa0fc02](https://github.com/teamkeel/keel/commit/fa0fc02abf7351587043afc2c9a995abaea4250e))

# [0.246.0](https://github.com/teamkeel/keel/compare/v0.245.2...v0.246.0) (2023-02-03)


### Features

* request headers in custom functions runtime ([#732](https://github.com/teamkeel/keel/issues/732)) ([61440ca](https://github.com/teamkeel/keel/commit/61440ca3be892fabe23eb519d4702377a6c47432))

## [0.245.2](https://github.com/teamkeel/keel/compare/v0.245.1...v0.245.2) (2023-02-02)


### Reverts

* Revert "feat: add sdkPackageTypes and testingPackage generate options" ([e6a6b2f](https://github.com/teamkeel/keel/commit/e6a6b2fcc4bee9a0056e75969d0caf6baaaefb91))

## [0.245.1](https://github.com/teamkeel/keel/compare/v0.245.0...v0.245.1) (2023-02-01)


### Bug Fixes

* update error handling of custom functions ([#720](https://github.com/teamkeel/keel/issues/720)) ([3eb2702](https://github.com/teamkeel/keel/commit/3eb27021d3ceba83536f46b691598b75204241f8))

# [0.245.0](https://github.com/teamkeel/keel/compare/v0.244.1...v0.245.0) (2023-02-01)


### Features

* validation of config files ([#719](https://github.com/teamkeel/keel/issues/719)) ([90ca9a7](https://github.com/teamkeel/keel/commit/90ca9a7375151aac954a419d195e4ccf0bb7642d))

## [0.244.1](https://github.com/teamkeel/keel/compare/v0.244.0...v0.244.1) (2023-02-01)


### Bug Fixes

* improved error responses on apis ([#726](https://github.com/teamkeel/keel/issues/726)) ([b25a57e](https://github.com/teamkeel/keel/commit/b25a57ea3bf6c78cb60bf0010a08ce4329dff44d))

# [0.244.0](https://github.com/teamkeel/keel/compare/v0.243.0...v0.244.0) (2023-02-01)


### Features

* add sdkPackageTypes and testingPackage generate options ([abb16e0](https://github.com/teamkeel/keel/commit/abb16e066f6d3f5c1752f036c81e43a331bcab0c))

# [0.243.0](https://github.com/teamkeel/keel/compare/v0.242.1...v0.243.0) (2023-01-30)


### Features

* config file loading for environment variables ([#708](https://github.com/teamkeel/keel/issues/708)) ([6f35e43](https://github.com/teamkeel/keel/commit/6f35e436410a8252a292fb4efdd7b88c97ec56d8))

## [0.242.1](https://github.com/teamkeel/keel/compare/v0.242.0...v0.242.1) (2023-01-30)


### Bug Fixes

* make db test suite reusable ([c89777d](https://github.com/teamkeel/keel/commit/c89777da975d93d82ba15d4fda089aae50642ce3))

# [0.242.0](https://github.com/teamkeel/keel/compare/v0.241.1...v0.242.0) (2023-01-28)


### Features

* in and not in expression support for literals ([#706](https://github.com/teamkeel/keel/issues/706)) ([91ee2c4](https://github.com/teamkeel/keel/commit/91ee2c4716f464bedec71e8ac3fe2ace0a58e60a))

## [0.241.1](https://github.com/teamkeel/keel/compare/v0.241.0...v0.241.1) (2023-01-27)


### Bug Fixes

* polyfill crypto.getRandomValues ([#713](https://github.com/teamkeel/keel/issues/713)) ([1b805a9](https://github.com/teamkeel/keel/commit/1b805a9a5c76437dd674bfbe718879ff9b3cd082))

# [0.241.0](https://github.com/teamkeel/keel/compare/v0.240.5...v0.241.0) (2023-01-27)


### Features

* wasm api change ([#712](https://github.com/teamkeel/keel/issues/712)) ([984202f](https://github.com/teamkeel/keel/commit/984202f229b0f6a67f44e41220f5847df767f80f))

## [0.240.5](https://github.com/teamkeel/keel/compare/v0.240.4...v0.240.5) (2023-01-27)


### Bug Fixes

* include dist files in npm pack for wasm package ([#709](https://github.com/teamkeel/keel/issues/709)) ([f4b1254](https://github.com/teamkeel/keel/commit/f4b1254179fe605e267d41257df24f0d2bea6b3e))

## [0.240.4](https://github.com/teamkeel/keel/compare/v0.240.3...v0.240.4) (2023-01-27)


### Bug Fixes

* revert checkout action in npm publish workflow ([67508b0](https://github.com/teamkeel/keel/commit/67508b0af0205af1f96ac69609df33b310ba680c))
* update npm publish github action workflow ([664a5cc](https://github.com/teamkeel/keel/commit/664a5cc6a39b14df1b73382f2099968648af6f20))

## [0.240.3](https://github.com/teamkeel/keel/compare/v0.240.2...v0.240.3) (2023-01-26)


### Bug Fixes

* generate correct typings for testing action executor ([c203c77](https://github.com/teamkeel/keel/commit/c203c770fcf20c8ed8a9b22906691d9e2d37a756))
* move ConnectionInfo struct to db package to remove wasm dep on docker ([#704](https://github.com/teamkeel/keel/issues/704)) ([0ba328b](https://github.com/teamkeel/keel/commit/0ba328b21014221326ec00d882630863f06b42ce))
* oneOf list operator ([#700](https://github.com/teamkeel/keel/issues/700)) ([6a2923f](https://github.com/teamkeel/keel/commit/6a2923f9bee1d603c4bcb855aea155aebb1fbc3d))

## [0.240.2](https://github.com/teamkeel/keel/compare/v0.240.1...v0.240.2) (2023-01-26)


### Bug Fixes

* fix explicit list inputs in json schema and add temp support for authenticate ([b1c90d4](https://github.com/teamkeel/keel/commit/b1c90d41e4076ba653397ade9c07835124c48e95))

## [0.240.1](https://github.com/teamkeel/keel/compare/v0.240.0...v0.240.1) (2023-01-25)


### Bug Fixes

* highlight missing model in validation error ([#694](https://github.com/teamkeel/keel/issues/694)) ([7835992](https://github.com/teamkeel/keel/commit/783599299c8cb3d6e863498001e1854cc7bbe53b))

# [0.240.0](https://github.com/teamkeel/keel/compare/v0.239.0...v0.240.0) (2023-01-25)


### Features

* release packages ([#692](https://github.com/teamkeel/keel/issues/692)) ([44fb124](https://github.com/teamkeel/keel/commit/44fb1241c31f2e9ea56e8f4f35628cc039ebf64d))

# [0.239.0](https://github.com/teamkeel/keel/compare/v0.238.0...v0.239.0) (2023-01-24)


### Features

* release packages ([#690](https://github.com/teamkeel/keel/issues/690)) ([9785cee](https://github.com/teamkeel/keel/commit/9785cee5c2af00bd0b8b30fad2203bf02da9bb1c))

# [0.238.0](https://github.com/teamkeel/keel/compare/v0.237.0...v0.238.0) (2023-01-24)


### Features

* custom toHaveError vitest matcher ([#688](https://github.com/teamkeel/keel/issues/688)) ([414af82](https://github.com/teamkeel/keel/commit/414af822771ed57f6d1b190d1e5e0e7c0bb4be74))

# [0.237.0](https://github.com/teamkeel/keel/compare/v0.236.3...v0.237.0) (2023-01-24)


### Bug Fixes

* drop unnecessary function ([1cfd88c](https://github.com/teamkeel/keel/commit/1cfd88c1b304adcd77125080c30261bb0ea9d8db))
* inputs can only be implicit for operations, never for functions ([305456e](https://github.com/teamkeel/keel/commit/305456eb18b9f32f742760e697e592333380ed9c))
* making local db internal transaction usage more explicit ([e60a694](https://github.com/teamkeel/keel/commit/e60a694f376c1b80db425f2e1a8da24a9f5e017c))
* update wasm handler callsite ([#687](https://github.com/teamkeel/keel/issues/687)) ([8a90a57](https://github.com/teamkeel/keel/commit/8a90a579b378ba41ddafb1bbb031097894d8e54c))


### Features

* database api ([5e8f08a](https://github.com/teamkeel/keel/commit/5e8f08a6564b154afc13255949d9046bf37182b1))
* publish testing-runtime pkg ([#686](https://github.com/teamkeel/keel/issues/686)) ([427ff4f](https://github.com/teamkeel/keel/commit/427ff4fa3dceb41df1ad5bc4679a4022fbb5de56))

## [0.236.3](https://github.com/teamkeel/keel/compare/v0.236.2...v0.236.3) (2023-01-19)


### Bug Fixes

* delete op support for nested model lookups ([#671](https://github.com/teamkeel/keel/issues/671)) ([9d3a486](https://github.com/teamkeel/keel/commit/9d3a486d97c772d0de7c57de072c6328867f5a61))

## [0.236.2](https://github.com/teamkeel/keel/compare/v0.236.1...v0.236.2) (2023-01-18)


### Bug Fixes

* argument naming of nested model inputs ([#673](https://github.com/teamkeel/keel/issues/673)) ([a10a4dd](https://github.com/teamkeel/keel/commit/a10a4ddeab502d6d0977a8f0da38a12c4e10a1fb))
* use pnpm install for functions-runtime package ([7a0c4be](https://github.com/teamkeel/keel/commit/7a0c4be3bc940974e41875182831babdc754cb7e))

## [0.236.1](https://github.com/teamkeel/keel/compare/v0.236.0...v0.236.1) (2023-01-13)


### Bug Fixes

* model usable multiple times in input or expression ([#665](https://github.com/teamkeel/keel/issues/665)) ([33f27e9](https://github.com/teamkeel/keel/commit/33f27e9f1db3a09723ab691bafeea095a5bbf57f))

# [0.236.0](https://github.com/teamkeel/keel/compare/v0.235.4...v0.236.0) (2023-01-12)


### Features

* use on delete set null when field is optional ([#664](https://github.com/teamkeel/keel/issues/664)) ([8ff6115](https://github.com/teamkeel/keel/commit/8ff61155d66bba946f8edf142ecef60e9b727851))

## [0.235.4](https://github.com/teamkeel/keel/compare/v0.235.3...v0.235.4) (2023-01-11)


### Bug Fixes

* handle empty graphql query object ([#663](https://github.com/teamkeel/keel/issues/663)) ([68eac9b](https://github.com/teamkeel/keel/commit/68eac9b1c5fa0e61347bbb51f4faf4abf4a08da9))

## [0.235.3](https://github.com/teamkeel/keel/compare/v0.235.2...v0.235.3) (2023-01-09)


### Bug Fixes

* permissions on related model data in create op ([#660](https://github.com/teamkeel/keel/issues/660)) ([c2dd0fc](https://github.com/teamkeel/keel/commit/c2dd0fc4deb235e46dcfd7f734703093ecee1efd))

## [0.235.2](https://github.com/teamkeel/keel/compare/v0.235.1...v0.235.2) (2023-01-04)


### Bug Fixes

* relax custom function unique filter validation ([#661](https://github.com/teamkeel/keel/issues/661)) ([f932a43](https://github.com/teamkeel/keel/commit/f932a43b2c1776e634a8238e299c1144cfca8404))

## [0.235.1](https://github.com/teamkeel/keel/compare/v0.235.0...v0.235.1) (2022-12-22)


### Bug Fixes

* check that identity exists for jwt ([#659](https://github.com/teamkeel/keel/issues/659)) ([999b129](https://github.com/teamkeel/keel/commit/999b129c31a5df4c59c173d03e4d664b048dd1aa))

# [0.235.0](https://github.com/teamkeel/keel/compare/v0.234.0...v0.235.0) (2022-12-16)


### Bug Fixes

* backport fix ([39fb83e](https://github.com/teamkeel/keel/commit/39fb83e02de1fa2ae09357710698251b2827084e))
* fix the issue ([7fc2b37](https://github.com/teamkeel/keel/commit/7fc2b370139bc7b4560816fe3ba1fbe00b1465f6))


### Features

* test with broken query ([a3b5ef9](https://github.com/teamkeel/keel/commit/a3b5ef93228d470ec477cbc68d51a8a0979f0cc1))

# [0.234.0](https://github.com/teamkeel/keel/compare/v0.233.0...v0.234.0) (2022-12-14)


### Features

* add [@default](https://github.com/default) validation for expression required ([2560d24](https://github.com/teamkeel/keel/commit/2560d246c58ae2708a88bf065ba00078db777e9d))
* add [@default](https://github.com/default) validation for multiple conditions ([472c727](https://github.com/teamkeel/keel/commit/472c727d832d6d9f605a88af1c2cdadd54441f72))
* add [@default](https://github.com/default) validation for operators ([9cef9e7](https://github.com/teamkeel/keel/commit/9cef9e7127a67e7fb438aef5d4076951224307e2))

# [0.233.0](https://github.com/teamkeel/keel/compare/v0.232.1...v0.233.0) (2022-12-14)


### Features

* add [@default](https://github.com/default) validation ([2730468](https://github.com/teamkeel/keel/commit/27304687b90f2c6060db3e2edc5fb212a1ff4baa))

## [0.232.1](https://github.com/teamkeel/keel/compare/v0.232.0...v0.232.1) (2022-12-13)


### Bug Fixes

* authenticate returning empty token ([#652](https://github.com/teamkeel/keel/issues/652)) ([1326963](https://github.com/teamkeel/keel/commit/13269631a3e41e8fb18f0d8b7f430088ba19fc69))

# [0.232.0](https://github.com/teamkeel/keel/compare/v0.231.0...v0.232.0) (2022-12-12)


### Features

* change rule of all fields required to only operations ([e0bd763](https://github.com/teamkeel/keel/commit/e0bd763e5ce791cbbdbc065f453ddf31f148cab8))

# [0.231.0](https://github.com/teamkeel/keel/compare/v0.230.1...v0.231.0) (2022-12-09)


### Features

* cascading delete by default ([#648](https://github.com/teamkeel/keel/issues/648)) ([4e26de4](https://github.com/teamkeel/keel/commit/4e26de416bfb16bff68d213056a7c4e4ef803975))

## [0.230.1](https://github.com/teamkeel/keel/compare/v0.230.0...v0.230.1) (2022-12-09)


### Bug Fixes

* api models need to exist ([d57a3c3](https://github.com/teamkeel/keel/commit/d57a3c3b10e7c0da8277a92c7334814035107f8f))

# [0.230.0](https://github.com/teamkeel/keel/compare/v0.229.0...v0.230.0) (2022-12-09)


### Features

* delete unused sort code ([65fa1f1](https://github.com/teamkeel/keel/commit/65fa1f17d5352dae31558d93d7274eba2aa40b89))

# [0.229.0](https://github.com/teamkeel/keel/compare/v0.228.4...v0.229.0) (2022-12-08)


### Features

* allow functions unused explicit input ([07f849c](https://github.com/teamkeel/keel/commit/07f849c15505e71d9f0d833db152957dc8ac94bf))

## [0.228.4](https://github.com/teamkeel/keel/compare/v0.228.3...v0.228.4) (2022-12-06)


### Bug Fixes

* present schema errors inline without panicking ([#646](https://github.com/teamkeel/keel/issues/646)) ([d28e5b0](https://github.com/teamkeel/keel/commit/d28e5b0529a764ab983a0474b17bbc46fe4b50e9))

## [0.228.3](https://github.com/teamkeel/keel/compare/v0.228.2...v0.228.3) (2022-12-05)


### Bug Fixes

* drop foreign keys that reference dropped table ([#645](https://github.com/teamkeel/keel/issues/645)) ([3ccd61e](https://github.com/teamkeel/keel/commit/3ccd61edd0009180981ae93908a6119c188abdad))

## [0.228.2](https://github.com/teamkeel/keel/compare/v0.228.1...v0.228.2) (2022-12-01)


### Bug Fixes

* use pq.QuoteIdentifier for sql idents ([#644](https://github.com/teamkeel/keel/issues/644)) ([f757bee](https://github.com/teamkeel/keel/commit/f757beeb03ba878c2654e9488a22208c76ede152))

## [0.228.1](https://github.com/teamkeel/keel/compare/v0.228.0...v0.228.1) (2022-12-01)


### Bug Fixes

* quote table names ([#643](https://github.com/teamkeel/keel/issues/643)) ([ed92418](https://github.com/teamkeel/keel/commit/ed92418ff1963ad1c870e611f5ea43a0c1b8e818))

# [0.228.0](https://github.com/teamkeel/keel/compare/v0.227.0...v0.228.0) (2022-12-01)


### Bug Fixes

* test not to throw ([117e3da](https://github.com/teamkeel/keel/commit/117e3daac9a5083510e9f8f848181207e2d415d3))


### Features

* add queryResolverFromEnv dataapi test ([8a39357](https://github.com/teamkeel/keel/commit/8a39357f334e4e58a844635821c25d60b418a5c9))
* add queryResolverFromEnv throws test ([12e79b4](https://github.com/teamkeel/keel/commit/12e79b48308f046629ddeb6cf384c1f2aa8c7802))

# [0.227.0](https://github.com/teamkeel/keel/compare/v0.226.0...v0.227.0) (2022-12-01)


### Features

* simplify raw query ([bd8f08a](https://github.com/teamkeel/keel/commit/bd8f08a329f4b45d1d091f96789095a82ea78b8a))

# [0.226.0](https://github.com/teamkeel/keel/compare/v0.225.1...v0.226.0) (2022-12-01)


### Features

* refactor query resolver ([c30d141](https://github.com/teamkeel/keel/commit/c30d141be9d6a584ab177592b2a156584e5b6761))

## [0.225.1](https://github.com/teamkeel/keel/compare/v0.225.0...v0.225.1) (2022-12-01)


### Bug Fixes

* message when what is thrown is not an error ([8c87f05](https://github.com/teamkeel/keel/commit/8c87f05d2ed96dc6015eb87bc154e7f49343aba9))

# [0.225.0](https://github.com/teamkeel/keel/compare/v0.224.3...v0.225.0) (2022-12-01)


### Bug Fixes

* improve comment ([2eb962d](https://github.com/teamkeel/keel/commit/2eb962d624f44d5a4f8c67279308aae83bd8c24e))
* logger comment ([f344fc0](https://github.com/teamkeel/keel/commit/f344fc0cf0ee5f40feb8fc890fbef0998fac7626))


### Features

* drop runInitialSql function ([3540cf3](https://github.com/teamkeel/keel/commit/3540cf395e4bd81c3c508f5f7c088d5271a0dd54))

## [0.224.3](https://github.com/teamkeel/keel/compare/v0.224.2...v0.224.3) (2022-11-30)


### Bug Fixes

* nil pointer crash vulnerability ([#635](https://github.com/teamkeel/keel/issues/635)) ([1ad8d5d](https://github.com/teamkeel/keel/commit/1ad8d5d591e13eafcac3d2d8b7443ba651b23850))

## [0.224.2](https://github.com/teamkeel/keel/compare/v0.224.1...v0.224.2) (2022-11-30)


### Bug Fixes

* delete operation response not resolved properly in GraphQl api layer ([#634](https://github.com/teamkeel/keel/issues/634)) ([3e55ea7](https://github.com/teamkeel/keel/commit/3e55ea7fa8ab3d44af70f7143f81e3e956820c4e))

## [0.224.1](https://github.com/teamkeel/keel/compare/v0.224.0...v0.224.1) (2022-11-30)


### Bug Fixes

* validation rule incorrect logic ([#633](https://github.com/teamkeel/keel/issues/633)) ([ef3a04f](https://github.com/teamkeel/keel/commit/ef3a04f9eab9f26ebe88262dda23ece4f2310115))

# [0.224.0](https://github.com/teamkeel/keel/compare/v0.223.0...v0.224.0) (2022-11-30)


### Features

* use correct typescript compiler ([acc1a5d](https://github.com/teamkeel/keel/commit/acc1a5d52ca923a4d8c418db797c97cbe0a975f0))

# [0.223.0](https://github.com/teamkeel/keel/compare/v0.222.4...v0.223.0) (2022-11-30)


### Features

* upgrade validation of missing Create inputs ([#626](https://github.com/teamkeel/keel/issues/626)) ([eda3880](https://github.com/teamkeel/keel/commit/eda3880e080c60abdd50f2db7a0cbfff9d8cdf02))

## [0.222.4](https://github.com/teamkeel/keel/compare/v0.222.3...v0.222.4) (2022-11-30)


### Bug Fixes

* add typescript installation to readme ([2cc4f58](https://github.com/teamkeel/keel/commit/2cc4f58202c81f722d20b905b3b325027687c46d))

## [0.222.3](https://github.com/teamkeel/keel/compare/v0.222.2...v0.222.3) (2022-11-29)


### Bug Fixes

* move datetime conversion to graphql resolver level ([#628](https://github.com/teamkeel/keel/issues/628)) ([5480cf0](https://github.com/teamkeel/keel/commit/5480cf0fe0dc61ae2cb968319d275a16cfb0097b))

## [0.222.2](https://github.com/teamkeel/keel/compare/v0.222.1...v0.222.2) (2022-11-29)


### Bug Fixes

* convert timestamps returned in iso8601 from custom functions ([#625](https://github.com/teamkeel/keel/issues/625)) ([3c209e6](https://github.com/teamkeel/keel/commit/3c209e6f4607f443d9d3866c4d969d23f5d416b5))

## [0.222.1](https://github.com/teamkeel/keel/compare/v0.222.0...v0.222.1) (2022-11-29)


### Bug Fixes

* assign model to null in set attribute ([#621](https://github.com/teamkeel/keel/issues/621)) ([1386786](https://github.com/teamkeel/keel/commit/138678663e36b471ad380e820e5caea878f846a1))

# [0.222.0](https://github.com/teamkeel/keel/compare/v0.221.4...v0.222.0) (2022-11-29)


### Features

* new validation rule (reciprocal relationship count <= 1) ([#622](https://github.com/teamkeel/keel/issues/622)) ([e052514](https://github.com/teamkeel/keel/commit/e052514c6e704c85e168c82feb9e4325f2f51820))

## [0.221.4](https://github.com/teamkeel/keel/compare/v0.221.3...v0.221.4) (2022-11-29)


### Bug Fixes

* date and timestamp input parsing fixed ([#620](https://github.com/teamkeel/keel/issues/620)) ([401cf5e](https://github.com/teamkeel/keel/commit/401cf5ee461adb7258b3857374fb737f3da6282c))

## [0.221.3](https://github.com/teamkeel/keel/compare/v0.221.2...v0.221.3) (2022-11-28)


### Bug Fixes

* when Model field is [@unique](https://github.com/unique), then generated FK field should be made  [@unique](https://github.com/unique) ([#619](https://github.com/teamkeel/keel/issues/619)) ([2062277](https://github.com/teamkeel/keel/commit/20622777d3e993355f8178d2a6f76aa4c4c3a208))

## [0.221.2](https://github.com/teamkeel/keel/compare/v0.221.1...v0.221.2) (2022-11-25)


### Bug Fixes

* create action to accept model, model.id and modelId ([#616](https://github.com/teamkeel/keel/issues/616)) ([3efbc3f](https://github.com/teamkeel/keel/commit/3efbc3f8816fb2d61e3bb44c33b1c98b2ba819ae))

## [0.221.1](https://github.com/teamkeel/keel/compare/v0.221.0...v0.221.1) (2022-11-25)


### Bug Fixes

* enums as explicit inputs ([#612](https://github.com/teamkeel/keel/issues/612)) ([027cf43](https://github.com/teamkeel/keel/commit/027cf43c3071d71c2a85eef9f071aed403cb5ce4))

# [0.221.0](https://github.com/teamkeel/keel/compare/v0.220.0...v0.221.0) (2022-11-24)


### Features

* expose raw sql method ([#609](https://github.com/teamkeel/keel/issues/609)) ([389b55c](https://github.com/teamkeel/keel/commit/389b55c9e54d1d0b0b2c975a7043abb00e001793))

# [0.220.0](https://github.com/teamkeel/keel/compare/v0.219.0...v0.220.0) (2022-11-24)


### Features

* add functions-runtime readme ([#608](https://github.com/teamkeel/keel/issues/608)) ([e760dba](https://github.com/teamkeel/keel/commit/e760dbaa8a473af4cd25b62a31fc05a7cda40de5))

# [0.219.0](https://github.com/teamkeel/keel/compare/v0.218.0...v0.219.0) (2022-11-24)


### Features

* combine sdk and runtime into functions-runtime ([#607](https://github.com/teamkeel/keel/issues/607)) ([c64996c](https://github.com/teamkeel/keel/commit/c64996c30fcfb8b6af12b54b731bc56660d3f757))

# [0.218.0](https://github.com/teamkeel/keel/compare/v0.217.1...v0.218.0) (2022-11-23)


### Features

* upgrade validation to support relationship fields in create operations ([8dde9a9](https://github.com/teamkeel/keel/commit/8dde9a9841a329600c90e96b55b574049f613b4b))

## [0.217.1](https://github.com/teamkeel/keel/compare/v0.217.0...v0.217.1) (2022-11-23)


### Bug Fixes

* omit graphql object in an optional m:1 lookup ([#604](https://github.com/teamkeel/keel/issues/604)) ([a4bf69e](https://github.com/teamkeel/keel/commit/a4bf69eb7ff5757b16038ecbf890a3ea7a87eb89))

# [0.217.0](https://github.com/teamkeel/keel/compare/v0.216.0...v0.217.0) (2022-11-22)


### Features

* graphql nested model pagination ([#600](https://github.com/teamkeel/keel/issues/600)) ([93d764e](https://github.com/teamkeel/keel/commit/93d764eedb998350bd619d981f509233ccd5e8fe))

# [0.216.0](https://github.com/teamkeel/keel/compare/v0.215.1...v0.216.0) (2022-11-21)


### Features

* graphql nested model resolving ([#599](https://github.com/teamkeel/keel/issues/599)) ([5979bc4](https://github.com/teamkeel/keel/commit/5979bc43ffd6499ac63f2f2e76ae37bc247e7351))

## [0.215.1](https://github.com/teamkeel/keel/compare/v0.215.0...v0.215.1) (2022-11-17)


### Bug Fixes

* send valid request for authenticate action ([1278a94](https://github.com/teamkeel/keel/commit/1278a94eb1c5a8be03141a13af67d439d0494501))

# [0.215.0](https://github.com/teamkeel/keel/compare/v0.214.1...v0.215.0) (2022-11-17)


### Features

* replacing gorm query builder in actions code ([#587](https://github.com/teamkeel/keel/issues/587)) ([18a6ebe](https://github.com/teamkeel/keel/commit/18a6ebe2a93bd38f14cd96e7077094f2df9d8e90))

## [0.214.1](https://github.com/teamkeel/keel/compare/v0.214.0...v0.214.1) (2022-11-17)


### Bug Fixes

* test ([5ee7fb6](https://github.com/teamkeel/keel/commit/5ee7fb6cfed5a2b2affebfe6ad63cd50e7b8405b))

# [0.214.0](https://github.com/teamkeel/keel/compare/v0.213.1...v0.214.0) (2022-11-17)


### Features

* add start of json schema support ([b6fa559](https://github.com/teamkeel/keel/commit/b6fa5599c1208529d35c179fd9e92773dc639b2b))

## [0.213.1](https://github.com/teamkeel/keel/compare/v0.213.0...v0.213.1) (2022-11-17)


### Bug Fixes

* usage of list actions in integration tests ([8f17eb2](https://github.com/teamkeel/keel/commit/8f17eb271f641f77ff38a23ef44addabe261f761))

# [0.213.0](https://github.com/teamkeel/keel/compare/v0.212.0...v0.213.0) (2022-11-17)


### Features

* add test ([cc703b2](https://github.com/teamkeel/keel/commit/cc703b2d3023b5701d221d67caaa54026416529c))

# [0.212.0](https://github.com/teamkeel/keel/compare/v0.211.0...v0.212.0) (2022-11-15)


### Features

* remove Date fromNow ([6f1e6d3](https://github.com/teamkeel/keel/commit/6f1e6d3357b4f8e5d5068758091d099f9ada7981))

# [0.211.0](https://github.com/teamkeel/keel/compare/v0.210.0...v0.211.0) (2022-11-14)


### Features

* relationship support in implicit inputs - first pass ([#571](https://github.com/teamkeel/keel/issues/571)) ([b75a4ee](https://github.com/teamkeel/keel/commit/b75a4ee0c77d37709d30739ff508bf508c605da5))

# [0.210.0](https://github.com/teamkeel/keel/compare/v0.209.0...v0.210.0) (2022-11-14)


### Features

* remove prefer offline optimization ([#573](https://github.com/teamkeel/keel/issues/573)) ([d8253a6](https://github.com/teamkeel/keel/commit/d8253a6fb6b3ea4bf389ef778553f8310d06bd4b))

# [0.209.0](https://github.com/teamkeel/keel/compare/v0.208.1...v0.209.0) (2022-11-14)


### Features

* relationships in expressions - first pass ([#567](https://github.com/teamkeel/keel/issues/567)) ([7e52035](https://github.com/teamkeel/keel/commit/7e520354fd449f2a6ad82ce499e819670668f734))

## [0.208.1](https://github.com/teamkeel/keel/compare/v0.208.0...v0.208.1) (2022-11-13)


### Bug Fixes

* bug in Role-based permissions (RUN-179) ([#569](https://github.com/teamkeel/keel/issues/569)) ([6ab0930](https://github.com/teamkeel/keel/commit/6ab09301687fcb692d192370b6bae006cd7ae6e8))

# [0.208.0](https://github.com/teamkeel/keel/compare/v0.207.0...v0.208.0) (2022-11-11)


### Features

* align RPC API with documented spec ([fe3e81c](https://github.com/teamkeel/keel/commit/fe3e81c7f0bb53c15265f360ee7d87b9276bcc86))

# [0.207.0](https://github.com/teamkeel/keel/compare/v0.206.1...v0.207.0) (2022-11-11)


### Features

* implement Role-based permission rules ([#565](https://github.com/teamkeel/keel/issues/565)) ([beaf548](https://github.com/teamkeel/keel/commit/beaf548883c654bca4e40494e5aa4742c8f26309))

## [0.206.1](https://github.com/teamkeel/keel/compare/v0.206.0...v0.206.1) (2022-11-10)


### Bug Fixes

* unique input or where required for get, update and delete ([#563](https://github.com/teamkeel/keel/issues/563)) ([f7209b1](https://github.com/teamkeel/keel/commit/f7209b1df65abecfe8105edea49be64b20ce11cf))

# [0.206.0](https://github.com/teamkeel/keel/compare/v0.205.1...v0.206.0) (2022-11-09)


### Features

* bump version ([#562](https://github.com/teamkeel/keel/issues/562)) ([33597e2](https://github.com/teamkeel/keel/commit/33597e212cb1d117e39042c517f0b20fb151dd0f))
* reorganise npm packages into single modules directory ([#561](https://github.com/teamkeel/keel/issues/561)) ([000f43b](https://github.com/teamkeel/keel/commit/000f43b9fc1ec0972fe0367a3617c008da39a265))

## [0.205.1](https://github.com/teamkeel/keel/compare/v0.205.0...v0.205.1) (2022-11-09)


### Bug Fixes

* migrations for relationships ([#560](https://github.com/teamkeel/keel/issues/560)) ([0c8ae81](https://github.com/teamkeel/keel/commit/0c8ae819e3867589f4854e66d5a9e88dc829ff10))

# [0.205.0](https://github.com/teamkeel/keel/compare/v0.204.0...v0.205.0) (2022-11-08)


### Features

* support explicit inputs for enums ([#559](https://github.com/teamkeel/keel/issues/559)) ([0c3a853](https://github.com/teamkeel/keel/commit/0c3a85364e33c3bb160e695afd5c351167147360))

# [0.204.0](https://github.com/teamkeel/keel/compare/v0.203.0...v0.204.0) (2022-11-08)


### Features

* change postgres to 11.13 ([e7222ad](https://github.com/teamkeel/keel/commit/e7222adb82a9c725310896a5af2b3bc5467ae7b8))

# [0.203.0](https://github.com/teamkeel/keel/compare/v0.202.0...v0.203.0) (2022-11-07)


### Features

* support relationship fields in migrations and database ([5def274](https://github.com/teamkeel/keel/commit/5def2746f4bf9a84c31926dccb9b65764b263946))

# [0.202.0](https://github.com/teamkeel/keel/compare/v0.201.0...v0.202.0) (2022-11-07)


### Features

* adds validation for invalid implicit belongs_to relationship ([#558](https://github.com/teamkeel/keel/issues/558)) ([95504a7](https://github.com/teamkeel/keel/commit/95504a796cce697cbbd55b836076a51f1ccfe430))

# [0.201.0](https://github.com/teamkeel/keel/compare/v0.200.0...v0.201.0) (2022-11-04)


### Features

* revert "feat: throw an error if an action returns any errors ([#541](https://github.com/teamkeel/keel/issues/541))" ([#554](https://github.com/teamkeel/keel/issues/554)) ([81b5d2d](https://github.com/teamkeel/keel/commit/81b5d2d5f3681ff02ab53d89447b259af8f6b4eb))

# [0.200.0](https://github.com/teamkeel/keel/compare/v0.199.6...v0.200.0) (2022-11-04)


### Features

* throw an error if an action returns any errors ([#541](https://github.com/teamkeel/keel/issues/541)) ([b1fc78c](https://github.com/teamkeel/keel/commit/b1fc78cb12f1c8e962e13b20adfc6e8a5a8d1fae))

## [0.199.6](https://github.com/teamkeel/keel/compare/v0.199.5...v0.199.6) (2022-11-04)


### Bug Fixes

* safer use of typeof ([de297f1](https://github.com/teamkeel/keel/commit/de297f170f42aeb5531277059051d923a1254405))

## [0.199.5](https://github.com/teamkeel/keel/compare/v0.199.4...v0.199.5) (2022-11-04)


### Bug Fixes

* casing bug ([5d711ac](https://github.com/teamkeel/keel/commit/5d711ac71eb1e8f681cfceda09e22626b518c505))

## [0.199.4](https://github.com/teamkeel/keel/compare/v0.199.3...v0.199.4) (2022-11-03)


### Bug Fixes

* cors all origins and allow credentials ([cd0adf9](https://github.com/teamkeel/keel/commit/cd0adf9d04b15fbc13ccea873078c1f141499d8a))

## [0.199.3](https://github.com/teamkeel/keel/compare/v0.199.2...v0.199.3) (2022-11-03)


### Bug Fixes

* fixed comparison operators for number, date, time ([#548](https://github.com/teamkeel/keel/issues/548)) ([ce759fe](https://github.com/teamkeel/keel/commit/ce759fe3befc2330b65a768181c7dfad064e4697))

## [0.199.2](https://github.com/teamkeel/keel/compare/v0.199.1...v0.199.2) (2022-11-03)


### Bug Fixes

* createdAt and updatedAt set to UTC ([#546](https://github.com/teamkeel/keel/issues/546)) ([eedf8cf](https://github.com/teamkeel/keel/commit/eedf8cf1de00575e552bbc5ce2cd20dfd48f9ebe))
* delete() not returning success from gql ([#545](https://github.com/teamkeel/keel/issues/545)) ([e32f9d8](https://github.com/teamkeel/keel/commit/e32f9d83b4b6563a2d0ecd3cc860bbed835d50f5))

## [0.199.1](https://github.com/teamkeel/keel/compare/v0.199.0...v0.199.1) (2022-11-02)


### Bug Fixes

* wasm ([#542](https://github.com/teamkeel/keel/issues/542)) ([6dac8b5](https://github.com/teamkeel/keel/commit/6dac8b543209744dc6af7be5b5e4c1296b1b1a4a))

# [0.199.0](https://github.com/teamkeel/keel/compare/v0.198.0...v0.199.0) (2022-11-02)


### Features

* update constraint types ([#540](https://github.com/teamkeel/keel/issues/540)) ([0d79aaa](https://github.com/teamkeel/keel/commit/0d79aaae3f58628d1c5a0be8dcdad54c6d8493ad))

# [0.198.0](https://github.com/teamkeel/keel/compare/v0.197.0...v0.198.0) (2022-11-02)


### Features

* unify operator names with main runtime ([#539](https://github.com/teamkeel/keel/issues/539)) ([c61cf7a](https://github.com/teamkeel/keel/commit/c61cf7ae88f94126a54acff74ee07880d8d74e02))

# [0.197.0](https://github.com/teamkeel/keel/compare/v0.196.0...v0.197.0) (2022-11-01)


### Features

* create tracing extension for graphql ([#532](https://github.com/teamkeel/keel/issues/532)) ([1cb118e](https://github.com/teamkeel/keel/commit/1cb118e0000cb499ffe36dcdedc4fbce81c2e66e))

# [0.196.0](https://github.com/teamkeel/keel/compare/v0.195.0...v0.196.0) (2022-11-01)


### Bug Fixes

* add comment ([adc843a](https://github.com/teamkeel/keel/commit/adc843a001f34622a81596cbc8590bdcbfa46fa2))


### Features

* data api query resolver ([60d9943](https://github.com/teamkeel/keel/commit/60d9943762bbb404aab026e4c27fdb235f35994d))

# [0.195.0](https://github.com/teamkeel/keel/compare/v0.194.0...v0.195.0) (2022-11-01)


### Features

* reverse complex typescript ([7b109ae](https://github.com/teamkeel/keel/commit/7b109ae55c55b878b5017d929bc1fe7b674547fd))
* tighten constraints to type ([af58063](https://github.com/teamkeel/keel/commit/af58063ca7ffd4b3d3ad2dd7fc56c053ea43f324))

# [0.194.0](https://github.com/teamkeel/keel/compare/v0.193.0...v0.194.0) (2022-11-01)


### Features

* make default operation permission = no permission ([#534](https://github.com/teamkeel/keel/issues/534)) ([8a81b1d](https://github.com/teamkeel/keel/commit/8a81b1dffc13a956d3cf38aab44e75f3a30e1427))

# [0.193.0](https://github.com/teamkeel/keel/compare/v0.192.0...v0.193.0) (2022-10-28)


### Features

* collect all tests ([#528](https://github.com/teamkeel/keel/issues/528)) ([d2b0e64](https://github.com/teamkeel/keel/commit/d2b0e64c09604d7df3b3a4bf1a896e7ba40b14b4))

# [0.192.0](https://github.com/teamkeel/keel/compare/v0.191.1...v0.192.0) (2022-10-28)


### Features

* report individual test results ([#523](https://github.com/teamkeel/keel/issues/523)) ([d603c0b](https://github.com/teamkeel/keel/commit/d603c0bac1126b8198021240f00b88163eebb0f1))

## [0.191.1](https://github.com/teamkeel/keel/compare/v0.191.0...v0.191.1) (2022-10-28)


### Bug Fixes

* update readme ([ea56faa](https://github.com/teamkeel/keel/commit/ea56faa0426bb76acbcb2bb2c3f4cd2f1ca5f64e))

# [0.191.0](https://github.com/teamkeel/keel/compare/v0.190.0...v0.191.0) (2022-10-27)


### Features

* adapt to updated sdk ([e5ade15](https://github.com/teamkeel/keel/commit/e5ade155058ca86412d90630036f668eddcfb68b))

# [0.190.0](https://github.com/teamkeel/keel/compare/v0.189.0...v0.190.0) (2022-10-27)


### Features

* export query resolver from sdk ([7ca53fc](https://github.com/teamkeel/keel/commit/7ca53fcdca7046862af9a9c2a1b8b8e68536fe58))

# [0.189.0](https://github.com/teamkeel/keel/compare/v0.188.0...v0.189.0) (2022-10-27)


### Features

* export sdk db resolver ([2f62f83](https://github.com/teamkeel/keel/commit/2f62f839e6e2677cc24c56babe89e5ec73f1e010))

# [0.188.0](https://github.com/teamkeel/keel/compare/v0.187.0...v0.188.0) (2022-10-27)


### Features

* fix import ([5f5c11d](https://github.com/teamkeel/keel/commit/5f5c11dc1a8d70c5af42590717562d598f06836c))

# [0.187.0](https://github.com/teamkeel/keel/compare/v0.186.0...v0.187.0) (2022-10-27)


### Bug Fixes

* update index.d.ts ([859d821](https://github.com/teamkeel/keel/commit/859d821682751590d434adb3c0fe6e4c62db06dd))


### Features

* add queryResolverFromEnv ([3e2026a](https://github.com/teamkeel/keel/commit/3e2026aa26136c292d3d3fd0add4a592217e27a2))
* separate query resolver ([e8121b3](https://github.com/teamkeel/keel/commit/e8121b33329e29f1bdca0596e98de242000f5eb7))

# [0.186.0](https://github.com/teamkeel/keel/compare/v0.185.0...v0.186.0) (2022-10-27)


### Features

* ctx.now supported on schema ([#511](https://github.com/teamkeel/keel/issues/511)) ([3893878](https://github.com/teamkeel/keel/commit/3893878e2fa81f27fbc227678237db1f65559d14))

# [0.185.0](https://github.com/teamkeel/keel/compare/v0.184.0...v0.185.0) (2022-10-27)


### Features

* support pattern in test cmd ([#510](https://github.com/teamkeel/keel/issues/510)) ([dc5e1c2](https://github.com/teamkeel/keel/commit/dc5e1c2eee1818e2cffa61e9c71dead6bdd3e631))

# [0.184.0](https://github.com/teamkeel/keel/compare/v0.183.1...v0.184.0) (2022-10-27)


### Features

* release packages ([#514](https://github.com/teamkeel/keel/issues/514)) ([edcdb1c](https://github.com/teamkeel/keel/commit/edcdb1c17fd4287879aad57fc2090b4591c7af97))

## [0.183.1](https://github.com/teamkeel/keel/compare/v0.183.0...v0.183.1) (2022-10-27)


### Bug Fixes

* asc and desc match casing ([#513](https://github.com/teamkeel/keel/issues/513)) ([8dd7713](https://github.com/teamkeel/keel/commit/8dd7713ddd89ffe763391ad85aeece713481045c))

# [0.183.0](https://github.com/teamkeel/keel/compare/v0.182.0...v0.183.0) (2022-10-27)


### Bug Fixes

* rename constraints to queryconstraints ([1d70356](https://github.com/teamkeel/keel/commit/1d70356e7ab85ca1f24f128c2b3e5f8b54d83c95))
* update index.d.ts ([a074a60](https://github.com/teamkeel/keel/commit/a074a60342d2995ff50513481b575306d3d995c8))
* update index.d.ts ([797d9d3](https://github.com/teamkeel/keel/commit/797d9d30f0d06ab5767989fff3c0c438457b54ab))


### Features

* drop slonik ([d89f4b1](https://github.com/teamkeel/keel/commit/d89f4b1fb1c066f6e03c6128bd3f8e4785106f28))

# [0.182.0](https://github.com/teamkeel/keel/compare/v0.181.0...v0.182.0) (2022-10-26)


### Bug Fixes

* postgres on correct ci job ([adf3c88](https://github.com/teamkeel/keel/commit/adf3c88714fd409189c0c43cc2ed790ebcf156f6))


### Features

* test sdk queries ([9d20961](https://github.com/teamkeel/keel/commit/9d20961967a187ee2f3d3014d5416b284d007f7d))

# [0.181.0](https://github.com/teamkeel/keel/compare/v0.180.0...v0.181.0) (2022-10-25)


### Features

* expressions to sql ([#499](https://github.com/teamkeel/keel/issues/499)) ([7ffceb6](https://github.com/teamkeel/keel/commit/7ffceb61a4391c58d1641016be27fec804039129))

# [0.180.0](https://github.com/teamkeel/keel/compare/v0.179.1...v0.180.0) (2022-10-24)


### Features

* crude funcitons support for rpc ([7f92281](https://github.com/teamkeel/keel/commit/7f92281952651368971dbf04eb814ca938c6443a))

## [0.179.1](https://github.com/teamkeel/keel/compare/v0.179.0...v0.179.1) (2022-10-24)


### Bug Fixes

* always start imports with a dot ([fdc6051](https://github.com/teamkeel/keel/commit/fdc60515c81f35d17070ad26ac98a1ebf1f9dadb))

# [0.179.0](https://github.com/teamkeel/keel/compare/v0.178.0...v0.179.0) (2022-10-22)


### Features

* release packages ([#495](https://github.com/teamkeel/keel/issues/495)) ([6896701](https://github.com/teamkeel/keel/commit/68967012f55b9cb0e92c025f612d6f528bfa77da))

# [0.178.0](https://github.com/teamkeel/keel/compare/v0.177.1...v0.178.0) (2022-10-22)


### Features

* convert js dates/timestamps to native go dates ([#493](https://github.com/teamkeel/keel/issues/493)) ([81bec8c](https://github.com/teamkeel/keel/commit/81bec8ca2c7ab82b30990f996d208d359cdaf526))

## [0.177.1](https://github.com/teamkeel/keel/compare/v0.177.0...v0.177.1) (2022-10-19)


### Bug Fixes

* fix graphql action response ([#488](https://github.com/teamkeel/keel/issues/488)) ([f7d6570](https://github.com/teamkeel/keel/commit/f7d6570db1f0ab02178b2d27411f21d1d30dea0f))

# [0.177.0](https://github.com/teamkeel/keel/compare/v0.176.0...v0.177.0) (2022-10-18)


### Features

* disable sdk query logging ([#481](https://github.com/teamkeel/keel/issues/481)) ([8ff5fd1](https://github.com/teamkeel/keel/commit/8ff5fd1be196dc0e60194743655c3cb1cd782374))

# [0.176.0](https://github.com/teamkeel/keel/compare/v0.175.1...v0.176.0) (2022-10-18)


### Features

* provide ability to silence testing package logging ([#479](https://github.com/teamkeel/keel/issues/479)) ([7e41834](https://github.com/teamkeel/keel/commit/7e41834ca87d4daedbbdd6322640f4c0dacb990b))

## [0.175.1](https://github.com/teamkeel/keel/compare/v0.175.0...v0.175.1) (2022-10-18)


### Bug Fixes

* test cmd ([#476](https://github.com/teamkeel/keel/issues/476)) ([937171a](https://github.com/teamkeel/keel/commit/937171afaf7fff2f3bba30a108b2c0e3d0f64e47))

# [0.175.0](https://github.com/teamkeel/keel/compare/v0.174.0...v0.175.0) (2022-10-18)


### Features

* move database reset to before each individual test ([#475](https://github.com/teamkeel/keel/issues/475)) ([5cbeb02](https://github.com/teamkeel/keel/commit/5cbeb02b93e0689791f45b9cc34a7d38bb1aaf4e))

# [0.174.0](https://github.com/teamkeel/keel/compare/v0.173.0...v0.174.0) (2022-10-18)


### Features

* add logging on HttpFunctionsClient response errors ([3066c09](https://github.com/teamkeel/keel/commit/3066c092648d4f9d6212fc8ba9ef6c2324807e14))

# [0.173.0](https://github.com/teamkeel/keel/compare/v0.172.0...v0.173.0) (2022-10-17)


### Bug Fixes

* fix wasm build ([3b05f19](https://github.com/teamkeel/keel/commit/3b05f198a4325ad849dd485d6192a3c273515490))


### Features

* simple rpc support ([fa22ea8](https://github.com/teamkeel/keel/commit/fa22ea80331e9966fcc6550648a35e4be5d1aa2d))

# [0.172.0](https://github.com/teamkeel/keel/compare/v0.171.0...v0.172.0) (2022-10-17)


### Features

* logging and gzip ([c35b3f3](https://github.com/teamkeel/keel/commit/c35b3f39a33896e0de9bbe78c76cae97ba13ebb7))

# [0.171.0](https://github.com/teamkeel/keel/compare/v0.170.0...v0.171.0) (2022-10-16)


### Features

* fix incorrect formatting of explicit inputs ([#469](https://github.com/teamkeel/keel/issues/469)) ([5073e42](https://github.com/teamkeel/keel/commit/5073e42f8f0136d1029ae5d64da212e1bc67d8fc))

# [0.170.0](https://github.com/teamkeel/keel/compare/v0.169.0...v0.170.0) (2022-10-14)


### Features

* support inputs in permission attributes ([#467](https://github.com/teamkeel/keel/issues/467)) ([5b733c3](https://github.com/teamkeel/keel/commit/5b733c39d27f2b9d4a36753d827eb59da500da4a))

# [0.169.0](https://github.com/teamkeel/keel/compare/v0.168.0...v0.169.0) (2022-10-12)


### Features

* add log on process exit ([5deb247](https://github.com/teamkeel/keel/commit/5deb24741c72aadb026c82992479a651b183aa12))

# [0.168.0](https://github.com/teamkeel/keel/compare/v0.167.0...v0.168.0) (2022-10-11)


### Features

* using implicit and explicit inputs in [@set](https://github.com/set) on update operation ([#463](https://github.com/teamkeel/keel/issues/463)) ([cdc6c5e](https://github.com/teamkeel/keel/commit/cdc6c5ed8728dcaaa9ce1ee14aba0fd4cacd207b))

# [0.167.0](https://github.com/teamkeel/keel/compare/v0.166.1...v0.167.0) (2022-10-11)


### Features

* setting optional enum to null in [@set](https://github.com/set) ([#461](https://github.com/teamkeel/keel/issues/461)) ([3a2cca8](https://github.com/teamkeel/keel/commit/3a2cca8f6fd5fa8bc4e299b449a3af7047fe9cc3))

## [0.166.1](https://github.com/teamkeel/keel/compare/v0.166.0...v0.166.1) (2022-10-11)


### Bug Fixes

* apply migrations if they don't have model field changes ([442ee6f](https://github.com/teamkeel/keel/commit/442ee6f6d843921b00f9ac62e574b3ce540ed5be))

# [0.166.0](https://github.com/teamkeel/keel/compare/v0.165.0...v0.166.0) (2022-10-11)


### Features

* set on create() for implicit and explicit inputs ([#459](https://github.com/teamkeel/keel/issues/459)) ([b04be1a](https://github.com/teamkeel/keel/commit/b04be1afdace0eedded6c7c8eef32d28e6c20efb))

# [0.165.0](https://github.com/teamkeel/keel/compare/v0.164.0...v0.165.0) (2022-10-10)


### Features

* validate implicit/explicit input clashes ([#458](https://github.com/teamkeel/keel/issues/458)) ([befd825](https://github.com/teamkeel/keel/commit/befd8252fed61e03920edef1bbbbb46fa1964ad6))

# [0.164.0](https://github.com/teamkeel/keel/compare/v0.163.0...v0.164.0) (2022-10-10)


### Features

* ctx.isAuthenticated support in set() and permission() ([#450](https://github.com/teamkeel/keel/issues/450)) ([48c8ef8](https://github.com/teamkeel/keel/commit/48c8ef8e671ec63fdfe0f4de4ddd1ccbd4dc0374))

# [0.163.0](https://github.com/teamkeel/keel/compare/v0.162.1...v0.163.0) (2022-10-10)


### Features

* support https and optional port on HttpFunctionsClient ([acf0eb8](https://github.com/teamkeel/keel/commit/acf0eb8b7646e2c15ed8e4f6392518d67107abc9))

## [0.162.1](https://github.com/teamkeel/keel/compare/v0.162.0...v0.162.1) (2022-10-07)


### Bug Fixes

* fix increment/decrement ([#456](https://github.com/teamkeel/keel/issues/456)) ([93fd4c3](https://github.com/teamkeel/keel/commit/93fd4c32965f7b83744c44a113663ec944742d60))

# [0.162.0](https://github.com/teamkeel/keel/compare/v0.161.0...v0.162.0) (2022-10-07)


### Features

* support comparing null literals with equality operators ([#455](https://github.com/teamkeel/keel/issues/455)) ([9d47e9d](https://github.com/teamkeel/keel/commit/9d47e9d56de8b95e7f7f8563ebd78d5cf9a69951))

# [0.161.0](https://github.com/teamkeel/keel/compare/v0.160.0...v0.161.0) (2022-10-07)


### Features

* allow comparison of matching field types (associations) ([#454](https://github.com/teamkeel/keel/issues/454)) ([7c7ce8e](https://github.com/teamkeel/keel/commit/7c7ce8e6579cfb214c10902adc5cf223da506b5d))

# [0.160.0](https://github.com/teamkeel/keel/compare/v0.159.1...v0.160.0) (2022-10-07)


### Features

* allow valid expression enum field comparison ([#453](https://github.com/teamkeel/keel/issues/453)) ([b428674](https://github.com/teamkeel/keel/commit/b42867489cab26caff9235a816fffda52a76c178))

## [0.159.1](https://github.com/teamkeel/keel/compare/v0.159.0...v0.159.1) (2022-10-07)


### Bug Fixes

* skip bootstrapping functions runtime if not used ([#451](https://github.com/teamkeel/keel/issues/451)) ([83a820a](https://github.com/teamkeel/keel/commit/83a820a806e48c1d47bee6c223d7119b5a69a85b))

# [0.159.0](https://github.com/teamkeel/keel/compare/v0.158.0...v0.159.0) (2022-10-06)


### Features

* initial enum support on [@set](https://github.com/set)() and permission() ([#448](https://github.com/teamkeel/keel/issues/448)) ([4870cf5](https://github.com/teamkeel/keel/commit/4870cf5f73b57804d82f853ed60f6a41fd7fb701))

# [0.158.0](https://github.com/teamkeel/keel/compare/v0.157.0...v0.158.0) (2022-10-05)


### Features

* make path to functions dir configurable ([d119bea](https://github.com/teamkeel/keel/commit/d119bead633a4830d96bccd04f26304e8189e420))

# [0.157.0](https://github.com/teamkeel/keel/compare/v0.156.0...v0.157.0) (2022-10-04)


### Features

* permission() support on update ([#439](https://github.com/teamkeel/keel/issues/439)) ([ab34e93](https://github.com/teamkeel/keel/commit/ab34e93c552623764814bfc3ecfbecaced004242))

# [0.156.0](https://github.com/teamkeel/keel/compare/v0.155.0...v0.156.0) (2022-10-04)


### Features

* support set expressions in update action ([#432](https://github.com/teamkeel/keel/issues/432)) ([17d071a](https://github.com/teamkeel/keel/commit/17d071ad9c564d6dd62c047440b6134ed08779d4))

# [0.155.0](https://github.com/teamkeel/keel/compare/v0.154.0...v0.155.0) (2022-10-03)


### Features

* better error capturing in test framework ([#435](https://github.com/teamkeel/keel/issues/435)) ([5188e18](https://github.com/teamkeel/keel/commit/5188e189efd4872cf8e119fae401bba46e614a45))

# [0.154.0](https://github.com/teamkeel/keel/compare/v0.153.0...v0.154.0) (2022-10-03)


### Features

* better js error capturing to reporter ([#433](https://github.com/teamkeel/keel/issues/433)) ([f5ef5af](https://github.com/teamkeel/keel/commit/f5ef5affc88ac5eac2f6b6746859b7f8e5483da3))

# [0.153.0](https://github.com/teamkeel/keel/compare/v0.152.0...v0.153.0) (2022-10-03)


### Features

* negative number operand ([#431](https://github.com/teamkeel/keel/issues/431)) ([357c03b](https://github.com/teamkeel/keel/commit/357c03b87a13a2f5d980c899f5e455a362eca9a1))

# [0.152.0](https://github.com/teamkeel/keel/compare/v0.151.0...v0.152.0) (2022-10-03)


### Features

* adds notToHaveAuthorizationError and notToHaveError ([#429](https://github.com/teamkeel/keel/issues/429)) ([1a0f557](https://github.com/teamkeel/keel/commit/1a0f557beeeae1900d32817a422d7e231701c028))

# [0.151.0](https://github.com/teamkeel/keel/compare/v0.150.0...v0.151.0) (2022-10-03)


### Features

* `toHaveAuthorizationError` matcher ([#427](https://github.com/teamkeel/keel/issues/427)) ([0a864b0](https://github.com/teamkeel/keel/commit/0a864b014315a91863432d9125feef0e172b183c))

# [0.150.0](https://github.com/teamkeel/keel/compare/v0.149.0...v0.150.0) (2022-10-03)


### Features

* testing package type fixes ([#425](https://github.com/teamkeel/keel/issues/425)) ([f25c9c0](https://github.com/teamkeel/keel/commit/f25c9c06e8f9c3d0c8096224b7dc88543bc0af3b))

# [0.149.0](https://github.com/teamkeel/keel/compare/v0.148.0...v0.149.0) (2022-10-03)


### Features

* more expectation types ([#416](https://github.com/teamkeel/keel/issues/416)) ([b7ebdef](https://github.com/teamkeel/keel/commit/b7ebdeff57fa3948c1b13297d28ea773e73cdda0))

# [0.148.0](https://github.com/teamkeel/keel/compare/v0.147.2...v0.148.0) (2022-10-02)


### Features

* test framework withIdentity() improvements ([#421](https://github.com/teamkeel/keel/issues/421)) ([4e3a1a3](https://github.com/teamkeel/keel/commit/4e3a1a311fd0f1667b6fcad0be986cc2412844f8))

## [0.147.2](https://github.com/teamkeel/keel/compare/v0.147.1...v0.147.2) (2022-10-02)


### Reverts

* Revert "fix: fix mismatching types (#422)" (#423) ([0126592](https://github.com/teamkeel/keel/commit/01265920c79bbeeca073fccfcff05b694ac50db6)), closes [#422](https://github.com/teamkeel/keel/issues/422) [#423](https://github.com/teamkeel/keel/issues/423)

## [0.147.1](https://github.com/teamkeel/keel/compare/v0.147.0...v0.147.1) (2022-10-02)


### Bug Fixes

* fix mismatching types ([#422](https://github.com/teamkeel/keel/issues/422)) ([ba65f09](https://github.com/teamkeel/keel/commit/ba65f096282ddb8a98d70c8f61ff3482760c3b86))

# [0.147.0](https://github.com/teamkeel/keel/compare/v0.146.0...v0.147.0) (2022-10-02)


### Features

* upgrading sdk withIdentity(model) ([#420](https://github.com/teamkeel/keel/issues/420)) ([eff0eb5](https://github.com/teamkeel/keel/commit/eff0eb581b4e3c59270520c2f97524c54d27e547))

# [0.146.0](https://github.com/teamkeel/keel/compare/v0.145.1...v0.146.0) (2022-09-30)


### Features

* move cors to runtime ([feafb8d](https://github.com/teamkeel/keel/commit/feafb8d96580ac0c2e14fcf1212d2f9c4f537e66))

## [0.145.1](https://github.com/teamkeel/keel/compare/v0.145.0...v0.145.1) (2022-09-30)


### Bug Fixes

* bump to rerun tests ([5397469](https://github.com/teamkeel/keel/commit/53974693159d24f1c7dd4b461fda85fbf5843d99))

# [0.145.0](https://github.com/teamkeel/keel/compare/v0.144.0...v0.145.0) (2022-09-30)


### Features

* permissions on delete() and integration tests ([#415](https://github.com/teamkeel/keel/issues/415)) ([8170b6e](https://github.com/teamkeel/keel/commit/8170b6e617b245463e59fafcd1e8842e06ab6fc6))

# [0.144.0](https://github.com/teamkeel/keel/compare/v0.143.0...v0.144.0) (2022-09-30)


### Features

* local cors ([d93db30](https://github.com/teamkeel/keel/commit/d93db30fae3686022352584edb5973d4efbd8f7d))

# [0.143.0](https://github.com/teamkeel/keel/compare/v0.142.0...v0.143.0) (2022-09-30)


### Features

* delete and list errors added in sdk ([#414](https://github.com/teamkeel/keel/issues/414)) ([cce1e9d](https://github.com/teamkeel/keel/commit/cce1e9dd4e3244bc6d1c4f5ffa818e2a3c772e17))

# [0.142.0](https://github.com/teamkeel/keel/compare/v0.141.1...v0.142.0) (2022-09-30)


### Features

* permission support on create() ([#408](https://github.com/teamkeel/keel/issues/408)) ([a84a183](https://github.com/teamkeel/keel/commit/a84a18308111ea6f90c805757e9d358f9a208b77))

## [0.141.1](https://github.com/teamkeel/keel/compare/v0.141.0...v0.141.1) (2022-09-29)


### Bug Fixes

* replace date formatting library ([#412](https://github.com/teamkeel/keel/issues/412)) ([26a1b05](https://github.com/teamkeel/keel/commit/26a1b052c05c51a81c327f21e3c5dfedd4d2374f))

# [0.141.0](https://github.com/teamkeel/keel/compare/v0.140.0...v0.141.0) (2022-09-29)


### Features

* add date fields to timestamp ([#411](https://github.com/teamkeel/keel/issues/411)) ([6131163](https://github.com/teamkeel/keel/commit/6131163e2eb101c55451e477cb11adba433c98fb))

# [0.140.0](https://github.com/teamkeel/keel/compare/v0.139.0...v0.140.0) (2022-09-29)


### Features

* fromNow ([#410](https://github.com/teamkeel/keel/issues/410)) ([56172e1](https://github.com/teamkeel/keel/commit/56172e12ffdf675f9beef3e5b1d4171b4682090e))

# [0.139.0](https://github.com/teamkeel/keel/compare/v0.138.0...v0.139.0) (2022-09-29)


### Features

* separate codegen entrypoint from its args ([8621396](https://github.com/teamkeel/keel/commit/8621396167053fa99ce236699c317dd32193aef8))

# [0.138.0](https://github.com/teamkeel/keel/compare/v0.137.0...v0.138.0) (2022-09-29)


### Features

* support formatted field for dates and timestamps ([#407](https://github.com/teamkeel/keel/issues/407)) ([143acd8](https://github.com/teamkeel/keel/commit/143acd8b9bde03b2041ff88ee198f267e4a102ac))

# [0.137.0](https://github.com/teamkeel/keel/compare/v0.136.0...v0.137.0) (2022-09-29)


### Features

* crude withIdentity(string) support in test framework ([#403](https://github.com/teamkeel/keel/issues/403)) ([5da1886](https://github.com/teamkeel/keel/commit/5da18868412e050db3011fb89f608bae6c4282ff))

# [0.136.0](https://github.com/teamkeel/keel/compare/v0.135.0...v0.136.0) (2022-09-29)


### Features

* implement gql date type sub fields ([#406](https://github.com/teamkeel/keel/issues/406)) ([113d308](https://github.com/teamkeel/keel/commit/113d308949c2416c1c1bcc9c60224cf3068a050e))

# [0.135.0](https://github.com/teamkeel/keel/compare/v0.134.0...v0.135.0) (2022-09-29)


### Features

* graphql timestamp implementation ([#404](https://github.com/teamkeel/keel/issues/404)) ([88ddee7](https://github.com/teamkeel/keel/commit/88ddee71d1c6519f39955667d3bdc693e3d51ae1))

# [0.134.0](https://github.com/teamkeel/keel/compare/v0.133.1...v0.134.0) (2022-09-28)


### Features

* optional list inputs ([4baee43](https://github.com/teamkeel/keel/commit/4baee43a366c5845f099e1dc89a210258ac65101))

## [0.133.1](https://github.com/teamkeel/keel/compare/v0.133.0...v0.133.1) (2022-09-28)


### Bug Fixes

* run command ([#402](https://github.com/teamkeel/keel/issues/402)) ([1bd87c1](https://github.com/teamkeel/keel/commit/1bd87c162c6ba38f2669e8c6e0c7915c1e2bb355))

# [0.133.0](https://github.com/teamkeel/keel/compare/v0.132.0...v0.133.0) (2022-09-28)


### Features

* identity support in testing library ([#401](https://github.com/teamkeel/keel/issues/401)) ([40c8d74](https://github.com/teamkeel/keel/commit/40c8d746762371782cf0dcdc9dfcd23de566549d))

# [0.132.0](https://github.com/teamkeel/keel/compare/v0.131.0...v0.132.0) (2022-09-28)


### Features

* integration test fix ([#397](https://github.com/teamkeel/keel/issues/397)) ([76a9bdb](https://github.com/teamkeel/keel/commit/76a9bdbe968ac410bf447607cb142263055f99bb))

# [0.131.0](https://github.com/teamkeel/keel/compare/v0.130.0...v0.131.0) (2022-09-28)


### Features

* added authenticate action to test framework ([#382](https://github.com/teamkeel/keel/issues/382)) ([b37a19f](https://github.com/teamkeel/keel/commit/b37a19f6ead4371dea0cb409ffc910bd6bd6bef7))

# [0.130.0](https://github.com/teamkeel/keel/compare/v0.129.0...v0.130.0) (2022-09-27)


### Features

* reset db ([#393](https://github.com/teamkeel/keel/issues/393)) ([19a16a4](https://github.com/teamkeel/keel/commit/19a16a490526ef32ae278b3ea43ce8f9f4591f08))

# [0.129.0](https://github.com/teamkeel/keel/compare/v0.128.0...v0.129.0) (2022-09-27)


### Features

* remove logs ([#392](https://github.com/teamkeel/keel/issues/392)) ([25220eb](https://github.com/teamkeel/keel/commit/25220eb1f9d258754259a04d8614528b766393d2))

# [0.128.0](https://github.com/teamkeel/keel/compare/v0.127.0...v0.128.0) (2022-09-27)


### Features

* catch uncaught exceptions whilst resetting db ([#391](https://github.com/teamkeel/keel/issues/391)) ([110cd87](https://github.com/teamkeel/keel/commit/110cd87037cda20659b01d81bfb1cf208ed0e555))

# [0.127.0](https://github.com/teamkeel/keel/compare/v0.126.0...v0.127.0) (2022-09-27)


### Features

* log reset db events ([#390](https://github.com/teamkeel/keel/issues/390)) ([0269240](https://github.com/teamkeel/keel/commit/0269240074b895613e127b638491cfc05b5cd759))

# [0.126.0](https://github.com/teamkeel/keel/compare/v0.125.0...v0.126.0) (2022-09-27)


### Features

* catch pg-protocol error ([#389](https://github.com/teamkeel/keel/issues/389)) ([fd7c10e](https://github.com/teamkeel/keel/commit/fd7c10e1b302d46655b35d589c5759801cd0b1c7))

# [0.125.0](https://github.com/teamkeel/keel/compare/v0.124.0...v0.125.0) (2022-09-27)


### Features

* handle backend termination ([#388](https://github.com/teamkeel/keel/issues/388)) ([b070429](https://github.com/teamkeel/keel/commit/b070429918dd5b0892253c30c186e8e169cb8515))

# [0.124.0](https://github.com/teamkeel/keel/compare/v0.123.0...v0.124.0) (2022-09-27)


### Features

* reset db after each test run ([#387](https://github.com/teamkeel/keel/issues/387)) ([e9db86f](https://github.com/teamkeel/keel/commit/e9db86f1ef2c3e8cf4cf36550761587b0c8fff23))

# [0.123.0](https://github.com/teamkeel/keel/compare/v0.122.0...v0.123.0) (2022-09-27)


### Features

* clear db ([#386](https://github.com/teamkeel/keel/issues/386)) ([30535ac](https://github.com/teamkeel/keel/commit/30535ac3e104b6f0ce03011102b43a78f1ec0b2d))

# [0.122.0](https://github.com/teamkeel/keel/compare/v0.121.0...v0.122.0) (2022-09-27)


### Features

* fixed auth response in sdk ([#385](https://github.com/teamkeel/keel/issues/385)) ([2ab7902](https://github.com/teamkeel/keel/commit/2ab79021181bb1d00039418d32e92c7576181fc7))

# [0.121.0](https://github.com/teamkeel/keel/compare/v0.120.0...v0.121.0) (2022-09-27)


### Features

* update action ([#381](https://github.com/teamkeel/keel/issues/381)) ([0d70158](https://github.com/teamkeel/keel/commit/0d70158fd59d412230ace17d044127fc3abf6792))

# [0.120.0](https://github.com/teamkeel/keel/compare/v0.119.0...v0.120.0) (2022-09-27)


### Features

* added authenticate response to runtime sdk ([#383](https://github.com/teamkeel/keel/issues/383)) ([41dd016](https://github.com/teamkeel/keel/commit/41dd0164b2f22fbd17ba349487ea39ea0cf4ff0c))

# [0.119.0](https://github.com/teamkeel/keel/compare/v0.118.0...v0.119.0) (2022-09-27)


### Features

* adds support for built-in delete action ([#380](https://github.com/teamkeel/keel/issues/380)) ([1624b89](https://github.com/teamkeel/keel/commit/1624b8943815249c111d175857862cc4bdf26ab6))

# [0.118.0](https://github.com/teamkeel/keel/compare/v0.117.0...v0.118.0) (2022-09-26)


### Features

* permission attribute support on the model definition ([#378](https://github.com/teamkeel/keel/issues/378)) ([8e8fb5e](https://github.com/teamkeel/keel/commit/8e8fb5ec4ba179ef6e0c2eff537596e1bd9de125))

# [0.117.0](https://github.com/teamkeel/keel/compare/v0.116.0...v0.117.0) (2022-09-26)


### Features

* init permissions support for string, bool, int, identity on create() ([#371](https://github.com/teamkeel/keel/issues/371)) ([6182ec1](https://github.com/teamkeel/keel/commit/6182ec190c5e582eef6d537fcd6ee86aaf1c289f))

# [0.116.0](https://github.com/teamkeel/keel/compare/v0.115.0...v0.116.0) (2022-09-23)


### Bug Fixes

* delete unused var ([237e24d](https://github.com/teamkeel/keel/commit/237e24d0679a07e39c03de0e9c3b9bc0e227e7f4))
* wasm main compilation ([#373](https://github.com/teamkeel/keel/issues/373)) ([abacb9e](https://github.com/teamkeel/keel/commit/abacb9e464817112b20ce5075fec58350a78070b))


### Features

* runtime request use path as a param ([879b14d](https://github.com/teamkeel/keel/commit/879b14db7437bad0ca42cf2329ba9b63a8e07547))

# [0.115.0](https://github.com/teamkeel/keel/compare/v0.114.0...v0.115.0) (2022-09-13)


### Features

* basic support for [@set](https://github.com/set) on create() ([#369](https://github.com/teamkeel/keel/issues/369)) ([d6f99a8](https://github.com/teamkeel/keel/commit/d6f99a8a553d99b48ae083c5001a339f90db4446))

# [0.114.0](https://github.com/teamkeel/keel/compare/v0.113.0...v0.114.0) (2022-09-09)


### Features

* list input types ([ebcfe46](https://github.com/teamkeel/keel/commit/ebcfe46710cd89acf41765e1469fc0fa2bc5fe17))

# [0.113.0](https://github.com/teamkeel/keel/compare/v0.112.2...v0.113.0) (2022-09-07)


### Features

* testing package output ([#365](https://github.com/teamkeel/keel/issues/365)) ([6d5cf6d](https://github.com/teamkeel/keel/commit/6d5cf6d6370db900cbb1a0d2060c7ebb7fd1dc4a))

## [0.112.2](https://github.com/teamkeel/keel/compare/v0.112.1...v0.112.2) (2022-09-07)


### Bug Fixes

* test output ([#362](https://github.com/teamkeel/keel/issues/362)) ([7dfe5a6](https://github.com/teamkeel/keel/commit/7dfe5a6c0e36a9d1a7537320b95691f706a339c0))

## [0.112.1](https://github.com/teamkeel/keel/compare/v0.112.0...v0.112.1) (2022-09-07)


### Bug Fixes

* prettier ([#363](https://github.com/teamkeel/keel/issues/363)) ([74d8785](https://github.com/teamkeel/keel/commit/74d87852dd188721cef9f0ab141e221ae8b94b62))

# [0.112.0](https://github.com/teamkeel/keel/compare/v0.111.6...v0.112.0) (2022-09-07)


### Features

* verify jwt and set context with identity id ([#350](https://github.com/teamkeel/keel/issues/350)) ([32a7d1b](https://github.com/teamkeel/keel/commit/32a7d1b4693a0e0e7c177fe9c28f9801e1d1305d))

## [0.111.6](https://github.com/teamkeel/keel/compare/v0.111.5...v0.111.6) (2022-09-07)


### Bug Fixes

* test exclusion ([#361](https://github.com/teamkeel/keel/issues/361)) ([68271e9](https://github.com/teamkeel/keel/commit/68271e9a4cc7e5eba6586cf604ab82011cf8f1fc))

## [0.111.5](https://github.com/teamkeel/keel/compare/v0.111.4...v0.111.5) (2022-09-07)


### Bug Fixes

* test pkg cli output ([#360](https://github.com/teamkeel/keel/issues/360)) ([a694343](https://github.com/teamkeel/keel/commit/a694343a50b333557244758564154cc7c537bf3a))

## [0.111.4](https://github.com/teamkeel/keel/compare/v0.111.3...v0.111.4) (2022-09-07)


### Bug Fixes

* update sdk in testing package ([#359](https://github.com/teamkeel/keel/issues/359)) ([87f569a](https://github.com/teamkeel/keel/commit/87f569a9c40cdec8d919baef987285e9aad6b946))

## [0.111.3](https://github.com/teamkeel/keel/compare/v0.111.2...v0.111.3) (2022-09-07)


### Bug Fixes

* downgrade to chalk v4 to avoid esm issues ([#358](https://github.com/teamkeel/keel/issues/358)) ([5628779](https://github.com/teamkeel/keel/commit/56287791cb74bd740bbb72845445a3fac0d5ecc6))

## [0.111.2](https://github.com/teamkeel/keel/compare/v0.111.1...v0.111.2) (2022-09-07)


### Bug Fixes

* test result output ([#357](https://github.com/teamkeel/keel/issues/357)) ([d92217f](https://github.com/teamkeel/keel/commit/d92217f943bc61815fff01ff91944a9a4f868b1e))

## [0.111.1](https://github.com/teamkeel/keel/compare/v0.111.0...v0.111.1) (2022-09-07)


### Bug Fixes

* expose success log level in logger package ([#356](https://github.com/teamkeel/keel/issues/356)) ([1f484e8](https://github.com/teamkeel/keel/commit/1f484e8411ca8b071b574185341c4093ecb8ebda))

# [0.111.0](https://github.com/teamkeel/keel/compare/v0.110.0...v0.111.0) (2022-09-07)


### Features

* hook up pattern flag to integration tests ([#355](https://github.com/teamkeel/keel/issues/355)) ([2ad0faa](https://github.com/teamkeel/keel/commit/2ad0faa8aef154eb8f064bd77968e9e0dc470038))

# [0.110.0](https://github.com/teamkeel/keel/compare/v0.109.0...v0.110.0) (2022-09-07)


### Features

* support isolation of tests via a pattern in testing package ([#354](https://github.com/teamkeel/keel/issues/354)) ([51877ef](https://github.com/teamkeel/keel/commit/51877ef1b6acc3b8a360718f51001ee1ad419e39))

# [0.109.0](https://github.com/teamkeel/keel/compare/v0.108.1...v0.109.0) (2022-09-06)


### Features

* hashing password secret data type ([#348](https://github.com/teamkeel/keel/issues/348)) ([dad5be9](https://github.com/teamkeel/keel/commit/dad5be92fdfd5046b199b7445ba03175171b333d))

## [0.108.1](https://github.com/teamkeel/keel/compare/v0.108.0...v0.108.1) (2022-09-06)


### Bug Fixes

* failed authenticate() response from gql ([#347](https://github.com/teamkeel/keel/issues/347)) ([ed175a0](https://github.com/teamkeel/keel/commit/ed175a0cacfa9e916f3a848fb8e86aa0470623f1))

# [0.108.0](https://github.com/teamkeel/keel/compare/v0.107.0...v0.108.0) (2022-09-02)


### Features

* adds faker into testing package ([#345](https://github.com/teamkeel/keel/issues/345)) ([da46e10](https://github.com/teamkeel/keel/commit/da46e10c5eac6b5c87c28e29f0d108d13936b5a9))

# [0.107.0](https://github.com/teamkeel/keel/compare/v0.106.7...v0.107.0) (2022-09-02)


### Features

* hashed secret data type ([#336](https://github.com/teamkeel/keel/issues/336)) ([599f188](https://github.com/teamkeel/keel/commit/599f1889761ac5f747a377b7ac8499daa1815844))

## [0.106.7](https://github.com/teamkeel/keel/compare/v0.106.6...v0.106.7) (2022-09-01)


### Bug Fixes

* integration tests  ([#344](https://github.com/teamkeel/keel/issues/344)) ([4e1dc43](https://github.com/teamkeel/keel/commit/4e1dc435fda17e13c50c45486f74b4975ec8dda1))

## [0.106.6](https://github.com/teamkeel/keel/compare/v0.106.5...v0.106.6) (2022-09-01)


### Bug Fixes

* integration tests ([#343](https://github.com/teamkeel/keel/issues/343)) ([5c6281c](https://github.com/teamkeel/keel/commit/5c6281cbdd0ef0a4d9f9ac7d019ecd42e6008723))

## [0.106.5](https://github.com/teamkeel/keel/compare/v0.106.4...v0.106.5) (2022-09-01)


### Bug Fixes

* action executor typings ([#342](https://github.com/teamkeel/keel/issues/342)) ([15cd224](https://github.com/teamkeel/keel/commit/15cd22480c8a4067fd09731ac025d2a11ce25c6e))

## [0.106.4](https://github.com/teamkeel/keel/compare/v0.106.3...v0.106.4) (2022-09-01)


### Bug Fixes

* update testing package to use return types from sdk  ([#341](https://github.com/teamkeel/keel/issues/341)) ([ae92ef5](https://github.com/teamkeel/keel/commit/ae92ef500868c54361cab1c89bd65a3131be68bc))

## [0.106.3](https://github.com/teamkeel/keel/compare/v0.106.2...v0.106.3) (2022-09-01)


### Bug Fixes

* ignore prettier on d.ts files ([#340](https://github.com/teamkeel/keel/issues/340)) ([1f9662f](https://github.com/teamkeel/keel/commit/1f9662f1e74727fe53e5646e941e38706f0720e2))

## [0.106.2](https://github.com/teamkeel/keel/compare/v0.106.1...v0.106.2) (2022-09-01)


### Bug Fixes

* correct sdk typings ([#339](https://github.com/teamkeel/keel/issues/339)) ([74c17cd](https://github.com/teamkeel/keel/commit/74c17cd9af9fc23f6c998b10fd9b705c390eef6d))

## [0.106.1](https://github.com/teamkeel/keel/compare/v0.106.0...v0.106.1) (2022-09-01)


### Bug Fixes

* revert zod changes ([#337](https://github.com/teamkeel/keel/issues/337)) ([12bae2d](https://github.com/teamkeel/keel/commit/12bae2d94c96366c6e86c2fd18538bb17839c0af))

# [0.106.0](https://github.com/teamkeel/keel/compare/v0.105.0...v0.106.0) (2022-09-01)


### Features

* adds zod for schema validation for database queries ([#335](https://github.com/teamkeel/keel/issues/335)) ([8aed8d1](https://github.com/teamkeel/keel/commit/8aed8d1e4c4802f08518647dab664c7b95de80c4))

# [0.105.0](https://github.com/teamkeel/keel/compare/v0.104.1...v0.105.0) (2022-08-30)


### Features

* built-in authenticate() operation with graphql support ([#328](https://github.com/teamkeel/keel/issues/328)) ([f7b76fa](https://github.com/teamkeel/keel/commit/f7b76fa3219652d804f6105d85347bcabbaed103))

## [0.104.1](https://github.com/teamkeel/keel/compare/v0.104.0...v0.104.1) (2022-08-30)


### Bug Fixes

* update sdk type defs to reflect new query api return types ([#330](https://github.com/teamkeel/keel/issues/330)) ([67dedb9](https://github.com/teamkeel/keel/commit/67dedb9263f7b0b3601df1c5ed7aff64b68ab719))

# [0.104.0](https://github.com/teamkeel/keel/compare/v0.103.0...v0.104.0) (2022-08-30)


### Features

* update codegenned wrapper func to use new return types from sdk ([#329](https://github.com/teamkeel/keel/issues/329)) ([8467c93](https://github.com/teamkeel/keel/commit/8467c9394b988abfccb8fdbeb172ae237a0dbbfd))

# [0.103.0](https://github.com/teamkeel/keel/compare/v0.102.21...v0.103.0) (2022-08-26)


### Features

* add custom function return type interfaces ([#320](https://github.com/teamkeel/keel/issues/320)) ([8aa8ddb](https://github.com/teamkeel/keel/commit/8aa8ddb529fc6021d0e479e786a2473adbcf90e2))

## [0.102.21](https://github.com/teamkeel/keel/compare/v0.102.20...v0.102.21) (2022-08-26)


### Bug Fixes

* await for loop to finish before moving to next test ([#324](https://github.com/teamkeel/keel/issues/324)) ([40cf8b2](https://github.com/teamkeel/keel/commit/40cf8b22ea8597878b90d0e3e3331b49ac7155d1))

## [0.102.20](https://github.com/teamkeel/keel/compare/v0.102.19...v0.102.20) (2022-08-26)


### Bug Fixes

* await for reporter to report before moving to next test ([#323](https://github.com/teamkeel/keel/issues/323)) ([43c6bc3](https://github.com/teamkeel/keel/commit/43c6bc3762c40e499cd624a7fd4e397a874cc0c0))

## [0.102.19](https://github.com/teamkeel/keel/compare/v0.102.18...v0.102.19) (2022-08-26)


### Bug Fixes

* custom function handling of unknown json types ([#322](https://github.com/teamkeel/keel/issues/322)) ([3c71d6b](https://github.com/teamkeel/keel/commit/3c71d6b2c96f5022cb5b8419c187c2cf3bc8c4fb))

## [0.102.18](https://github.com/teamkeel/keel/compare/v0.102.17...v0.102.18) (2022-08-25)


### Bug Fixes

* add tests for npm modules ([#321](https://github.com/teamkeel/keel/issues/321)) ([4b66dbd](https://github.com/teamkeel/keel/commit/4b66dbdf2d0cb002e26ff91f9aa2774aaaac24df))

## [0.102.17](https://github.com/teamkeel/keel/compare/v0.102.16...v0.102.17) (2022-08-25)


### Bug Fixes

* test ci ([#317](https://github.com/teamkeel/keel/issues/317)) ([c86e83b](https://github.com/teamkeel/keel/commit/c86e83b82e40e23384f3bfb62f609c297b917f1a))

## [0.102.16](https://github.com/teamkeel/keel/compare/v0.102.15...v0.102.16) (2022-08-25)


### Bug Fixes

* test ([#316](https://github.com/teamkeel/keel/issues/316)) ([742b5aa](https://github.com/teamkeel/keel/commit/742b5aa30ae81be60fcb9eb0d2f7625a3a135e95))

## [0.102.15](https://github.com/teamkeel/keel/compare/v0.102.14...v0.102.15) (2022-08-25)


### Bug Fixes

* something ([#315](https://github.com/teamkeel/keel/issues/315)) ([a087141](https://github.com/teamkeel/keel/commit/a08714132ac26b1738479ff6b1cee63e2d8efae5))

## [0.102.14](https://github.com/teamkeel/keel/compare/v0.102.13...v0.102.14) (2022-08-25)


### Bug Fixes

* use correct dir option ([#314](https://github.com/teamkeel/keel/issues/314)) ([cdf681d](https://github.com/teamkeel/keel/commit/cdf681d8293aad82639fb1cbf571c1035b49f5ba))

## [0.102.13](https://github.com/teamkeel/keel/compare/v0.102.12...v0.102.13) (2022-08-25)


### Bug Fixes

* cast check to pointer ([#312](https://github.com/teamkeel/keel/issues/312)) ([8a4ac29](https://github.com/teamkeel/keel/commit/8a4ac296c13a94077dc8d2cf82aa5332007f58ce))

## [0.102.12](https://github.com/teamkeel/keel/compare/v0.102.11...v0.102.12) (2022-08-25)


### Bug Fixes

* debug wasm errors ([#311](https://github.com/teamkeel/keel/issues/311)) ([4b738ea](https://github.com/teamkeel/keel/commit/4b738ea72cd9c6173f5f9956b69954d4bad2773e))

## [0.102.11](https://github.com/teamkeel/keel/compare/v0.102.10...v0.102.11) (2022-08-25)


### Bug Fixes

* cover unexpected errors returned from wasm binary in typescript wrapper ([#310](https://github.com/teamkeel/keel/issues/310)) ([ef91a4a](https://github.com/teamkeel/keel/commit/ef91a4a5b8e5f3702fc9ea8c2d2eb5f7db97e5b9))

## [0.102.10](https://github.com/teamkeel/keel/compare/v0.102.9...v0.102.10) (2022-08-24)


### Bug Fixes

* revert non working commit ([#309](https://github.com/teamkeel/keel/issues/309)) ([39aad8a](https://github.com/teamkeel/keel/commit/39aad8a1e247d45e32c728c87d255d2bc2455fd5))

## [0.102.6](https://github.com/teamkeel/keel/compare/v0.102.5...v0.102.6) (2022-08-24)


### Bug Fixes

* wrap col names in sql idents ([#303](https://github.com/teamkeel/keel/issues/303)) ([59d6fe6](https://github.com/teamkeel/keel/commit/59d6fe61f2085a2b8a4b0a1d81efe97c79c53d5a))

## [0.102.5](https://github.com/teamkeel/keel/compare/v0.102.4...v0.102.5) (2022-08-24)


### Bug Fixes

* change testing model api to account for async chaining ([#302](https://github.com/teamkeel/keel/issues/302)) ([ac0403c](https://github.com/teamkeel/keel/commit/ac0403ca5d2afd3d89670c81fbb66dd6c39b5c8c))

## [0.102.4](https://github.com/teamkeel/keel/compare/v0.102.3...v0.102.4) (2022-08-24)


### Bug Fixes

* log stack when catching err in tests ([#301](https://github.com/teamkeel/keel/issues/301)) ([a4720be](https://github.com/teamkeel/keel/commit/a4720bec5012500b859611a3dba96a64db42da37))

## [0.102.3](https://github.com/teamkeel/keel/compare/v0.102.2...v0.102.3) (2022-08-24)


### Bug Fixes

* change query class to delegate connection logic to methods in async context ([#300](https://github.com/teamkeel/keel/issues/300)) ([77de48f](https://github.com/teamkeel/keel/commit/77de48f78df8da76b6b246ca835a3024023bf0b1))

## [0.102.2](https://github.com/teamkeel/keel/compare/v0.102.1...v0.102.2) (2022-08-24)


### Bug Fixes

* export ChainableQuery from sdk package ([#299](https://github.com/teamkeel/keel/issues/299)) ([4d92dc0](https://github.com/teamkeel/keel/commit/4d92dc0bc7c78ec9150aa169b6d1dff1e40cb35d))

## [0.102.1](https://github.com/teamkeel/keel/compare/v0.102.0...v0.102.1) (2022-08-24)


### Bug Fixes

* adds date and enum constraints to generic constraint union type ([#298](https://github.com/teamkeel/keel/issues/298)) ([3cc3e36](https://github.com/teamkeel/keel/commit/3cc3e369464c2a0bff111e616fb041408cef1009))

# [0.102.0](https://github.com/teamkeel/keel/compare/v0.101.0...v0.102.0) (2022-08-23)


### Features

* support sql limit on non-unique hasOne ([#296](https://github.com/teamkeel/keel/issues/296)) ([2b3a7bb](https://github.com/teamkeel/keel/commit/2b3a7bb243fd47a0f858bf94b8e7018a227d3a11))

# [0.101.0](https://github.com/teamkeel/keel/compare/v0.100.3...v0.101.0) (2022-08-23)


### Features

* support date constraints ([#294](https://github.com/teamkeel/keel/issues/294)) ([e40a990](https://github.com/teamkeel/keel/commit/e40a990e8ecc3bea4082ac9856f852c1a0982168))

## [0.100.3](https://github.com/teamkeel/keel/compare/v0.100.2...v0.100.3) (2022-08-23)


### Bug Fixes

* use query syntax compatible with rds data api ([a549458](https://github.com/teamkeel/keel/commit/a549458dba67118ed4cf44d9ed7fc970cb8eafb3))

## [0.100.2](https://github.com/teamkeel/keel/compare/v0.100.1...v0.100.2) (2022-08-23)


### Bug Fixes

* use stdout for logger error calls ([#291](https://github.com/teamkeel/keel/issues/291)) ([4ce76da](https://github.com/teamkeel/keel/commit/4ce76dad934c99bb0956ff6f12e30db58b2be0e2))

## [0.100.1](https://github.com/teamkeel/keel/compare/v0.100.0...v0.100.1) (2022-08-23)


### Bug Fixes

* log unrelated errors in testing package ([#290](https://github.com/teamkeel/keel/issues/290)) ([65343c4](https://github.com/teamkeel/keel/commit/65343c40e71dd758e3c78ee40c11f6072df17167))

# [0.100.0](https://github.com/teamkeel/keel/compare/v0.99.1...v0.100.0) (2022-08-23)


### Features

* add identity to action executor ([#289](https://github.com/teamkeel/keel/issues/289)) ([fb3389e](https://github.com/teamkeel/keel/commit/fb3389e3304d70879ac62770b22a14ac52822bbb))

## [0.99.1](https://github.com/teamkeel/keel/compare/v0.99.0...v0.99.1) (2022-08-23)


### Bug Fixes

* incorrect import ([#288](https://github.com/teamkeel/keel/issues/288)) ([26c1136](https://github.com/teamkeel/keel/commit/26c1136fb7b56938337a0395618fa4ba1e060415))

# [0.99.0](https://github.com/teamkeel/keel/compare/v0.98.3...v0.99.0) (2022-08-23)


### Features

* adds Identity type to @teamkeel/sdk ([#287](https://github.com/teamkeel/keel/issues/287)) ([e2fceb9](https://github.com/teamkeel/keel/commit/e2fceb9b09d2d70451516fe755b659f30e18e3f6))

## [0.98.3](https://github.com/teamkeel/keel/compare/v0.98.2...v0.98.3) (2022-08-23)


### Bug Fixes

* schema validation issue in the todo example ([#285](https://github.com/teamkeel/keel/issues/285)) ([35f5adf](https://github.com/teamkeel/keel/commit/35f5adf6e8dcde234c39b5d52419d41a42159022))

## [0.98.2](https://github.com/teamkeel/keel/compare/v0.98.1...v0.98.2) (2022-08-22)


### Bug Fixes

* log output ([#286](https://github.com/teamkeel/keel/issues/286)) ([b8308c1](https://github.com/teamkeel/keel/commit/b8308c17ae4ac9e06e96cae9b96425b1d77805df))

## [0.98.1](https://github.com/teamkeel/keel/compare/v0.98.0...v0.98.1) (2022-08-22)


### Bug Fixes

* logger tweaks ([#284](https://github.com/teamkeel/keel/issues/284)) ([1136b52](https://github.com/teamkeel/keel/commit/1136b5215e35fcbdb236e794f125fb30cd49f5a8))

# [0.98.0](https://github.com/teamkeel/keel/compare/v0.97.1...v0.98.0) (2022-08-22)


### Features

* testing package logger ([#283](https://github.com/teamkeel/keel/issues/283)) ([a5e8b6a](https://github.com/teamkeel/keel/commit/a5e8b6ae26b0e7ec0a5ca72a8bcd9660a7b4ad7c))

## [0.97.1](https://github.com/teamkeel/keel/compare/v0.97.0...v0.97.1) (2022-08-20)


### Bug Fixes

* move generic param to method level ([#278](https://github.com/teamkeel/keel/issues/278)) ([2485a57](https://github.com/teamkeel/keel/commit/2485a57abf6fe28ed02ca14f178def8a59eb3864))

# [0.97.0](https://github.com/teamkeel/keel/compare/v0.96.2...v0.97.0) (2022-08-20)


### Features

* setup action codegen for testing package ([#277](https://github.com/teamkeel/keel/issues/277)) ([6be4ce3](https://github.com/teamkeel/keel/commit/6be4ce3124c0defcf361ac23a5d8398823155be4))

## [0.96.2](https://github.com/teamkeel/keel/compare/v0.96.1...v0.96.2) (2022-08-20)


### Bug Fixes

* debug logs ([#275](https://github.com/teamkeel/keel/issues/275)) ([261e439](https://github.com/teamkeel/keel/commit/261e439b81be20f608df895b8342ec6d1d377d2b))

## [0.96.1](https://github.com/teamkeel/keel/compare/v0.96.0...v0.96.1) (2022-08-20)


### Bug Fixes

* small tweaks to [@testing](https://github.com/testing) package ([#274](https://github.com/teamkeel/keel/issues/274)) ([d20926a](https://github.com/teamkeel/keel/commit/d20926acc84496ef8e5ed6247784fff59ef45759))

# [0.96.0](https://github.com/teamkeel/keel/compare/v0.95.10...v0.96.0) (2022-08-20)


### Features

* bump js package versions ([#273](https://github.com/teamkeel/keel/issues/273)) ([79033cd](https://github.com/teamkeel/keel/commit/79033cd6fb49bfc69e58bbb9f9f209e5181ac1ce))

## [0.95.10](https://github.com/teamkeel/keel/compare/v0.95.9...v0.95.10) (2022-08-20)


### Bug Fixes

* remove top level await attempts ([#272](https://github.com/teamkeel/keel/issues/272)) ([c76b4e9](https://github.com/teamkeel/keel/commit/c76b4e90b0c878d9b7ec669b88814f98b316a9d3))

## [0.95.9](https://github.com/teamkeel/keel/compare/v0.95.8...v0.95.9) (2022-08-20)


### Bug Fixes

* set type module ([#271](https://github.com/teamkeel/keel/issues/271)) ([0ebb5f4](https://github.com/teamkeel/keel/commit/0ebb5f4ee85120c7869757b37c161d98c736a5bf))

## [0.95.8](https://github.com/teamkeel/keel/compare/v0.95.7...v0.95.8) (2022-08-20)


### Bug Fixes

* update tsconfig to allow for top level await in generated module ([#270](https://github.com/teamkeel/keel/issues/270)) ([aecece3](https://github.com/teamkeel/keel/commit/aecece35ff0379032ba4da0e8b96fcf2c959820c))

## [0.95.7](https://github.com/teamkeel/keel/compare/v0.95.6...v0.95.7) (2022-08-19)


### Bug Fixes

* flaky test ([#269](https://github.com/teamkeel/keel/issues/269)) ([82d9dd3](https://github.com/teamkeel/keel/commit/82d9dd3812dc6bce13213545474fa8e874d308d0))

## [0.95.6](https://github.com/teamkeel/keel/compare/v0.95.5...v0.95.6) (2022-08-19)


### Bug Fixes

* setup and migrate database per test case ([#268](https://github.com/teamkeel/keel/issues/268)) ([298dc54](https://github.com/teamkeel/keel/commit/298dc546ae41d5d29e851144af382f2cb2850a04))

## [0.95.5](https://github.com/teamkeel/keel/compare/v0.95.4...v0.95.5) (2022-08-19)


### Bug Fixes

* await result of fn() in try/catch block ([#267](https://github.com/teamkeel/keel/issues/267)) ([d851876](https://github.com/teamkeel/keel/commit/d8518762d70d272dc353af47e7bccb7d2fa66cac))

## [0.95.4](https://github.com/teamkeel/keel/compare/v0.95.3...v0.95.4) (2022-08-19)


### Bug Fixes

* do not console.error as this prints to stderr and causes test to exit too soon ([#266](https://github.com/teamkeel/keel/issues/266)) ([2c8f32b](https://github.com/teamkeel/keel/commit/2c8f32bcf6e0056fadf0a63c22d127fa2442b5d5))

## [0.95.3](https://github.com/teamkeel/keel/compare/v0.95.2...v0.95.3) (2022-08-19)


### Bug Fixes

* small fixes ([#265](https://github.com/teamkeel/keel/issues/265)) ([df5dcb7](https://github.com/teamkeel/keel/commit/df5dcb7052a542e8bc1dd959c6f45dd88bfedcb0))

## [0.95.2](https://github.com/teamkeel/keel/compare/v0.95.1...v0.95.2) (2022-08-19)


### Bug Fixes

* small fix ([#264](https://github.com/teamkeel/keel/issues/264)) ([ffbb48e](https://github.com/teamkeel/keel/commit/ffbb48ea09de317aa8bd7b553409cee789d10ff4))

## [0.95.1](https://github.com/teamkeel/keel/compare/v0.95.0...v0.95.1) (2022-08-19)


### Bug Fixes

* log catching of assertion failures ([#263](https://github.com/teamkeel/keel/issues/263)) ([1f943b5](https://github.com/teamkeel/keel/commit/1f943b5519894193e41493557153bfaa264d8329))

# [0.95.0](https://github.com/teamkeel/keel/compare/v0.94.0...v0.95.0) (2022-08-19)


### Features

* generate test flavoured variants of @teamkeel/client ([#262](https://github.com/teamkeel/keel/issues/262)) ([dcf73d0](https://github.com/teamkeel/keel/commit/dcf73d031a059c2a8248fc1c8bce9080c44ea146))

# [0.94.0](https://github.com/teamkeel/keel/compare/v0.93.2...v0.94.0) (2022-08-17)


### Features

* bump all packages to next major ([#256](https://github.com/teamkeel/keel/issues/256)) ([cd32880](https://github.com/teamkeel/keel/commit/cd32880437e36b8ec48cfc9c0d45af11c2f614c1))

## [0.93.2](https://github.com/teamkeel/keel/compare/v0.93.1...v0.93.2) (2022-08-17)


### Bug Fixes

* use common js setup for testing package ([#255](https://github.com/teamkeel/keel/issues/255)) ([f0f0cc6](https://github.com/teamkeel/keel/commit/f0f0cc65fc498d242ea2d53766e670fcc0913efd))

## [0.93.1](https://github.com/teamkeel/keel/compare/v0.93.0...v0.93.1) (2022-08-17)


### Bug Fixes

* make testing package publish public ([#254](https://github.com/teamkeel/keel/issues/254)) ([373e621](https://github.com/teamkeel/keel/commit/373e6215039da471ac7cab21c85db85a21b5a0d5))

# [0.93.0](https://github.com/teamkeel/keel/compare/v0.92.5...v0.93.0) (2022-08-16)


### Features

* @teamkeel/testing package ([#252](https://github.com/teamkeel/keel/issues/252)) ([7e493a5](https://github.com/teamkeel/keel/commit/7e493a524df7e38981d33cdfb3e2d46dcc25203d))

## [0.92.5](https://github.com/teamkeel/keel/compare/v0.92.4...v0.92.5) (2022-08-15)


### Bug Fixes

* fix validate command ([aeb7fc1](https://github.com/teamkeel/keel/commit/aeb7fc1991d90937f454113217972751b6b8dc2b))

## [0.92.4](https://github.com/teamkeel/keel/compare/v0.92.3...v0.92.4) (2022-08-15)


### Bug Fixes

* haschanges was the wrong way around ([f8ab479](https://github.com/teamkeel/keel/commit/f8ab479d5646188d73c35612c073f7cd4d8b0717))

## [0.92.3](https://github.com/teamkeel/keel/compare/v0.92.2...v0.92.3) (2022-08-11)


### Bug Fixes

* do not use sql literal when building order by clauses ([#248](https://github.com/teamkeel/keel/issues/248)) ([f1cf3dd](https://github.com/teamkeel/keel/commit/f1cf3dd13432fc26dc031355e889337dca4799f6))

## [0.92.2](https://github.com/teamkeel/keel/compare/v0.92.1...v0.92.2) (2022-08-11)


### Bug Fixes

* pass order clauses to query builder ([#247](https://github.com/teamkeel/keel/issues/247)) ([eff6ee6](https://github.com/teamkeel/keel/commit/eff6ee615bfa8bba913d43a3212ebe16b2815ef6))

## [0.92.1](https://github.com/teamkeel/keel/compare/v0.92.0...v0.92.1) (2022-08-11)


### Bug Fixes

* order by fixes ([#246](https://github.com/teamkeel/keel/issues/246)) ([637c3e6](https://github.com/teamkeel/keel/commit/637c3e67b36b66557ed43a02774e246ff0166d90))

# [0.92.0](https://github.com/teamkeel/keel/compare/v0.91.3...v0.92.0) (2022-08-11)


### Features

* implement order by in sql api ([#245](https://github.com/teamkeel/keel/issues/245)) ([cc9b853](https://github.com/teamkeel/keel/commit/cc9b8533c9818df59170fd27076ee9190ccbe079))

## [0.91.3](https://github.com/teamkeel/keel/compare/v0.91.2...v0.91.3) (2022-08-11)


### Bug Fixes

* format sql values correctly based on their type ([#244](https://github.com/teamkeel/keel/issues/244)) ([d18316a](https://github.com/teamkeel/keel/commit/d18316a64d125165c7c7f9bfdc43dcf6da60c160))

## [0.91.2](https://github.com/teamkeel/keel/compare/v0.91.1...v0.91.2) (2022-08-11)


### Bug Fixes

* colorization and query output format ([#243](https://github.com/teamkeel/keel/issues/243)) ([53b9de2](https://github.com/teamkeel/keel/commit/53b9de2257000c575703f53491382e924ea26ba1))

## [0.91.1](https://github.com/teamkeel/keel/compare/v0.91.0...v0.91.1) (2022-08-11)


### Bug Fixes

* hook up logger - codegen & sdk logger typing fixes ([#242](https://github.com/teamkeel/keel/issues/242)) ([972b776](https://github.com/teamkeel/keel/commit/972b776710b032812203ded26a8f638a6add5cc6))

# [0.91.0](https://github.com/teamkeel/keel/compare/v0.90.7...v0.91.0) (2022-08-11)


### Features

* adds basic logger implementation to sdk ([#241](https://github.com/teamkeel/keel/issues/241)) ([84a9b93](https://github.com/teamkeel/keel/commit/84a9b93888685a73c1ae722f75560b2192f24879))

## [0.90.7](https://github.com/teamkeel/keel/compare/v0.90.6...v0.90.7) (2022-08-10)


### Bug Fixes

* small fixes for codegen-ed typings ([#240](https://github.com/teamkeel/keel/issues/240)) ([b212e68](https://github.com/teamkeel/keel/commit/b212e68ce02ba99f608fc40c43833f141b0998bc))

## [0.90.6](https://github.com/teamkeel/keel/compare/v0.90.5...v0.90.6) (2022-08-10)


### Bug Fixes

* fix ORM api & add query logging ([#238](https://github.com/teamkeel/keel/issues/238)) ([9727579](https://github.com/teamkeel/keel/commit/9727579508d960f9184ffc97b74bfd31406a2317))

## [0.90.5](https://github.com/teamkeel/keel/compare/v0.90.4...v0.90.5) (2022-08-10)


### Bug Fixes

* revert to TEXT column type for keel schema ([#237](https://github.com/teamkeel/keel/issues/237)) ([29c6303](https://github.com/teamkeel/keel/commit/29c630315132dfd1c83bc02c609ce99280ee7517))

## [0.90.4](https://github.com/teamkeel/keel/compare/v0.90.3...v0.90.4) (2022-08-09)


### Bug Fixes

* add query logging ([#236](https://github.com/teamkeel/keel/issues/236)) ([5c65b15](https://github.com/teamkeel/keel/commit/5c65b15b629682326b19b8ebb2b5ad5499ec112d))

## [0.90.3](https://github.com/teamkeel/keel/compare/v0.90.2...v0.90.3) (2022-08-09)


### Bug Fixes

* serialize dates as iso strings ([#235](https://github.com/teamkeel/keel/issues/235)) ([eeb6e78](https://github.com/teamkeel/keel/commit/eeb6e78bbef4a8c786f71a709f04f4abced96c60))

## [0.90.2](https://github.com/teamkeel/keel/compare/v0.90.1...v0.90.2) (2022-08-09)


### Bug Fixes

* fix sql generation for create statements ([#234](https://github.com/teamkeel/keel/issues/234)) ([096a5cc](https://github.com/teamkeel/keel/commit/096a5cc7e597c3389d83399c57c3de069e159f25))

## [0.90.1](https://github.com/teamkeel/keel/compare/v0.90.0...v0.90.1) (2022-08-09)


### Bug Fixes

* hook up runtime to db ([#233](https://github.com/teamkeel/keel/issues/233)) ([73fc00a](https://github.com/teamkeel/keel/commit/73fc00a550b5e5d4918fb65ac05830c08f26f455))

# [0.90.0](https://github.com/teamkeel/keel/compare/v0.89.4...v0.90.0) (2022-08-09)


### Features

* work on database connection setup ([#232](https://github.com/teamkeel/keel/issues/232)) ([fdfffe8](https://github.com/teamkeel/keel/commit/fdfffe8f4071a519b2a83f07bc675cc73292d4da))

## [0.89.4](https://github.com/teamkeel/keel/compare/v0.89.3...v0.89.4) (2022-08-08)


### Bug Fixes

* use full module path ([#228](https://github.com/teamkeel/keel/issues/228)) ([bdb06db](https://github.com/teamkeel/keel/commit/bdb06db293b36fa973a59d91a5581e4888f7c861))

## [0.89.3](https://github.com/teamkeel/keel/compare/v0.89.2...v0.89.3) (2022-08-08)


### Bug Fixes

* typings ([#227](https://github.com/teamkeel/keel/issues/227)) ([89d87e7](https://github.com/teamkeel/keel/commit/89d87e7c93b0cde22e0c9755393df4e377311940))

## [0.89.2](https://github.com/teamkeel/keel/compare/v0.89.1...v0.89.2) (2022-08-08)


### Bug Fixes

* fix package.json typings reference ([#226](https://github.com/teamkeel/keel/issues/226)) ([489e5ee](https://github.com/teamkeel/keel/commit/489e5eec2f4f36d8802d041f1b9fe9668de1fc72))

## [0.89.1](https://github.com/teamkeel/keel/compare/v0.89.0...v0.89.1) (2022-08-08)


### Bug Fixes

* sdk typings ([#225](https://github.com/teamkeel/keel/issues/225)) ([eea174b](https://github.com/teamkeel/keel/commit/eea174b5bfe27f01a2a2c4b26c730c0050ec1dc3))

# [0.89.0](https://github.com/teamkeel/keel/compare/v0.88.2...v0.89.0) (2022-08-08)


### Features

* codegen model api implementation ([#222](https://github.com/teamkeel/keel/issues/222)) ([0de87ff](https://github.com/teamkeel/keel/commit/0de87ffad4d2425b3740b3077c853f2c295e247d))

## [0.88.2](https://github.com/teamkeel/keel/compare/v0.88.1...v0.88.2) (2022-08-05)


### Bug Fixes

* use string values for enums instead of index values ([#220](https://github.com/teamkeel/keel/issues/220)) ([974569c](https://github.com/teamkeel/keel/commit/974569c428234bf4dbf5f4297017503fe91100b4))

## [0.88.1](https://github.com/teamkeel/keel/compare/v0.88.0...v0.88.1) (2022-08-04)


### Bug Fixes

* fix few minor bugs with the runtime / functions codegen interaction ([#219](https://github.com/teamkeel/keel/issues/219)) ([5d47c72](https://github.com/teamkeel/keel/commit/5d47c72e8d83b2aea1b2fe5cf8f8ad523e2ebd06))

# [0.88.0](https://github.com/teamkeel/keel/compare/v0.87.0...v0.88.0) (2022-08-04)


### Features

* hook up custom code runtime to run command ([#213](https://github.com/teamkeel/keel/issues/213)) ([dd2a9e7](https://github.com/teamkeel/keel/commit/dd2a9e75e97aaf99b7836f740d89fd4d7620c9fd))

# [0.87.0](https://github.com/teamkeel/keel/compare/v0.86.0...v0.87.0) (2022-08-02)


### Features

* suggest action names ([#208](https://github.com/teamkeel/keel/issues/208)) ([2f75b81](https://github.com/teamkeel/keel/commit/2f75b815015a11bcd8a6ed38f58b517d8c5fe3bf))

# [0.86.0](https://github.com/teamkeel/keel/compare/v0.85.0...v0.86.0) (2022-08-01)


### Features

* autocomplete model and enum names based on undefined types in other models ([#207](https://github.com/teamkeel/keel/issues/207)) ([5575638](https://github.com/teamkeel/keel/commit/557563877576b766d8cfb6f88a2f183e32c2aab6))

# [0.85.0](https://github.com/teamkeel/keel/compare/v0.84.1...v0.85.0) (2022-08-01)


### Features

* implement completions using text/scanner tokenisation ([789b116](https://github.com/teamkeel/keel/commit/789b116e3d56dd294586c7b703770e68ed482cb6))

## [0.84.1](https://github.com/teamkeel/keel/compare/v0.84.0...v0.84.1) (2022-08-01)


### Bug Fixes

* dont panic on empty section ([824878d](https://github.com/teamkeel/keel/commit/824878df04f8d9710bc0e61c2f2a1d9c59da028d))
* handle optional and repeated fields in schema formatter ([9207446](https://github.com/teamkeel/keel/commit/92074462e01e6682fffacbc493c0806e3227bcf9))
* make schema parser more flexible with attribute argument syntax ([fa4d4db](https://github.com/teamkeel/keel/commit/fa4d4dbbd02af122f08655975b533c9a67aa5579))

# [0.84.0](https://github.com/teamkeel/keel/compare/v0.83.2...v0.84.0) (2022-07-29)


### Features

* provide friendly wrapper functions to custom code files to abstract away internal types ([#200](https://github.com/teamkeel/keel/issues/200)) ([19f7fb8](https://github.com/teamkeel/keel/commit/19f7fb83917607e0ee3b76f805a9cbc01529e62e))

## [0.83.2](https://github.com/teamkeel/keel/compare/v0.83.1...v0.83.2) (2022-07-29)


### Bug Fixes

* use module.exports for compatibility with esbuild ([#201](https://github.com/teamkeel/keel/issues/201)) ([c44b903](https://github.com/teamkeel/keel/commit/c44b903e330e39b234eae42a0bf37e07b0837fd7))

## [0.83.1](https://github.com/teamkeel/keel/compare/v0.83.0...v0.83.1) (2022-07-28)


### Bug Fixes

* use npx version of tsc when generating client typings ([#199](https://github.com/teamkeel/keel/issues/199)) ([a805a1d](https://github.com/teamkeel/keel/commit/a805a1daf582502e143ce4c4593e1841ecb1b57a))

# [0.83.0](https://github.com/teamkeel/keel/compare/v0.82.0...v0.83.0) (2022-07-28)


### Features

* update static analysis runtime cli to process directories ([#197](https://github.com/teamkeel/keel/issues/197)) ([2001e7f](https://github.com/teamkeel/keel/commit/2001e7f68d87b86e5152a57dbe7fdc15fcde2e7a))

# [0.82.0](https://github.com/teamkeel/keel/compare/v0.81.0...v0.82.0) (2022-07-28)


### Features

* adds static analysis package to runtime npm package ([#196](https://github.com/teamkeel/keel/issues/196)) ([a2c59d1](https://github.com/teamkeel/keel/commit/a2c59d165451305766029bc60167a6f92c45c161))

# [0.81.0](https://github.com/teamkeel/keel/compare/v0.80.0...v0.81.0) (2022-07-28)


### Features

* fully fledged generate command ([#195](https://github.com/teamkeel/keel/issues/195)) ([fff8239](https://github.com/teamkeel/keel/commit/fff8239005c51cc56ebf1b7efda4de2a20502c85))

# [0.80.0](https://github.com/teamkeel/keel/compare/v0.79.2...v0.80.0) (2022-07-27)


### Features

* generate command output & integration test fixes ([#193](https://github.com/teamkeel/keel/issues/193)) ([54c3036](https://github.com/teamkeel/keel/commit/54c303647383f1764b5c3b7efd3d30ba0a3707ca))

## [0.79.2](https://github.com/teamkeel/keel/compare/v0.79.1...v0.79.2) (2022-07-27)


### Bug Fixes

* handle actions that have no inputs in graphql ([e816082](https://github.com/teamkeel/keel/commit/e816082ec5f102dee779a90512f7b3c2d26c8d74))

## [0.79.1](https://github.com/teamkeel/keel/compare/v0.79.0...v0.79.1) (2022-07-27)


### Bug Fixes

* fix runtime export ([#192](https://github.com/teamkeel/keel/issues/192)) ([f0a45fa](https://github.com/teamkeel/keel/commit/f0a45fab063036817765db7c4bdf77128fc659a5))

# [0.79.0](https://github.com/teamkeel/keel/compare/v0.78.0...v0.79.0) (2022-07-27)


### Features

* adds runtime to releaserc.json ([#191](https://github.com/teamkeel/keel/issues/191)) ([16b8032](https://github.com/teamkeel/keel/commit/16b803280f73ab96ea99aad03568b980d60720fb))
* publish new npm modules ([#190](https://github.com/teamkeel/keel/issues/190)) ([ffe5cfb](https://github.com/teamkeel/keel/commit/ffe5cfbc0dce149cd88e5ef4c076c51797629237))

# [0.78.0](https://github.com/teamkeel/keel/compare/v0.77.0...v0.78.0) (2022-07-26)


### Features

* support delete action type ([cb747df](https://github.com/teamkeel/keel/commit/cb747df9d2aa55823eb915a4cff0b64293c88cb0))

# [0.77.0](https://github.com/teamkeel/keel/compare/v0.76.0...v0.77.0) (2022-07-26)


### Features

* adds sdk postinstall script to generate dynamic code  ([#186](https://github.com/teamkeel/keel/issues/186)) ([216ea53](https://github.com/teamkeel/keel/commit/216ea5321328fbcf974935549fb850642b0bd740))

# [0.76.0](https://github.com/teamkeel/keel/compare/v0.75.0...v0.76.0) (2022-07-26)


### Features

* add sdk dep install to workflow ([#183](https://github.com/teamkeel/keel/issues/183)) ([ce7048f](https://github.com/teamkeel/keel/commit/ce7048f1c0d576b8c2b9bd6b959dfab382f7e0d8))

# [0.75.0](https://github.com/teamkeel/keel/compare/v0.74.0...v0.75.0) (2022-07-26)


### Features

* pin esbuild deps ([#182](https://github.com/teamkeel/keel/issues/182)) ([840a32a](https://github.com/teamkeel/keel/commit/840a32a2472e8f4a18fdc16031a684745ab830bf))

# [0.74.0](https://github.com/teamkeel/keel/compare/v0.73.0...v0.74.0) (2022-07-26)


### Features

* include correct esbuild dep ([#181](https://github.com/teamkeel/keel/issues/181)) ([992183c](https://github.com/teamkeel/keel/commit/992183c0c80f8c90fa8e83a3771e230eb046fcdd))

# [0.73.0](https://github.com/teamkeel/keel/compare/v0.72.0...v0.73.0) (2022-07-26)


### Features

* adds @teamkeel/sdk package ([#180](https://github.com/teamkeel/keel/issues/180)) ([b834177](https://github.com/teamkeel/keel/commit/b8341776595e76a8e1619a05f6ac810f7f7a35c3))

# [0.72.0](https://github.com/teamkeel/keel/compare/v0.71.1...v0.72.0) (2022-07-26)


### Features

* generate input interfaces from input definitions in schema ([#176](https://github.com/teamkeel/keel/issues/176)) ([013c947](https://github.com/teamkeel/keel/commit/013c947880c3a993aa3114d8761ab0c63e8b36b5))

## [0.71.1](https://github.com/teamkeel/keel/compare/v0.71.0...v0.71.1) (2022-07-25)


### Bug Fixes

* fix validation for permission attribute ([ab0c6bd](https://github.com/teamkeel/keel/commit/ab0c6bdf1c86e5e753e53d9461b99c1e6f2c6bcf))
* handle panics in wasm binary promise handler ([9123b53](https://github.com/teamkeel/keel/commit/9123b530be0ce3ebe22f84f13d153951d828d124))

# [0.71.0](https://github.com/teamkeel/keel/compare/v0.70.0...v0.71.0) (2022-07-25)


### Features

* reconcile client applications' package.json required dependencies automatically ([#170](https://github.com/teamkeel/keel/issues/170)) ([0873c0a](https://github.com/teamkeel/keel/commit/0873c0add3a2f30d8b3396d049f88bbd2bb7d54d))

# [0.70.0](https://github.com/teamkeel/keel/compare/v0.69.0...v0.70.0) (2022-07-25)


### Features

* provide scaffolding for function codegen ([#172](https://github.com/teamkeel/keel/issues/172)) ([3b41fd4](https://github.com/teamkeel/keel/commit/3b41fd44dc4b405bd97b80d6a32c1fd1208b4375))

# [0.69.0](https://github.com/teamkeel/keel/compare/v0.68.1...v0.69.0) (2022-07-25)


### Features

* connect run command to runtime ([33a7e4f](https://github.com/teamkeel/keel/commit/33a7e4fd443dbed9eac4bbe3e78bd91794d0ce53))

## [0.68.1](https://github.com/teamkeel/keel/compare/v0.68.0...v0.68.1) (2022-07-22)


### Bug Fixes

* time-based inputs ([995d725](https://github.com/teamkeel/keel/commit/995d725d52d0792b272822cd0eaa1d31632a74f6))

# [0.68.0](https://github.com/teamkeel/keel/compare/v0.67.0...v0.68.0) (2022-07-21)


### Features

* support basic runtime codegeneration ([#163](https://github.com/teamkeel/keel/issues/163)) ([542324e](https://github.com/teamkeel/keel/commit/542324e19d96511caeee25289901896602cc0d19))

# [0.67.0](https://github.com/teamkeel/keel/compare/v0.66.0...v0.67.0) (2022-07-21)


### Features

* improved handling of database setup in run command ([daab4b5](https://github.com/teamkeel/keel/commit/daab4b5c5e495b6a105d63dd8ab00439c3f0b5bc))

# [0.66.0](https://github.com/teamkeel/keel/compare/v0.65.0...v0.66.0) (2022-07-20)


### Features

* generate stub api type definitions for models ([#161](https://github.com/teamkeel/keel/issues/161)) ([5d03e83](https://github.com/teamkeel/keel/commit/5d03e833cd5007351f37f66c3a08bcd0a8457755))

# [0.65.0](https://github.com/teamkeel/keel/compare/v0.64.0...v0.65.0) (2022-07-19)


### Features

* run command ([f43ab41](https://github.com/teamkeel/keel/commit/f43ab4127815d36464a4fc16c8bf00907250a637))

# [0.64.0](https://github.com/teamkeel/keel/compare/v0.63.0...v0.64.0) (2022-07-19)


### Features

* generate typescript model + enum type definitions ([#159](https://github.com/teamkeel/keel/issues/159)) ([9f1b80d](https://github.com/teamkeel/keel/commit/9f1b80d660c4a46ea9caca9a5f8b468936560634))

# [0.63.0](https://github.com/teamkeel/keel/compare/v0.62.0...v0.63.0) (2022-07-19)


### Features

* set required fields to be NOT NULL ([173f6a7](https://github.com/teamkeel/keel/commit/173f6a781bc0cec3f93a34dddeba7989996053d5))

# [0.62.0](https://github.com/teamkeel/keel/compare/v0.61.0...v0.62.0) (2022-07-19)


### Features

* new migrations API ([5dce2eb](https://github.com/teamkeel/keel/commit/5dce2ebf09e8792d7f565c492a298395d483d9b9))

# [0.61.0](https://github.com/teamkeel/keel/compare/v0.60.0...v0.61.0) (2022-07-19)


### Features

* add support for relationships in graphql schema ([7efc97e](https://github.com/teamkeel/keel/commit/7efc97eeff8bd9a7dba5eede3fbe08c95f633c94))
* allow comparing and assigning null in expressions ([4bbff82](https://github.com/teamkeel/keel/commit/4bbff826a2ad9a45bec0cbdced83cf9804ed385e))

# [0.60.0](https://github.com/teamkeel/keel/compare/v0.59.0...v0.60.0) (2022-07-18)


### Features

* extract graphql schema language generation from tests to use in wasm binary ([d991c21](https://github.com/teamkeel/keel/commit/d991c216f93b0d8e867627749deae59592d78bdf))

# [0.59.0](https://github.com/teamkeel/keel/compare/v0.58.0...v0.59.0) (2022-07-18)


### Features

* add support [@validate](https://github.com/validate) attribute inside actions ([0bba4fe](https://github.com/teamkeel/keel/commit/0bba4fe73b68988f4be13efeee1af0c7475b196c))

# [0.58.0](https://github.com/teamkeel/keel/compare/v0.57.0...v0.58.0) (2022-07-15)


### Features

* input types for list actions ([8523230](https://github.com/teamkeel/keel/commit/85232305118834801ef8f271fa4fb611c780a7fe))

# [0.57.0](https://github.com/teamkeel/keel/compare/v0.56.0...v0.57.0) (2022-07-14)


### Features

* add comment support to parser and formatter ([746afef](https://github.com/teamkeel/keel/commit/746afeffe375259414e259a73921357bcdee582a))

# [0.56.0](https://github.com/teamkeel/keel/compare/v0.55.0...v0.56.0) (2022-07-14)


### Features

* add default value information to proto fields ([f46d485](https://github.com/teamkeel/keel/commit/f46d485e5d3f572c638424336f24157722a1e79c))

# [0.55.0](https://github.com/teamkeel/keel/compare/v0.54.0...v0.55.0) (2022-07-14)


### Features

* add grahpql input type generation for update operations ([6bef878](https://github.com/teamkeel/keel/commit/6bef8782fec4934695169b82b6bc7121472eb6f4))

# [0.54.0](https://github.com/teamkeel/keel/compare/v0.53.0...v0.54.0) (2022-07-13)


### Features

* add ctx.now as a way of getting the current time in an expression ([834bad2](https://github.com/teamkeel/keel/commit/834bad21b8289c494ec8ef685b832db5da807bf2))

# [0.53.0](https://github.com/teamkeel/keel/compare/v0.52.0...v0.53.0) (2022-07-13)


### Features

* add support for input types in GraphQL resolvers ([2faf779](https://github.com/teamkeel/keel/commit/2faf7796a5e2afae61d9316b6e5a5e2940114ebc))

# [0.52.0](https://github.com/teamkeel/keel/compare/v0.51.0...v0.52.0) (2022-07-13)


### Features

* add enum support to graphql resolver generation ([a309815](https://github.com/teamkeel/keel/commit/a309815d4840d0cf2ac09d557934633307dde18f))

# [0.51.0](https://github.com/teamkeel/keel/compare/v0.50.0...v0.51.0) (2022-07-13)


### Features

* validate invalid one-to-one relationships ([#140](https://github.com/teamkeel/keel/issues/140)) ([f303f24](https://github.com/teamkeel/keel/commit/f303f24da9b00ae29b823c2704754abde1aeeba7))

# [0.50.0](https://github.com/teamkeel/keel/compare/v0.49.0...v0.50.0) (2022-07-12)


### Features

* support [@unique](https://github.com/unique) attribute at model level for compount unique constraints ([88b17cb](https://github.com/teamkeel/keel/commit/88b17cbb9b5cdd5762abe844993d55d791b6f959))

# [0.49.0](https://github.com/teamkeel/keel/compare/v0.48.0...v0.49.0) (2022-07-11)


### Features

* update proto schema to better represent different input behaviours ([4ae9793](https://github.com/teamkeel/keel/commit/4ae9793dfde444f7eac7f0fc32752a8f5412e2d4))

# [0.48.0](https://github.com/teamkeel/keel/compare/v0.47.0...v0.48.0) (2022-07-10)


### Features

* regen and restart GraphQL server on each schema change ([#132](https://github.com/teamkeel/keel/issues/132)) ([87339a4](https://github.com/teamkeel/keel/commit/87339a4400b8cb3d9f6c011394595c72af9d97be))

# [0.47.0](https://github.com/teamkeel/keel/compare/v0.46.0...v0.47.0) (2022-07-08)


### Features

* expose basic completions api via wasm binary ([#131](https://github.com/teamkeel/keel/issues/131)) ([90703d2](https://github.com/teamkeel/keel/commit/90703d2d996794d76b030dd8ba32159c56e146fc))

# [0.46.0](https://github.com/teamkeel/keel/compare/v0.45.0...v0.46.0) (2022-07-06)


### Features

* do API server stop/start lifecycle properly ([#128](https://github.com/teamkeel/keel/issues/128)) ([9a38a0c](https://github.com/teamkeel/keel/commit/9a38a0ca501317d7cec5e06ca53491c15a450cf6))

# [0.45.0](https://github.com/teamkeel/keel/compare/v0.44.4...v0.45.0) (2022-07-04)


### Features

* gql create ([#125](https://github.com/teamkeel/keel/issues/125)) ([e5b66aa](https://github.com/teamkeel/keel/commit/e5b66aaa4100feaadd651fa1857e75dcefb4a201))

## [0.44.4](https://github.com/teamkeel/keel/compare/v0.44.3...v0.44.4) (2022-07-04)


### Bug Fixes

* fix formatter for short-hand inputs that use dot-notation ([#123](https://github.com/teamkeel/keel/issues/123)) ([8e68f90](https://github.com/teamkeel/keel/commit/8e68f9076097b8e664da42670715095bde7cf569))

## [0.44.3](https://github.com/teamkeel/keel/compare/v0.44.2...v0.44.3) (2022-07-04)


### Bug Fixes

* fix handling of repeated fields in expression validation ([#122](https://github.com/teamkeel/keel/issues/122)) ([cba2cba](https://github.com/teamkeel/keel/commit/cba2cba0e541117625dd1d9dc2ac0c24484dcf59))

## [0.44.2](https://github.com/teamkeel/keel/compare/v0.44.1...v0.44.2) (2022-07-01)


### Bug Fixes

* fix some errors in the todo schema example and dont require repeated fields to be set in a create operation ([6f72f79](https://github.com/teamkeel/keel/commit/6f72f79ea259514f3ccc0c1247d97696487b85cd))

## [0.44.1](https://github.com/teamkeel/keel/compare/v0.44.0...v0.44.1) (2022-07-01)


### Bug Fixes

* fix typings of wasm npm module ([#121](https://github.com/teamkeel/keel/issues/121)) ([3be0dc3](https://github.com/teamkeel/keel/commit/3be0dc36e319acb68d6e441baae0723601022636))

# [0.44.0](https://github.com/teamkeel/keel/compare/v0.43.0...v0.44.0) (2022-07-01)


### Features

* add validation for unused inputs ([4431776](https://github.com/teamkeel/keel/commit/4431776cb657a86b3048352cea1027e37958407e))

# [0.43.0](https://github.com/teamkeel/keel/compare/v0.42.0...v0.43.0) (2022-07-01)


### Features

* add validation for update operation inputs ([717c82a](https://github.com/teamkeel/keel/commit/717c82a0d8f6570037a6f1ca9a8cc3958f5939d4))

# [0.42.0](https://github.com/teamkeel/keel/compare/v0.41.0...v0.42.0) (2022-06-30)


### Features

* guard against no errors returned from validate call ([#118](https://github.com/teamkeel/keel/issues/118)) ([0bfb2f4](https://github.com/teamkeel/keel/commit/0bfb2f47cdaee5dd147f099911d8459d1fc04ab7))

# [0.41.0](https://github.com/teamkeel/keel/compare/v0.40.0...v0.41.0) (2022-06-30)


### Features

* remove comment and trigger new npm release ([#117](https://github.com/teamkeel/keel/issues/117)) ([e8e7f1f](https://github.com/teamkeel/keel/commit/e8e7f1f1589f579645a95c12d169827f53e9d2d1))

# [0.40.0](https://github.com/teamkeel/keel/compare/v0.39.0...v0.40.0) (2022-06-30)


### Features

* update ts wrapper to fix bugs ([#116](https://github.com/teamkeel/keel/issues/116)) ([58fa219](https://github.com/teamkeel/keel/commit/58fa219bc375b583ffa852af9faecd390c812dd2))

# [0.39.0](https://github.com/teamkeel/keel/compare/v0.38.0...v0.39.0) (2022-06-30)


### Features

* update typings ([#115](https://github.com/teamkeel/keel/issues/115)) ([cfbdd76](https://github.com/teamkeel/keel/commit/cfbdd76cc2f32e1a43bd87925a0d82885c615f31))

# [0.38.0](https://github.com/teamkeel/keel/compare/v0.37.0...v0.38.0) (2022-06-30)


### Features

* update validate function typing ([#114](https://github.com/teamkeel/keel/issues/114)) ([9ba4daf](https://github.com/teamkeel/keel/commit/9ba4daff10e716a9ce8cb8a5aa07bb4e7e1f23b2))

# [0.37.0](https://github.com/teamkeel/keel/compare/v0.36.0...v0.37.0) (2022-06-30)


### Features

* update typings ([#113](https://github.com/teamkeel/keel/issues/113)) ([e187e46](https://github.com/teamkeel/keel/commit/e187e463e166d6d40bd1d08b2aaba733be7bfa26))

# [0.36.0](https://github.com/teamkeel/keel/compare/v0.35.0...v0.36.0) (2022-06-30)


### Features

* update typings for wasm module ([#111](https://github.com/teamkeel/keel/issues/111)) ([e0ae3a9](https://github.com/teamkeel/keel/commit/e0ae3a910586c63825dadc711cca5e91c1ee9765))

# [0.35.0](https://github.com/teamkeel/keel/compare/v0.34.1...v0.35.0) (2022-06-30)


### Features

* add two neww validation rules for create actions ([bee72ff](https://github.com/teamkeel/keel/commit/bee72ff9916c063c2012972cde52db4150980f75))

## [0.34.1](https://github.com/teamkeel/keel/compare/v0.34.0...v0.34.1) (2022-06-30)


### Bug Fixes

* intermittent test failure in CI ([#109](https://github.com/teamkeel/keel/issues/109)) ([74b409e](https://github.com/teamkeel/keel/commit/74b409e5dcb4188687dd3627ad3c4d84dcfc15e2))

# [0.34.0](https://github.com/teamkeel/keel/compare/v0.33.0...v0.34.0) (2022-06-30)


### Features

* support with() syntax in schema formatter ([127ebf3](https://github.com/teamkeel/keel/commit/127ebf36362a646872a2a034fd70ef9a06fcb170))

# [0.33.0](https://github.com/teamkeel/keel/compare/v0.32.0...v0.33.0) (2022-06-30)


### Features

* update esbuild deps to be dev dependencies ([#108](https://github.com/teamkeel/keel/issues/108)) ([8117aeb](https://github.com/teamkeel/keel/commit/8117aebbbd716e175860ed3915f11c16f535971b))

# [0.32.0](https://github.com/teamkeel/keel/compare/v0.31.0...v0.32.0) (2022-06-30)


### Features

* define types manually to avoid ambient module relative import issues temporarily ([#106](https://github.com/teamkeel/keel/issues/106)) ([b86dc04](https://github.com/teamkeel/keel/commit/b86dc04799c88ba5ff485a937aae9a7a066a8080))

# [0.31.0](https://github.com/teamkeel/keel/compare/v0.30.0...v0.31.0) (2022-06-30)


### Features

* generate wasm typescript wrapper typings ([#105](https://github.com/teamkeel/keel/issues/105)) ([3c8f386](https://github.com/teamkeel/keel/commit/3c8f386e1bf100aed7bf1bf1ace146764801a429))

# [0.30.0](https://github.com/teamkeel/keel/compare/v0.29.0...v0.30.0) (2022-06-30)


### Features

* fix main entry script to point to dist/ ([#104](https://github.com/teamkeel/keel/issues/104)) ([93b1172](https://github.com/teamkeel/keel/commit/93b117207d4473ae0c4e7b85fd128e115931e7d7))

# [0.29.0](https://github.com/teamkeel/keel/compare/v0.28.0...v0.29.0) (2022-06-30)


### Features

* fix wasm generation ([#103](https://github.com/teamkeel/keel/issues/103)) ([c05fd1a](https://github.com/teamkeel/keel/commit/c05fd1a4b1fdd2bf46f0ef3492aaaf2333bba684))
* generate wasm binary prior to build ([#102](https://github.com/teamkeel/keel/issues/102)) ([82e2ee1](https://github.com/teamkeel/keel/commit/82e2ee1b7f9105456ccb67500d2148f764c937e6))

# [0.28.0](https://github.com/teamkeel/keel/compare/v0.27.0...v0.28.0) (2022-06-30)


### Features

* install esbuild ([#101](https://github.com/teamkeel/keel/issues/101)) ([fbfcc27](https://github.com/teamkeel/keel/commit/fbfcc278ef6fca0c62e3282d040954740b6809a0))

# [0.27.0](https://github.com/teamkeel/keel/compare/v0.26.0...v0.27.0) (2022-06-30)


### Features

* specify prepublish build ([#100](https://github.com/teamkeel/keel/issues/100)) ([b64939b](https://github.com/teamkeel/keel/commit/b64939b6f9e7037fe478bdf524ff953a23541e8d))

# [0.26.0](https://github.com/teamkeel/keel/compare/v0.25.0...v0.26.0) (2022-06-30)


### Features

* make wasm npm package public ([#99](https://github.com/teamkeel/keel/issues/99)) ([7513c81](https://github.com/teamkeel/keel/commit/7513c81d573710cd188aa5e40a07039961ece615))

# [0.25.0](https://github.com/teamkeel/keel/compare/v0.24.0...v0.25.0) (2022-06-30)


### Features

* install specific version of go in github action ([#98](https://github.com/teamkeel/keel/issues/98)) ([6e49261](https://github.com/teamkeel/keel/commit/6e49261d203bead26dc67dd3362756fc618e1234))

# [0.24.0](https://github.com/teamkeel/keel/compare/v0.23.0...v0.24.0) (2022-06-30)


### Features

* publish wasm on npm ([#93](https://github.com/teamkeel/keel/issues/93)) ([f195909](https://github.com/teamkeel/keel/commit/f19590979454395689aeba85aa9037ed615b5894))
* remove tarBallDirectory ([#96](https://github.com/teamkeel/keel/issues/96)) ([ce1c762](https://github.com/teamkeel/keel/commit/ce1c7623000d85e49a7577be82e485590189f98f))
* update conventional commits release workflow with NPM_TOKEN env value ([#97](https://github.com/teamkeel/keel/issues/97)) ([8b34d77](https://github.com/teamkeel/keel/commit/8b34d77573e894a0c17f6910e664a7f2c48b90cd))

# [0.23.0](https://github.com/teamkeel/keel/compare/v0.22.0...v0.23.0) (2022-06-29)


### Features

* Auto generate GraphQL server in Run package - first cut ([04c99d5](https://github.com/teamkeel/keel/commit/04c99d51593bf47c4a770ea733265fbdca04dd8f))

# [0.22.0](https://github.com/teamkeel/keel/compare/v0.21.0...v0.22.0) (2022-06-29)


### Features

* add parser and proto generation support for 'with' syntax ([2fb533a](https://github.com/teamkeel/keel/commit/2fb533a77513f4ddfbde432c19fae9bf2a128824))

# [0.21.0](https://github.com/teamkeel/keel/compare/v0.20.0...v0.21.0) (2022-06-29)


### Features

* add TypeInfo message to proto to align field and action input types ([6cb8b4e](https://github.com/teamkeel/keel/commit/6cb8b4e1ce3d2291aeb07694f8e2f213b9dee74d))

# [0.20.0](https://github.com/teamkeel/keel/compare/v0.19.0...v0.20.0) (2022-06-28)


### Features

* validate that lhs singular types match the rhs array type of T ([#88](https://github.com/teamkeel/keel/issues/88)) ([5b7300e](https://github.com/teamkeel/keel/commit/5b7300e374ca27265d598e7f4af50a149e42c7d1))

# [0.19.0](https://github.com/teamkeel/keel/compare/v0.18.0...v0.19.0) (2022-06-28)


### Features

* trimmed validation output ([da6f9c0](https://github.com/teamkeel/keel/commit/da6f9c00bf6493612290bec0c4f24181117a0589))

# [0.18.0](https://github.com/teamkeel/keel/compare/v0.17.0...v0.18.0) (2022-06-28)


### Features

* support enum resolution in expressions ([#87](https://github.com/teamkeel/keel/issues/87)) ([30a3a18](https://github.com/teamkeel/keel/commit/30a3a18bb478e9bc8975473797d4ce3627cf4b1f))

# [0.17.0](https://github.com/teamkeel/keel/compare/v0.16.0...v0.17.0) (2022-06-28)


### Features

* support the long-form definition of action inputs ([8e3c7b2](https://github.com/teamkeel/keel/commit/8e3c7b23263fb75e44203de1d559a5e7e26cb6e6))

# [0.16.0](https://github.com/teamkeel/keel/compare/v0.15.1...v0.16.0) (2022-06-27)


### Features

* check operator is acceptable based on attribute prior to typechecking lhs and rhs ([#85](https://github.com/teamkeel/keel/issues/85)) ([a582b97](https://github.com/teamkeel/keel/commit/a582b97c40f877fb0ff9014cad7f66a3407b9891))

## [0.15.1](https://github.com/teamkeel/keel/compare/v0.15.0...v0.15.1) (2022-06-27)


### Bug Fixes

* fix missing continue in attribute validation ([d81c1e9](https://github.com/teamkeel/keel/commit/d81c1e9e83e94e8060b16e1733062c2c7638abcc))

# [0.15.0](https://github.com/teamkeel/keel/compare/v0.14.0...v0.15.0) (2022-06-27)


### Features

* add formatting for schema ([7d69eb6](https://github.com/teamkeel/keel/commit/7d69eb6b18505c9e635ac84144a2dc33d9bb9f20))

# [0.14.0](https://github.com/teamkeel/keel/compare/v0.13.0...v0.14.0) (2022-06-27)


### Features

* implement initial type checking of expressions ([#83](https://github.com/teamkeel/keel/issues/83)) ([2672b58](https://github.com/teamkeel/keel/commit/2672b588227e7ffd36498c18ac97d6183c24ff2b))

# [0.13.0](https://github.com/teamkeel/keel/compare/v0.12.0...v0.13.0) (2022-06-24)


### Features

* turn parser errors into validation errors ([a295e4b](https://github.com/teamkeel/keel/commit/a295e4be35154de27a88d30b00a182f92cd3537d))

# [0.12.0](https://github.com/teamkeel/keel/compare/v0.11.0...v0.12.0) (2022-06-24)


### Features

* add wasm module ([5ed57a0](https://github.com/teamkeel/keel/commit/5ed57a00cbd7db767fe4cb234baedd568553ffa0))

# [0.11.0](https://github.com/teamkeel/keel/compare/v0.10.1...v0.11.0) (2022-06-24)


### Features

* add support for optional syntax on field types ([131decc](https://github.com/teamkeel/keel/commit/131decc21b87561967c9d7cd35a52bf3d3c2d7cb))

## [0.10.1](https://github.com/teamkeel/keel/compare/v0.10.0...v0.10.1) (2022-06-21)


### Bug Fixes

* adds missing Number field type. All of the proto generation for numerical types exists already ([#76](https://github.com/teamkeel/keel/issues/76)) ([2ab97b1](https://github.com/teamkeel/keel/commit/2ab97b1395f473d957a6bed23a12c9c696454385))

# [0.10.0](https://github.com/teamkeel/keel/compare/v0.9.0...v0.10.0) (2022-06-21)


### Features

* validate that [@unique](https://github.com/unique) doesn't accept any args ([#75](https://github.com/teamkeel/keel/issues/75)) ([842bbac](https://github.com/teamkeel/keel/commit/842bbac6f8567f827f85767da03cbd1d2ebd3562))

# [0.9.0](https://github.com/teamkeel/keel/compare/v0.8.0...v0.9.0) (2022-06-21)


### Features

* validate correct usage of expressions in [@set](https://github.com/set) and [@where](https://github.com/where) attributes ([#74](https://github.com/teamkeel/keel/issues/74)) ([a933f24](https://github.com/teamkeel/keel/commit/a933f24df6ad63423c26eabc3e88b6890422e823))

# [0.8.0](https://github.com/teamkeel/keel/compare/v0.7.0...v0.8.0) (2022-06-20)


### Bug Fixes

* adds spaces around operator ([9275f14](https://github.com/teamkeel/keel/commit/9275f1491ec5ef467184b8cfaf66b2de1d547025))
* change error symbol name ([e350ec0](https://github.com/teamkeel/keel/commit/e350ec0d67140385ca6b9ca2f7e965b91c692905))
* check length of variadic includeBuiltIn parameter ([9df74ca](https://github.com/teamkeel/keel/commit/9df74ca3ed2024f8ac96ce491fb43bd41611d431))
* fix bugginess of association resolution ([1283dce](https://github.com/teamkeel/keel/commit/1283dceb9b54621947f8b2a93c6e5d060e9b7ac4))
* fix use of variadic args ([3125c67](https://github.com/teamkeel/keel/commit/3125c672af739bd7eb6b85fda594c16b266c7240))
* fixes bug with expression validation display whereby positioning info is incorrect ([1e09a35](https://github.com/teamkeel/keel/commit/1e09a359e83864b4239356599e9feccd1ee51d5c))
* guard against root model not being found ([229a043](https://github.com/teamkeel/keel/commit/229a0435c760b5753713a7ca1818f046ed7ce162))
* panic if nil pointer ([0ec83a4](https://github.com/teamkeel/keel/commit/0ec83a422c1180ffea8cb0507a294409c6efac26))
* remove unnecessary lookup ([8beff68](https://github.com/teamkeel/keel/commit/8beff681c0080118a8bff33c10495741cbe42d3c))
* remove unnecessary splitting of ident ([a3b4355](https://github.com/teamkeel/keel/commit/a3b435577300040f6d73dca96536e7bacbbcb39b))
* remove unused methods ([b0d752e](https://github.com/teamkeel/keel/commit/b0d752e600814a9ec6ea77bca8b4b03804470a91))
* update older proto test cases that don't use proper syntax for expressions ([77bdd8c](https://github.com/teamkeel/keel/commit/77bdd8ceb2504fa02ad6ee5a6b4357f2488d124d))
* use fuzzy find models for pluralised model names ([6492a82](https://github.com/teamkeel/keel/commit/6492a8262ab15dc55456d8cd3a425cb7c35a0962))
* use mutated value instead of original ([2681c76](https://github.com/teamkeel/keel/commit/2681c76206ff38689897fd8fd8ad7a7c78144009))


### Features

* begin lhs / rhs parsing of expressions ([706f85a](https://github.com/teamkeel/keel/commit/706f85a8d467bc3f7abfa457fb37c9f61eb5412a))
* more work on resolving associations ([6533cdc](https://github.com/teamkeel/keel/commit/6533cdc2d3c1e1314897feae7420227c8c5b6fc0))

# [0.7.0](https://github.com/teamkeel/keel/compare/v0.6.0...v0.7.0) (2022-06-17)


### Features

* add validation roles for unique roles and apis ([91e783c](https://github.com/teamkeel/keel/commit/91e783c3cb4dc6228fc9ce8aee490cc846872978))
* add validation rule for unique enum definition ([473e8c4](https://github.com/teamkeel/keel/commit/473e8c44418c37407667bc44faee2b730726ae48))

# [0.6.0](https://github.com/teamkeel/keel/compare/v0.5.0...v0.6.0) (2022-06-16)


### Features

* introduce first version of the Run command ([5ed54e0](https://github.com/teamkeel/keel/commit/5ed54e03a4c14e512eb94563c008866fb2c91dd1))

# [0.5.0](https://github.com/teamkeel/keel/compare/v0.4.0...v0.5.0) (2022-06-13)


### Features

* add validation for the [@permission](https://github.com/permission) attribute ([03a06ab](https://github.com/teamkeel/keel/commit/03a06ab2c6f40ff45d29945238ed7591c3c898fe))

# [0.4.0](https://github.com/teamkeel/keel/compare/v0.3.1...v0.4.0) (2022-06-09)


### Features

* add support for enums ([ee0b674](https://github.com/teamkeel/keel/commit/ee0b6744539f19e6530dceb1c60cc1caecbbdfde))

## [0.3.1](https://github.com/teamkeel/keel/compare/v0.3.0...v0.3.1) (2022-06-07)


### Bug Fixes

* add GetPositionRange function to all parser nodes ([b799430](https://github.com/teamkeel/keel/commit/b799430194ad4f94e7e3c3b34f6e423fb66ff5dc))

# [0.3.0](https://github.com/teamkeel/keel/compare/v0.2.5...v0.3.0) (2022-06-06)


### Bug Fixes

* cleanup newlines ([66202aa](https://github.com/teamkeel/keel/commit/66202aaa17aa2f31030316a0bc08a44624d699c2))
* get all error highlighting working in right position in schema string ([28c51d7](https://github.com/teamkeel/keel/commit/28c51d75fa0b58ccf06a7b58c2ae31e87456dc0e))
* revert back to output package receiving generic interface for writing ([c88ff5f](https://github.com/teamkeel/keel/commit/c88ff5f7fe41b17653d3f64965337a8c809222e0))


### Features

* adds line numbers to schema output ([3f36c77](https://github.com/teamkeel/keel/commit/3f36c77dda34661aca9bdad6de0b2df9f0eea6b1))
* align error messages on the same column as violated token ([602b530](https://github.com/teamkeel/keel/commit/602b5301dd6b0534ae27eb3f19eaaa25ca87d7dc))
* fancy inline errors ([daaf839](https://github.com/teamkeel/keel/commit/daaf8397dc04cfa3d18ba15547821aef5cf88a71))
* green valid text ([5afa5d9](https://github.com/teamkeel/keel/commit/5afa5d9ea25874e27be3a7f4b1df2db73c5fd13e))
* highlight start of error in schema ([9e3d57e](https://github.com/teamkeel/keel/commit/9e3d57ee8cffe66eec6b6203cc7ca7106efed54b))
* remove extra down line ([2d371de](https://github.com/teamkeel/keel/commit/2d371de53a900a02b7ce7b46248dabcaa6a178de))
* restructure output package and share models via new package ([da15a22](https://github.com/teamkeel/keel/commit/da15a22455c6023375a55c31b739df05ba476c93))
* sort out end pos bug ([5338611](https://github.com/teamkeel/keel/commit/533861122c3070a830c02a78393b87db211c415d))
* update all nodes to include EndPos ([73f8167](https://github.com/teamkeel/keel/commit/73f8167701c31c190ca65ca9fa509d4d7af282af))
* update rules to reference NameToken.Name.Pos and EndPos ([90b35f5](https://github.com/teamkeel/keel/commit/90b35f566b87bfde29ff6d813bd48d15c286bd51))

## [0.2.5](https://github.com/teamkeel/keel/compare/v0.2.4...v0.2.5) (2022-05-25)


### Bug Fixes

* adds npm package.json version number bump within semantic release ([fe2c0ad](https://github.com/teamkeel/keel/commit/fe2c0ad19f7eb3a2859f0301163a089c17f696b3))

## [0.2.4](https://github.com/teamkeel/keel/compare/v0.2.3...v0.2.4) (2022-05-24)


### Bug Fixes

* adds npm plugin to sem release ([5f4029d](https://github.com/teamkeel/keel/commit/5f4029da697fdae3267e835dcd4c3b2148a7df88))

## [0.2.3](https://github.com/teamkeel/keel/compare/v0.2.2...v0.2.3) (2022-05-24)


### Bug Fixes

* use github access token for checkout to avoid protected branch issue ([88e258a](https://github.com/teamkeel/keel/commit/88e258a4a1b9798d2763b398ee17b1d94e429c16))

## [0.2.2](https://github.com/teamkeel/keel/compare/v0.2.1...v0.2.2) (2022-05-24)


### Bug Fixes

* **test:** test file ([da3cff1](https://github.com/teamkeel/keel/commit/da3cff13b87bf52e2022b6459fab90f24fd3c74d))

## [0.2.1](https://github.com/teamkeel/keel/compare/v0.2.0...v0.2.1) (2022-05-24)


### Bug Fixes

* cleanup semantic release versions ([289bb7c](https://github.com/teamkeel/keel/commit/289bb7c4866fe20f83c5177da796248c81b94a8e))
* reapply gitignore after node_module change ([b612d5f](https://github.com/teamkeel/keel/commit/b612d5f3345bc63ff719e2ecab2b271fbba4be78))

# [0.2.0](https://github.com/teamkeel/keel/compare/v0.1.0...v0.2.0) (2022-05-24)


### Bug Fixes

* remove test files ([720a714](https://github.com/teamkeel/keel/commit/720a714682ab33561cd2b0d99caf7b3c6d443682))


### Features

* adds commitlint conventional commits check ([f95b2de](https://github.com/teamkeel/keel/commit/f95b2de0b32c00a6f50b182b13f571c1edf60b74))
* test tagging ([7fbe2a2](https://github.com/teamkeel/keel/commit/7fbe2a2a99f0b8324ff6326956db147362800440))
