# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [0.0.14](https://github.com/SKalt/git-cc/compare/v0.0.13...v0.0.14) (2023-02-19)


### Bug Fixes

* **dist:** use separate gh token for brew tap ([1d97296](https://github.com/SKalt/git-cc/commit/1d972963e6defd56675285ce7dfb9dfde06dc1f8))

### [0.0.13](https://github.com/SKalt/git-cc/compare/v0.0.12...v0.0.13) (2023-02-19)


### Features

* **devtools:** use ./dist/git-cc as primary git-cc ([2c9ecdb](https://github.com/SKalt/git-cc/commit/2c9ecdbfbdc442d343c5e77b7ef8ef637508dfc0))


### Bug Fixes

* **dist:** use `homebrew`-prefixed repo as homebrew tap ([7a82692](https://github.com/SKalt/git-cc/commit/7a826929891468e1052d3c3b2fdbb6b17f353e1a))

### [0.0.12](https://github.com/SKalt/git-cc/compare/v0.0.11...v0.0.12) (2023-02-18)


### Features

* **cli:** handle mouse-based scrolling ([7bcb1c2](https://github.com/SKalt/git-cc/commit/7bcb1c2be84f0ae14c6832d4e02f3ee140964f5c))


### Bug Fixes

* **cli:** don't reset search on every cursor-blink ([f7164a4](https://github.com/SKalt/git-cc/commit/f7164a4fe19f38abe83945e623cbac1c1504ea4a)), closes [#34](https://github.com/SKalt/git-cc/issues/34)

### [0.0.11](https://github.com/SKalt/git-cc/compare/v0.0.10...v0.0.11) (2023-02-05)


### Features

* **cli:** accept map[string]string config ([e12c2f1](https://github.com/SKalt/git-cc/commit/e12c2f1ad8f34e0f10441b2340e1a41b61120596))
* **cli:** add --redo flag ([923d3a1](https://github.com/SKalt/git-cc/commit/923d3a1d5aba532862daff8bb9f86411242de5b0)), closes [#26](https://github.com/SKalt/git-cc/issues/26)
* **cli:** send new scope to clipboard!!! ([adcb96c](https://github.com/SKalt/git-cc/commit/adcb96c2d9c8acb7e8c9c6c31b02f4fc6f394e7f))
* **dist:** add homebrew tap ([13ed429](https://github.com/SKalt/git-cc/commit/13ed429968c1d546eee3e3bd87ac4ad3f167b436)), closes [#18](https://github.com/SKalt/git-cc/issues/18)
* **dist:** attempt manually triggering the gh release pipeline ([05a12b7](https://github.com/SKalt/git-cc/commit/05a12b73af777a8ee63a2f585d7588c98c821e85))


### Bug Fixes

* **cli:** fail fast on empty redo commits, remove old comments/unused code ([45e4d4a](https://github.com/SKalt/git-cc/commit/45e4d4a290d335960032e78f37f6d3ce97e20bb6))
* **cli:** return error rather than immediately panicing ([b0e02b4](https://github.com/SKalt/git-cc/commit/b0e02b45e85bb3c91a6da3f9b78df03a01ea3823))
* **cli:** validation-error conditions ([03b3a0c](https://github.com/SKalt/git-cc/commit/03b3a0cac9d538e44d95ea2fa9f3f110bf613048))

### [0.0.10](https://github.com/SKalt/git-cc/compare/v0.0.9...v0.0.10) (2022-11-06)


### Features

* **cli:** adds a warning for editor not found when creating scope ([8a31a1c](https://github.com/SKalt/git-cc/commit/8a31a1c0535a147652bd558743323136ce4d7777))
* **devtools:** draft nix flake ([804e87f](https://github.com/SKalt/git-cc/commit/804e87f8747a826ffebdb2560a29147b16c5637c))


### Bug Fixes

* **cli:** correctly handle spawning editors in a subprocess ([3a96806](https://github.com/SKalt/git-cc/commit/3a9680663ce6707882155aae26953e97939e1b8d))
* **devtools:** get `go build` working with nix ([71cc1dd](https://github.com/SKalt/git-cc/commit/71cc1dd962e68466d59b823607286284367b5e2e))
* **dist:** fix get_arch in install.sh ([1ef1bd7](https://github.com/SKalt/git-cc/commit/1ef1bd70c151be2ca1c9448be6de4877488669b7))
* prevent video from being stripped from GFM ([d35e20a](https://github.com/SKalt/git-cc/commit/d35e20a707c0c2cd07aa1e53a5a13dcda7afb27c))

### [0.0.9](https://github.com/SKalt/git-cc/compare/v0.0.8...v0.0.9) (2022-10-23)


### Bug Fixes

* **cli:** prevents next step until an available commit type is selected ([60563af](https://github.com/SKalt/git-cc/commit/60563af5745bf7617115693cddc549401cf80ab4))

### [0.0.8](https://github.com/SKalt/git-cc/compare/v0.0.7...v0.0.8) (2022-10-05)


### Features

* **dist:** add installer script ([760be44](https://github.com/SKalt/git-cc/commit/760be44cf057b1befb9986853ab1f5a0c36f3351))
* search upwards for config files ([1909295](https://github.com/SKalt/git-cc/commit/1909295443708b2ff0a4077dcaed583bf5161725))


### Bug Fixes

* **cli:** support `--no-signoff` `git-commit(1)` flag ([36af1b1](https://github.com/SKalt/git-cc/commit/36af1b16a5715ae35d2d5b83b88505fe85b61870))
* **dist:** ensure unpacked executable is, y'know, executable ([543060e](https://github.com/SKalt/git-cc/commit/543060ef160668a321d79aaff189c6637b38a1ad))
* **dist:** tar should unpack into a scratch directory, not cwd ([d927e66](https://github.com/SKalt/git-cc/commit/d927e66008dde5e49dbe44da0b70531a3bb57c82))
* support git-commit --no-verify flag ([a5b311e](https://github.com/SKalt/git-cc/commit/a5b311ea3fee760d6e07147452ce0ac33f4c2962))

### [0.0.7](https://github.com/SKalt/git-cc/compare/v0.0.6...v0.0.7) (2022-03-27)


### Features

* **dist:** adopt standard-version; draft goreleaser gh action ([787a667](https://github.com/SKalt/git-cc/commit/787a66733395fcfbe25fadafaf11ae20d23ddb9d))
* **installer:** warn about unsupported arch ([ca10e39](https://github.com/SKalt/git-cc/commit/ca10e395c7a54cc6d29e50caa44c71d9acac610c))


### Bug Fixes

* **cli:** update scope options once commit_convention.yaml gets closed ([3b4cce2](https://github.com/SKalt/git-cc/commit/3b4cce243be8fa56960cda77461b22e586e3180f))
* **installer:** use portable uname flag ([1c28ece](https://github.com/SKalt/git-cc/commit/1c28ecefa7dc1772f95f99e363f8597ed83eec23))

# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.
