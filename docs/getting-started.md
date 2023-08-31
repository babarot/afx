# Getting Started

## Install the pre-compiled binary

You can install the pre-compiled binary (in several different ways), compile from source.

Below you can find the steps for each of them.

### bash script

bash installer has been provided so you can install afx by running this one command at your own risk.

=== "Latest"

    ```bash
    curl -sL https://raw.githubusercontent.com/babarot/afx/HEAD/hack/install | bash
    ```

=== "Version"

    ```bash
    curl -sL https://raw.githubusercontent.com/babarot/afx/HEAD/hack/install | AFX_VERSION=v0.1.24 bash
    ```

env | description | default
---|---|---
`AFX_VERSION` | afx version, available versions are on [releases](https://github.com/babarot/afx/releases) | `latest`
`AFX_BIN_DIR` | Path to install | `~/bin`

### go install

For Go developers.

```bash
go install github.com/babarot/afx@latest
```

### manually

Download the pre-compiled binaries from the [OSS releases page][releases] and copy them to the desired location.

[releases]: https://github.com/babarot/afx/releases

## Write YAML

Let's say you want to install `jq` and `enhancd` with afx. So please write YAML file like this:

```yaml
github:
- name: stedolan/jq
  description: Command-line JSON processor
  owner: stedolan
  repo: jq
  release:
    name: jq
    tag: jq-1.6
  command:
    link:
    - from: '*jq*'
      to: jq
- name: babarot/enhancd
  description: A next-generation cd command with your interactive filter
  owner: babarot
  repo: enhancd
  plugin:
    env:
      ENHANCD_FILTER: fzf --height 25% --reverse --ansi:fzy
    sources:
    - init.sh
```

This declaration means afx gets `jq` v1.6 from GitHub release and install it into PATH as a command.

Okay, then let's save this file in `~/.config/afx/main.yaml`.

## Install packages

After preparing YAML files, you become able to run `install` command:

```sh
$ afx install
```

This command runs install based on what were declared in YAML files.

## Initialize packages

After installed, you need to run this command to enable commands/plugins you installed.

```sh
$ source <(afx init)
```

`afx init` is just showing what needed to run commands/plugins. As a test, try to run.

```sh
$ afx init
source /Users/babarot/.afx/github.com/babarot/enhancd/init.sh
export ENHANCD_FILTER="fzf --height 25% --reverse --ansi:fzy"
```

As long as you don't run it with `source` command, it doesn't effect your current shell.

## Initialize when starting shell

Add this command to your shell config (e.g. .zshrc) enable plugins and commands you installed when starting shell.

```bash
# enable packages
source <(afx init)
```

## Update packages

If you want to update package to new version etc, all you have to do is just to modify YAML file and then run `afx update`:

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

```sh
$ afx update
✔ stedolan/jq
```

## Configure shell completions

You can also use shell completion with afx. To enable completion at starting a shell, you need to add below to your each shell "rc" files.

=== "Bash"

    ```console
    $ source <(afx completion bash)
    ```

=== "Zsh"

    ```console
    $ source <(afx completion zsh)
    ```

=== "Fish"

    ```console
    $ afx completion fish | source
    ```
