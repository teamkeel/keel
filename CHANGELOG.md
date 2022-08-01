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
