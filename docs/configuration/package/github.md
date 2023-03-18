# GitHub

GitHub type allows you to get GitHub repository or GitHub Release. To get releases, you need to specify `release` field.

In GitHub type, there are mainly two type of package style. One is a "repository" and the other is "release". In afx configuration, a `release` field is optional so basically all of GitHub packages are regard as "repository". It's a same reason why actual GitHub Release pages exists on its GitHub Repository. But if `release` field is specified, in afx, it's regard as also "release".

It may be good to think about whether to configure `release` field depending on where you install it from.

=== "Repository"
    ```yaml
    github:
    - name: ahmetb/kubectx
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

=== "Release"
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
    ```

## Parameters

### name

Type | Default
---|---
string | (required)

Package name. Basically you can name it as you like. However, in GitHub package, "owner/repo" style may be suitable.

### description

Type | Default
---|---
string | `""`

Package description.

### owner

Type | Default
---|---
string | (required)

Repository owner.

### repo

Type | Default
---|---
string | (required)

Repository name.

### branch

Type | Default
---|---
string | `master`

Remote branch name.

### with.depth

Type | Default
---|---
number | `0` (all commits)

Limit fetching to the specified number of commits from the tip of each remote branch history. If fetching to a shallow repository, specify 1 or more number, deepen or shorten the history to the specified number of commits.

### as

Key | Type | Default
---|---|---
gh-extension | Object | `null`

Change the installation behavior of the packages based on specified package type. In current afx, all packages are based on where it's hosted e.g. `github`. Almost all cases are not problem on that but some package types (such as "brew" or "gh extension") will be able to do more easily if there is a dedicated parameters to install the packages. In this `as` section, it expands more this installation method. Some of package types (especially "brew") will be coming soon in near future.

=== "gh-extension"

    Install a package as [gh extension](https://github.blog/2023-01-13-new-github-cli-extension-tools/). Officially gh extensions can be installed with `gh extension install owern/repo` command ([guide](https://cli.github.com/manual/gh_extension_install)) but it's difficult to manage what we downloaded as code. In afx, by handling them as the same as other packages, it allows us to codenize them.

    Key | Type | Default
    ---|---|---
    name | string | (required)
    rename-to | string | `""`

    ```yaml
    - name: yusukebe/gh-markdown-preview
      description: GitHub CLI extension to preview Markdown looks like GitHub.
      owner: yusukebe
      repo: gh-markdown-preview
      as:
        gh-extension:
          name: gh-markdown-preview
          rename-to: gh-md    # markdown-preview is long so rename it to shorten.
    ```

### release.name

Type | Default
---|---
string | `""`

Allows you to specify a package name managed in GitHub Release. You can find this by visiting release page of packages you want to install.

### relase.tag

Type | Default
---|---
string | `""`

Allows you to specify a tag version of GitHub Release. You can find this by visiting release page of packages you want to install.

### release.asset.filename

Type | Default
---|---
string | `""`

Allows you to specify a filename of GitHub Release asset you want to install. Moreover, this field in afx config file supports templating.

!!! tip "(Basically) NO NEED TO SET THIS"

    Basically in afx, it's no problem even if you don't specify this field when downloading a package from GitHub Release. Because afx automatically filters release assets that can work on your system (OS/Architecture, etc) even if several release assets are uploaded.

    But the filename of the package uploaded to GitHub Release can be named by its author freely. So there are cases that afx cannot filter the package which is suitable on your system when too much special wording is included in a filename.

    For example, the following case is,

    - a filename has "steve-jobs" instead of "mac" or "darwin": `some-package-v1.0.0-steve-jobs-amd64.tar.gz`

=== "Case 1"

    ```yaml hl_lines="9 10" title="Specify asset filename directly"
    github:
    - name: direnv/direnv
      description: Unclutter your .profile
      owner: direnv
      repo: direnv
      release:
        name: direnv
        tag: v2.30.3
        asset:
          filename: direnv.darwin-amd64
      command:
        link:
        - from: direnv
    ```

=== "Case 2"

    ```yaml hl_lines="9 10" title="Specify asset filename with templating"
    github:
    - name: direnv/direnv
      description: Unclutter your .profile
      owner: direnv
      repo: direnv
      release:
        name: direnv
        tag: v2.30.3
        asset:
          filename: '{{ .Release.Name }}.{{ .OS }}-{{ .Arch }}'
      command:
        link:
        - from: direnv
    ```

You can specify a filename from asset list on GitHub Release page. It allows to specify a filename directly and also to use name templating feature by using these variables provided by afx:

Key | Description
---|---
`.Release.Name` | Same as `release.name`
`.Release.Tag` | Same as `release.tag`
`.OS` | [GOOS](https://go.dev/doc/install/source#environment)[^1] (e.g. `darwin` etc)
`.Arch` | [GOARCH](https://go.dev/doc/install/source#environment)[^1] (e.g. `amd64` etc)

[^1]: This can be overwritten by `replace.asset.replacements`.

### release.asset.replacements

Type | Default
---|---
map | `{}`

Allows you to replace pre-defined OS/Architecture wording with yours. In afx, the templating variables of `.OS` and `.Arch` are coming from `runtime.GOOS` and `runtime.GOARCH` (The Go Programming Language). For example, your system is Mac: In this case, `GOOS` returns `darwin` string, but let's say the filename of the assets on GitHub Release you want has `mac` instead of `darwin`. In this case, you can replace it with `darwin` by defining this `replacements` map.

```yaml hl_lines="4"
asset:
  filename: '{{ .Release.Name }}-{{ .Release.Tag }}-{{ .Arch }}-{{ .OS }}.tar.gz'
  replacements:
    darwin: mac
```

Keys should be valid `GOOS`s or `GOARCH`s. Valid name is below (full is are on [Environment - The Go Programming Language](https://go.dev/doc/install/source#environment)). Values are the respective replacements.

`GOOS` | `GOARCH`
---|---
darwin|amd64
darwin|arm64
linux|386
linux|amd64
linux|arm64
windows|386
windows|amd64
windows|arm64

=== "Case 1"

    ```yaml hl_lines="9 10 11 12 13" title=""
    github:
    - name: sharkdp/bat
      description: A cat(1) clone with wings.
      owner: sharkdp
      repo: bat
      release:
        name: bat
        tag: v0.11.0
        asset:
          filename: '{{ .Release.Name }}-{{ .Release.Tag }}-{{ .Arch }}-{{ .OS }}.tar.gz'
          replacements:
            darwin: apple-darwin
            amd64: x86_64
      command:
        link:
        - from: '**/bat'
    ```

Due to specifying `release.asset.filename` field, you can choose what you install explicitly. It's not only but also you can replace these `.OS` and `.Arch` with what you like.

Above example will be templated from:

```
'{{ .Release.Name }}-{{ .Release.Tag }}-{{ .Arch }}-{{ .OS }}.tar.gz'
```
to:
```
bat-v0.11.0-x86_64-apple-darwin.tar.gz
```


### depends-on

Type | Default
---|---
list | `[]`

Allows you to specify dependency list between packages to handle hidden dependency that afx can't automatically infer.

Explicitly specifying a dependency is helpful when a package relies on some other package's behavior. Concretely it's good for handling the order of loading files listed on `plugin.sources` when running `afx init`.

Let's say you want to manage `pkg-a` and `pkg-b` with afx. Also let's say `pkg-a` needs to be loaded after `pkg-b` (This means `pkg-a` depends on `pkg-b`).

In this case you can specify dependencies:

=== "Case 1"

    ```yaml hl_lines="9 10"
    local:
    - name: zsh
      directory: ~/.zsh
      plugin:
        if: |
          [[ $SHELL == *zsh* ]]
        sources:
        - '[0-9]*.zsh'
      depends-on:
      - google-cloud-sdk
    - name: google-cloud-sdk
      directory: ~/Downloads/google-cloud-sdk
      plugin:
        env:
          PATH: ~/Downloads/google-cloud-sdk/bin
        sources:
        - '*.zsh.inc'
    ```

Thanks to `depends-on`, the order of loading sources are:

```console
* zsh -> google-cloud-sdk
```

Let's see the actual output with `afx init` in case we added `depends-on` like above config:

```console
$ afx init
...
source /Users/babarot/Downloads/google-cloud-sdk/completion.zsh.inc
source /Users/babarot/Downloads/google-cloud-sdk/path.zsh.inc
export PATH="$PATH:/Users/babarot/Downloads/google-cloud-sdk/bin"
...
source /Users/babarot/.zsh/10_utils.zsh
source /Users/babarot/.zsh/20_keybinds.zsh
source /Users/babarot/.zsh/30_aliases.zsh
source /Users/babarot/.zsh/50_setopt.zsh
source /Users/babarot/.zsh/70_misc.zsh
...
...
```

### command

See [Command](../command.md) page

### plugin

See [Plugin](../plugin.md) page
