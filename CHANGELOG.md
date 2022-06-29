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
