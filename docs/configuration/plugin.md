# Plugin

afx's goal is to finally support to install packages as `command`, `plugin` or both. In afx, several pacakge types (e.g. `github`) are supported but you can specify `command` and `plugin` field in all of sources.

## Parameters

### sources

Type | Required
---|---
list | yes

`sources` allows you to select what to load files when starting shell.

=== "Case 1"

    ```yaml hl_lines="9 10" title="Simple case, just register to init.sh as load scripts"
    github:
    - name: enhancd
      description: A next-generation cd command with your interactive filter
      owner: b4b4r07
      repo: enhancd
      plugin:
        env:
          ENHANCD_FILTER: fzf --height 25% --reverse --ansi:fzy
        sources:
        - init.sh
    ```

=== "Case 2"

    ```yaml hl_lines="10 11" title="Using wildcards to register multiple files"
    github:
    - name: zsh-prompt-minimal
      description: Super super super minimal prompt for zsh
      owner: b4b4r07
      repo: zsh-prompt-minimal
      plugin:
        env:
          PROMPT_PATH_STYLE: minimal
          PROMPT_USE_VIM_MODE: true
        sources:
        - '*.zsh-theme'
    ```

=== "Case 3"

    ```yaml hl_lines="5 6" title="Filenames starting with numbers"
    local:
    - name: zsh
      directory: ~/.zsh
      plugin:
        sources:
        - '[0-9]*.zsh'
    ```

### env

Type | Required
---|---
list | no

`env` allows you to set environment variables. By having this section in same YAML file of package declaration, you can manage it with same file. When we don't have afx, we should have environment variables in shell config (e.g. zshrc) even if not installed it yet or failed to install it. But thanks to afx, afx users can keep it with same files and enable it only while a package is installed.

!!! notes "Needs to login new shell"

    To enable environment variables to your shell, you need to run this command or start new shell after adding this command to your shel config (e.g. .zshrc):

    ```bash
    source <(afx init)
    ```

=== "Case 1"

    ```yaml hl_lines="7 8 9"
    github:
    - name: zsh-prompt-minimal
      description: Super super super minimal prompt for zsh
      owner: b4b4r07
      repo: zsh-prompt-minimal
      plugin:
        env:
          PROMPT_PATH_STYLE: minimal
          PROMPT_USE_VIM_MODE: true
        sources:
        - '*.zsh-theme'
    ```

### snippet

Type | Required
---|---
string | no

`snippet` allows you to specify the command which are runned when starting new shell.

=== "Case 1"

    ```yaml hl_lines="11 12 13" title="Login message if tpm is installed"
    github:
    - name: enhancd
      description: A next-generation cd command with your interactive filter
      owner: b4b4r07
      repo: enhancd
      plugin:
        env:
          ENHANCD_FILTER: fzf --height 25% --reverse --ansi:fzy
        sources:
        - init.sh
        snippet: |
          echo "enhancd is enabled, cd command is overrided by enhancd"
          echo "see github.com/b4b4r07/enhancd"
    ```

### snippet-prepare (beta)

Type | Required
---|---
string | no

`snippet-prepare` allows you to specify the command which are runned when starting new shell. Unlike `snippet`, this `snippet-prepare` is run before `source` command.

1. Run `snippet-prepare`
2. Load `sources`
3. Run `snippet`

This option comes from https://github.com/b4b4r07/afx/issues/6.

=== "Case 1"

    ```yaml hl_lines="7 8 9 10 11 12" title="Run snippet before sources"
    github:
    - name: sindresorhus/pure
      description: Pretty, minimal and fast ZSH prompt
      owner: sindresorhus
      repo: pure
      plugin:
        snippet-prepare: |
          zstyle :prompt:pure:git:branch color magenta
          zstyle :prompt:pure:git:branch:cached color yellow
          zstyle :prompt:pure:git:dirty color 091
          zstyle :prompt:pure:user color blue
          zstyle :prompt:pure:host color blue
        sources:
        - pure.zsh
    ```
