# Changelog

## 1.0.0 (2025-12-22)


### ⚠ BREAKING CHANGES

* Minimum Go version increased from 1.21 to 1.24. The testing/synctest package requires Go 1.24+.

### Features

* add support for agents, setting sources, and plugins ([35014aa](https://github.com/panbanda/claude-agent-sdk-go/commit/35014aad8bdf2e17e170baefe9fe5978cb805f40))
* add support for agents, setting sources, and plugins ([e887a58](https://github.com/panbanda/claude-agent-sdk-go/commit/e887a58f807499bcdb9b8483cec3c0d16dc5cd56))
* **client:** add dynamic control methods for mid-conversation changes ([de9859d](https://github.com/panbanda/claude-agent-sdk-go/commit/de9859d9c1f3ee8bb2fcc199ed8e393a8c27e432))
* **client:** add dynamic control methods for mid-conversation changes ([28d49d9](https://github.com/panbanda/claude-agent-sdk-go/commit/28d49d981abc4d367375dcc2adcec196f735cb11))
* **client:** add file checkpointing and rewind support ([bef3a46](https://github.com/panbanda/claude-agent-sdk-go/commit/bef3a46b39bef5f5cf5e122c456f878752cf2f89))
* **client:** add file checkpointing and rewind support ([39a347b](https://github.com/panbanda/claude-agent-sdk-go/commit/39a347b55ea8ab0ea012621ad8b685ec4d19e445))
* **hooks:** add pre-tool hook execution via CLI control protocol ([6ffd614](https://github.com/panbanda/claude-agent-sdk-go/commit/6ffd61422a16cd5982c927282a21a8a653d591ba))
* **options:** add extraArgs, addDirs, settings, user, and betas options ([73d6b84](https://github.com/panbanda/claude-agent-sdk-go/commit/73d6b84df9531957537b722cc5737448144a4656))
* **options:** add extraArgs, addDirs, settings, user, and betas options ([7aec40a](https://github.com/panbanda/claude-agent-sdk-go/commit/7aec40ae9494b875775e0e5b0eeff3fb450b541d))
* **options:** add MCP server configuration support ([2ac9f2d](https://github.com/panbanda/claude-agent-sdk-go/commit/2ac9f2da286631955d31a4806c821f3aad77086c))
* **options:** add stderr callback for capturing CLI debug output ([e02ea4c](https://github.com/panbanda/claude-agent-sdk-go/commit/e02ea4cce8900213cf530c3c93034bf9aa68c80d))
* **options:** add stderr callback for capturing CLI debug output ([31b27d9](https://github.com/panbanda/claude-agent-sdk-go/commit/31b27d954484d9fd75c10673b672fc19511ab517))
* **options:** add structured output, sandbox, and streaming options ([28b67fe](https://github.com/panbanda/claude-agent-sdk-go/commit/28b67fec7aa79b6400ec1e895d0b987c5497070e))
* **options:** add structured output, sandbox, and streaming options ([402ed30](https://github.com/panbanda/claude-agent-sdk-go/commit/402ed30e11af53964456d2ff4697f90502de40a5))


### Bug Fixes

* address lint errors and adjust linter config ([6e8012e](https://github.com/panbanda/claude-agent-sdk-go/commit/6e8012e6cc5451f84bef7d673ba8754ca7ef9e2c))
* align betas flag with Python SDK, clarify user option ([d2ee87f](https://github.com/panbanda/claude-agent-sdk-go/commit/d2ee87fe361ce03734248e50c3c1e2013bc12bd6))
* align CLI flags with Python SDK implementation ([f301df9](https://github.com/panbanda/claude-agent-sdk-go/commit/f301df91ac1b1cf19f9e2a3b881dececb9c9ce7f))
* format struct field alignment and remove go version badge ([be51d0c](https://github.com/panbanda/claude-agent-sdk-go/commit/be51d0cc5b7a4e9bbaf08a04103f815f3b62e66c))
* **lint:** add osWindows constant for goconst linter ([c0afb5f](https://github.com/panbanda/claude-agent-sdk-go/commit/c0afb5fbb905720a5f227a1ce07886587d7d939c))
* **lint:** add OutputFormatType constant to fix goconst warning ([29469cd](https://github.com/panbanda/claude-agent-sdk-go/commit/29469cd99c81d1539066d6c777e5815fc5a40447))
* **lint:** add PluginType constant to fix goconst warning ([8ce537e](https://github.com/panbanda/claude-agent-sdk-go/commit/8ce537e0d5815dd3b254d8cee480f27902b93b3c))
* **lint:** nest settings under linters for golangci-lint v2 ([56ca046](https://github.com/panbanda/claude-agent-sdk-go/commit/56ca04686dbf1c07131a9dfdb5cc468cdf8c731f))
* **lint:** remove unused nolint directive ([9bf68a3](https://github.com/panbanda/claude-agent-sdk-go/commit/9bf68a333ea620336a6278cb8431b6d1eb6399e3))
* **lint:** resolve goconst, gosec, and exhaustive warnings ([3c048af](https://github.com/panbanda/claude-agent-sdk-go/commit/3c048af5f6e7eadf5fabcc144c6a749f9d079f39))
* **query:** close channel on ResultMessage instead of transport close ([c61ce82](https://github.com/panbanda/claude-agent-sdk-go/commit/c61ce82c6e63d96f76cc85621449d998ff299a1a))
* resolve all lint errors and add Taskfile ([02fe353](https://github.com/panbanda/claude-agent-sdk-go/commit/02fe353fb1e48b0d355a8851b54c936b927f3fe8))
* use --json-schema flag matching Python SDK ([6246bd1](https://github.com/panbanda/claude-agent-sdk-go/commit/6246bd140026c25e815efc1389db4d4d8a67f383))
* use env var for file checkpointing to match Python SDK ([0e5a5f2](https://github.com/panbanda/claude-agent-sdk-go/commit/0e5a5f2d7b96478e8133f0a8779df4beda07e021))


### Tests

* improve coverage to 93% with synctest for concurrency ([e807004](https://github.com/panbanda/claude-agent-sdk-go/commit/e807004cc00dee4f634b1b437d1bb82bf7b63dc9))

## Changelog
