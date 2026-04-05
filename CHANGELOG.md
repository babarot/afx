# Changelog

## [v0.2.3](https://github.com/babarot/afx/compare/v0.2.2...v0.2.3) - 2026-04-05
### Deprecated features
- Drop Windows support, rename pkg to manager, add state file lock by @babarot in https://github.com/babarot/afx/pull/71
### Refactorings
- Refactor package structure: break god package and migrate to internal/ by @babarot in https://github.com/babarot/afx/pull/69
- Replace custom error library with stdlib errors and fmt.Errorf by @babarot in https://github.com/babarot/afx/pull/73
- Migrate from deprecated mholt/archiver v3 to mholt/archives by @babarot in https://github.com/babarot/afx/pull/74
- Delegate gh extension management to gh CLI subprocess by @babarot in https://github.com/babarot/afx/pull/75
- Fix critical bugs and clean up dead code by @babarot in https://github.com/babarot/afx/pull/76
- Replace go-git library with git subprocess execution by @babarot in https://github.com/babarot/afx/pull/77

## [v0.2.2](https://github.com/babarot/afx/compare/v0.2.1...v0.2.2) - 2025-12-13
- fix: resolve issue with unrecognized symblic links by @elecdeer in https://github.com/babarot/afx/pull/62
- Add Git Authentication Support for Private Repositories by @babarot in https://github.com/babarot/afx/pull/64
- Migrate to tagpr-based release cycle by @babarot in https://github.com/babarot/afx/pull/65

## [v0.2.1](https://github.com/babarot/afx/compare/v0.2.0...v0.2.1) - 2023-05-18
- fix: fix install script by @ebiyu in https://github.com/babarot/afx/pull/59
- Create ~/.config/afx if not exist by @ebiyu in https://github.com/babarot/afx/pull/60

## [v0.2.0](https://github.com/babarot/afx/compare/v0.1.26...v0.2.0) - 2023-03-19
- Support gh extension by @babarot in https://github.com/babarot/afx/pull/58

## [v0.1.26](https://github.com/babarot/afx/compare/v0.1.25...v0.1.26) - 2023-03-10
- Delete existing before updating by @babarot in https://github.com/babarot/afx/pull/56
- Do not delete local dir on uninstall by @babarot in https://github.com/babarot/afx/pull/57

## [v0.1.25](https://github.com/babarot/afx/compare/v0.1.24...v0.1.25) - 2022-06-11
- Fix self-update #47 by @uga-rosa in https://github.com/babarot/afx/pull/50

## [v0.1.24](https://github.com/babarot/afx/compare/v0.1.23...v0.1.24) - 2022-05-29
- Support gz case (decompress) by @babarot in https://github.com/babarot/afx/pull/49

## [v0.1.23](https://github.com/babarot/afx/compare/v0.1.22...v0.1.23) - 2022-05-28
- Add build.directory by @babarot in https://github.com/babarot/afx/pull/48

## [v0.1.22](https://github.com/babarot/afx/compare/v0.1.21...v0.1.22) - 2022-05-04
- Improve resolveSymlink function by @babarot in https://github.com/babarot/afx/pull/44

## [v0.1.21](https://github.com/babarot/afx/compare/v0.1.20...v0.1.21) - 2022-04-23
- Resolve afx dir symlink by @babarot in https://github.com/babarot/afx/pull/41

## [v0.1.20](https://github.com/babarot/afx/compare/v0.1.19...v0.1.20) - 2022-03-24
- Refactor errors package by @babarot in https://github.com/babarot/afx/pull/35
- Avoid panic even if state file is empty by @babarot in https://github.com/babarot/afx/pull/39

## [v0.1.19](https://github.com/babarot/afx/compare/v0.1.18...v0.1.19) - 2022-03-21
- Fix bug when state.json does not exist by @babarot in https://github.com/babarot/afx/pull/37

## [v0.1.18](https://github.com/babarot/afx/compare/v0.1.17...v0.1.18) - 2022-03-10
- Add new command to check updates on each package by @babarot in https://github.com/babarot/afx/pull/31

## [v0.1.17](https://github.com/babarot/afx/compare/v0.1.16...v0.1.17) - 2022-03-09
- Fix `http.Installed()` behavior by @babarot in https://github.com/babarot/afx/pull/34

## [v0.1.16](https://github.com/babarot/afx/compare/v0.1.15...v0.1.16) - 2022-03-08
- Fix Installed() method behavior on plugin install by @babarot in https://github.com/babarot/afx/pull/33

## [v0.1.15](https://github.com/babarot/afx/compare/v0.1.14...v0.1.15) - 2022-03-01
- Refactor github package by @babarot in https://github.com/babarot/afx/pull/29
- Add error handling in build by @babarot in https://github.com/babarot/afx/pull/30

## [v0.1.14](https://github.com/babarot/afx/compare/v0.1.13...v0.1.14) - 2022-02-24
- Allow to use template variables in http.url by @babarot in https://github.com/babarot/afx/pull/28

## [v0.1.13](https://github.com/babarot/afx/compare/v0.1.12...v0.1.13) - 2022-02-23
- Refactor meta command by @babarot in https://github.com/babarot/afx/pull/25
- Resolve dependency on config package of state package by @babarot in https://github.com/babarot/afx/pull/26
- Improve show command by @babarot in https://github.com/babarot/afx/pull/27

## [v0.1.12](https://github.com/babarot/afx/compare/v0.1.11...v0.1.12) - 2022-02-23
- Fix inconsistency parts and remove bad implementations coming from old package by @babarot in https://github.com/babarot/afx/pull/24

## [v0.1.11](https://github.com/babarot/afx/compare/v0.1.10...v0.1.11) - 2022-02-22
- Update state logic by @babarot in https://github.com/babarot/afx/pull/23

## [v0.1.10](https://github.com/babarot/afx/compare/v0.1.9...v0.1.10) - 2022-02-21
- Add feature to ask if it's ok to run command by @babarot in https://github.com/babarot/afx/pull/22

## [v0.1.9](https://github.com/babarot/afx/compare/v0.1.8...v0.1.9) - 2022-02-20
- Add release.asset field to github configuration by @babarot in https://github.com/babarot/afx/pull/21

## [v0.1.8](https://github.com/babarot/afx/compare/v0.1.7...v0.1.8) - 2022-02-18
- Add update-notifier feature by @babarot in https://github.com/babarot/afx/pull/20
- Add completion command by @babarot in https://github.com/babarot/afx/pull/19
- Release afx with patch updates by @babarot in https://github.com/babarot/afx/pull/18
