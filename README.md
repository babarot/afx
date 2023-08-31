AFX - Package manager for CLI
---

AFX is a package manager for command-line tools and shell plugins. afx can allow us to manage almost all things available on GitHub, Gist and so on. Before, we needed to trawl web pages to download each package one by one. It's very annoying every time we set up new machine and also it's difficult to get how many commands/plugins we installed. So afx's motivation is coming from that and to manage them with YAML files (as a code).

[![][afx-mark]][afx-link] [![][test-mark]][test-link] [![][release-mark]][release-link]

[afx-mark]: https://img.shields.io/github/v/release/babarot/afx?color=EF2D5E&display_name=release&label=AFX&logo=alchemy&logoColor=EF2D5E&sort=semver
[afx-link]: https://github.com/babarot/afx/releases

[test-mark]: https://github.com/babarot/afx/actions/workflows/go.yaml/badge.svg
[test-link]: https://github.com/babarot/afx/actions/workflows/go.yaml

[release-mark]: https://github.com/babarot/afx/actions/workflows/release.yaml/badge.svg
[release-link]: https://github.com/babarot/afx/actions/workflows/release.yaml

Full document is here: [AFX](https://babarot.me/afx/)

<img src="https://user-images.githubusercontent.com/4442708/224565945-2c09b729-82b7-4829-9cbc-e247b401b689.gif">

<!--
<img src="https://vhs.charm.sh/vhs-577hHga4xJRSvZFshv47y3.gif" width="">
<img src="https://vhs.charm.sh/vhs-46LPru8ovWFCQV6DnyKwGm.gif" width="">
<img src="https://vhs.charm.sh/vhs-6tz3U4NZyh9LBzmTlT98c9.gif" width="">
-->

## Features

- Allows to manage various packages types:
  - GitHub / GitHub Release / Gist / HTTP (web) / Local
  - [gh extensions](https://github.com/topics/gh-extension)
- Manages as CLI commands, shell plugins or both
- Easy to install/update/uninstall
- Easy to configure with YAML
  - Environment variables for each packages
  - Aliases for each packges
  - Conditional branches
  - Build steps
  - Run snippet code
  - Dependency between packages
  - etc...
- Works on bash, zsh and fish

## Quick Start [<sup>plus!</sup>](https://babarot.me/afx/getting-started/)

- [1. Install packages](#1-install-packages)
- [2. Load packages](#2-load-packages)
- [3. Update packages](#3-update-packages)
- [4. Uninstall packages](#4-uninstall-packages)

### 1. Install packages

Write YAML file with name as you like in `~/.config/afx/`. Let's say you write this code and then put it into `github.yaml`. After than you can install packages with `install` command.
```diff
+ github:
+ - name: stedolan/jq
+   description: Command-line JSON processor
+   owner: stedolan
+   repo: jq
+   release:
+     name: jq
+     tag: jq-1.5
+   command:
+     link:
+     - from: '*jq*'
+       to: jq
```

```console
$ afx install
```

### 2. Load packages

You can enable installed packages to your current shell with this command:

```console
$ source <(afx init)
```

Take it easy to run `afx init` because it just shows what to apply in your shell to Stdout.

If you want to automatically load packages when you start new shell, you need to add above to your shell-rc file.

### 3. Update packages

All you have to do for updating is just to update version part (release.tag) to next version then run `update` command.

```diff
  github:
  - name: stedolan/jq
    description: Command-line JSON processor
    owner: stedolan
    repo: jq
    release:
      name: jq
-     tag: jq-1.5
+     tag: jq-1.6
    command:
      link:
      - from: '*jq*'
        to: jq
```

```console
$ afx update
```

### 4. Uninstall packages

Uninstalling is also almost same as `install`. Just to remove unneeded part from YAML file then run `uninstall` command.

```diff
- github:
- - name: stedolan/jq
-   description: Command-line JSON processor
-   owner: stedolan
-   repo: jq
-   release:
-     name: jq
-     tag: jq-1.6
-   command:
-     link:
-     - from: '*jq*'
-       to: jq
```

```console
$ afx uninstall
```

## Advanced tips

### Shell completion

For zsh user, you can enable shell completion for afx:

```bash
# .zshrc
source <(afx completion zsh)
```

bash and fish users are also available.

## Installation

Download the binary from [GitHub Release][release] and drop it in your `$PATH`.

- [Darwin / Mac][release]
- [Linux][release]

Or, bash installer has been provided so you can install afx by running this one command at your own risk ([detail](./hack/README.md)).

```bash
curl -sL https://raw.githubusercontent.com/babarot/afx/HEAD/hack/install | bash
```

[release]: https://github.com/babarot/afx/releases/latest
[website]: https://babarot.me/afx/

## License

MIT
