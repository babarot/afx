# Command

afx's goal is to finally support to install packages as `command`, `plugin` or both. In afx, several pacakge types (e.g. `github`) are supported but you can specify `command` and `plugin` field in all of sources.

## Parameters

### build.steps

Type | Required
---|---
list | yes (when using `build`)

`build.steps` can be specified build commands to build a package.

=== "Case 1"

    ```yaml hl_lines="7 8 9 10" title="Using sudo"
    github:
    - name: fzy
      description: A better fuzzy finder
      owner: jhawthorn
      repo: fzy
      command:
        build:
          steps:
          - make
          - sudo make install
    ```

    In this case, build steps has `sudo` command but it can be run as expected. But in advance you will be asked to input sudo password.

=== "Case 2"

    ```yaml hl_lines="7 8 9" title="Using go build"
    github:
    - name: kubectl-trace
      description: Schedule bpftrace programs on your kubernetes cluster using the kubectl
      owner: iovisor
      repo: kubectl-trace
      command:
        build:
          steps:
          - go build -o kubectl-trace cmd/kubectl-trace/root.go
        link:
        - from: kubectl-trace
          to: kubectl-trace
    ```

    In this case, build steps run `go build` command because this package does not provide GitHub releases on its own page. So we need to build by ourselves. afx build feature is very helpful in such a case.

    `go build` command creates `red` command to current working directory so you need to have `link` section to install the built binary to your PATH.

### build.env

Type | Required
---|---
map | no

`build.env` can be specified environemnt variables used when running build a package.

=== "Case 1"

    ```yaml hl_lines="11 12"
    github:
    - name: fzy
      description: A better fuzzy finder
      owner: jhawthorn
      repo: fzy
      command:
        build:
          steps:
          - make
          - sudo make install
          env:
            VERSION: 1.0
    ```

    In this case, VERSION is specified to change version used in build steps.

### link.from

Type | Required
---|---
string | yes (when using `link`)

`link.from` can be specified where to install from.

=== "Case 1"

    ```yaml hl_lines="7 8" title="Just install from current directory"
    github:
    - name: diff-so-fancy
      description: Good-lookin' diffs. Actually… nah… The best-lookin' diffs.
      owner: so-fancy
      repo: diff-so-fancy
      command:
        link:
        - from: diff-so-fancy
    ```

    To specify where to install from, just need to fill in `link.from`.

=== "Case 2"

    ```yaml hl_lines="10 11 12" title="Case of including version string etc in file name"
    github:
    - name: jq
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
    ```

    This `link.from` field is based on downloaded package directory. If the binary name has unneeded string and it's difficult to find a binary with fixed string. In that case, you can use wildcard in `from` field.

### link.to

Type | Required
---|---
string | no

`link.to` can be specified where to install to.

=== "Case 1"

    ```yaml hl_lines="7 8 9" title="Simple case, with renaming by using `to` field"
    github:
    - name: prok
      description: easy process grep with ps output
      owner: mutantcornholio
      repo: prok
      command:
        link:
        - from: prok.sh
          to: prok
    ```

    By filling in `link.to` field, you can specify where to install to. This field can be omitted but in that case it will be regarded as same thing of `link.from` value. In short, you don't need to fill in this field if you don't need to rename `link.from` to new one while linking.

    If you want to rename command name from `link.from` to new one, you can use this field like above example (this `prok.sh` will be renamed to `prok` and then install it into PATH)

=== "Case 2"

    ```yaml hl_lines="7 8 9" title="from current working dir to external dir of afx"
    github:
    - name: tpm
      description: Tmux Plugin Manager
      owner: tmux-plugins
      repo: tpm
      command:
        link:
        - from: .
          to: $HOME/.tmux/plugins/tpm
    ```

    This example shows you to link all contents in current working directory to `~/.tmux/plugins/tpm` directory. You can use outside directory of afx in `link.to` field, but in that case you need to specify full path. 

    Using tilda `~` is also ok to specify `$HOME`.

=== "Case 3"

    ```yaml hl_lines="8 9 10 11" title="Several links"
    github:
    - name: kubectx
      description: Switch faster between clusters and namespaces in kubectl
      owner: ahmetb
      repo: kubectx
      command:
        link:
        - from: kubectx
          to: kubectl-ctx
        - from: kubens
          to: kubectl-ns
    ```

    `link` section is a list, so you can specify several pairs of `from` and `to`.

### env

Type | Required
---|---
map | no

`env` allows you to set environment variables. By having this section in same YAML file of package declaration, you can manage it with same file. When we don't have afx, we should have environment variables in shell config (e.g. zshrc) even if not installed it yet or failed to install it. But thanks to afx, afx users can keep it with same files and enable it only while a package is installed.

!!! notes "Needs to login new shell"

    To enable environment variables to your shell, you need to run this command or start new shell after adding this command to your shel config (e.g. .zshrc):

    ```bash
    source <(afx init)
    ```

=== "Case 1"

    ```yaml hl_lines="12 13 14 15"
    github:
    - name: bat
      description: A cat(1) clone with wings.
      owner: sharkdp
      repo: bat
      release:
        name: bat
        tag: v0.11.0
      command:
        alias:
          bat-theme: bat --list-themes | fzf --preview='bat --theme={} --color=always ~/.zshrc'
        env:
          BAT_PAGER: less -RF
          BAT_STYLE: numbers,changes
          BAT_THEME: ansi-dark
        link:
        - from: '**/bat'
    ```

### alias

Type | Required
---|---
map | no

`alias` allows you to set command aliases.

=== "Case 1"

    ```yaml hl_lines="7 8"
    github:
    - name: colordiff
      description: Primary development for colordiff
      owner: daveewart
      repo: colordiff
      command:
        alias:
          diff: colordiff -u
        link:
        - from: colordiff.pl
          to: colordiff
    ```

=== "Case 2"

    ```yaml hl_lines="10 11 12 13 14 15"
    github:
    - name: exa
      description: A modern version of 'ls'.
      owner: ogham
      repo: exa
      release:
        name: exa
        tag: v0.9.0
      command:
        alias:
          l: exa --group-directories-first -T --git-ignore --level 2
          la: exa --group-directories-first -a --header --git
          ll: exa --group-directories-first -l --header --git
          lla: exa --group-directories-first -la --header --git
          ls: exa --group-directories-first
        link:
        - from: '*exa*'
          to: exa
    ```

=== "Case 3"

    ```yaml hl_lines="10 11"
    github:
    - name: bat
      description: A cat(1) clone with wings.
      owner: sharkdp
      repo: bat
      release:
        name: bat
        tag: v0.11.0
      command:
        alias:
          bat-theme: bat --list-themes | fzf --preview='bat --theme={} --color=always ~/.zshrc'
        env:
          BAT_PAGER: less -RF
          BAT_STYLE: numbers,changes
          BAT_THEME: ansi-dark
        link:
        - from: '**/bat'
    ```

### snippet

Type | Required
---|---
string | no

`snippet` allows you to specify the command which are runned when starting new shell.

=== "Case 1"

    ```yaml hl_lines="10 11 12" title="Login message if tpm is installed"
    github:
    - name: tpm
      description: Tmux Plugin Manager
      owner: tmux-plugins
      repo: tpm
      command:
        link:
        - from: .
          to: $HOME/.tmux/plugins/tpm
        snippet: |
          echo "tpm is installed, so tmux will be automatically launched"
          echo "see github.com/tmux-plugins/tpm"

    ```

### if

Type | Required
---|---
string | no

`if` allows you to specify the condition to load packages. If it returns true, then the command will be linked. But if it returns false, the command will not be linked.

In `if` field, you can write shell scripts (currently `bash` is only supported). The exit code finally returned from that shell script is used to determine whether it links command or not.

=== "Case 1"

    ```yaml hl_lines="10 11" title="link commands if git is installed"
    github:
    - name: chmln/sd
      description: Intuitive find & replace CLI (sed alternative)
      owner: chmln
      repo: sd
      release:
        name: sd
        tag: 0.6.5
      command:
        if: |
          type git &>/dev/null
        snippet: |
          replace() {
            case "${#}" in
              1) git grep "${1}" ;;
              2) git grep -l "${1}" | xargs -I% sd "${1}" "${2}" % ;;
            esac
          }
    ```
