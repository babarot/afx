AFX - Package manager for CLI
---

AFX is a package manager for command-line tools and shell plugins. afx can allow us to manage almost all things available on GitHub, Gist and so on. Before, we needed to trawl web pages to download each package one by one. It's very annoying every time we set up new machine and also it's difficult to get how many commands/plugins we installed. So afx's motivation is coming from that and to manage them with YAML files (as a code).

[![Tests][test-mark]][test-link] [![Release][release-mark]][release-link]

[test-mark]: https://github.com/b4b4r07/afx/actions/workflows/go.yaml/badge.svg
[test-link]: https://github.com/b4b4r07/afx/actions/workflows/go.yaml

[release-mark]: https://github.com/b4b4r07/afx/actions/workflows/release.yaml/badge.svg
[release-link]: https://github.com/b4b4r07/afx/actions/workflows/release.yaml

Full document is here: [AFX](https://babarot.me/afx/)

## Features

- Allows to manage various packages types:
  - GitHub / GitHub Release / Gist / HTTP (web) / Local
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

## Quick Start

Details are here: [Getting Started - AFX](https://babarot.me/afx/getting-started/)

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

```console
$ source <(afx completion zsh)
```

bash and fish users are also available.

## Installation

Download the binary from [GitHub Release][release] and drop it in your `$PATH`.

- [Darwin / Mac][release]
- [Linux][release]

[release]: https://github.com/b4b4r07/afx/releases/latest
[website]: https://babarot.me/afx/

## License

MIT
