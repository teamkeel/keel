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
