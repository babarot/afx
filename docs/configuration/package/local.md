# Local

Local type allows you to manage a plugin or command located locally in your system. You can also just run `source` command on your rc files (e.g. zshrc) to load your settings divided into other shell scripts without using this Local type. But by using this package type, you can manage them as same like other packages on afx ecosystem.

```yaml
local:
- name: zsh
  description: My zsh scripts
  directory: ~/.zsh
  plugin:
    sources:
    - '[0-9]*.zsh'
- name: google-cloud-sdk
  description: Google Cloud SDK
  directory: ~/Downloads/google-cloud-sdk
  plugin:
    sources:
    - '*.zsh.inc'
```

## Parameters

### name

Type | Default
---|---
string | (required)

Package name.

### description

Type | Default
---|---
string | `""`

Package description.

### directory

Type | Default
---|---
string | (required)

Specify a directory path that files you want to load are put. Allow to use `~` (tilde) and environment variables (e.g. `$HOME`) here. Of course, specifying full path is also acceptable.

### depends-on

See [GitHub#depends-on](github.md#depends-on) page. Same as that.

### command

See [Command](../command.md) page

### plugin

See [Plugin](../plugin.md) page
