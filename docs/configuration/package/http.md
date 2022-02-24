# HTTP

HTTP type allows you to manage a plugin or command hosted on any websites except for a source code hosting site such as GitHub etc.

```yaml
http:
- name: gcping
  description: Like gcping.com but a command line tool
  url: https://storage.googleapis.com/gcping-release/gcping_darwin_arm64_latest
  command:
    link:
    - from: gcping_*
      to: gcping
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

### url

Type | Default
---|---
string | (required)

Specify a URL that a command or plugin you want to install are hosted.

In this field, you can use template variables:

=== "Case 1"

    ```yaml hl_lines="4 5 6 7"
    http:
    - name: gcping
      description: Like gcping.com but a command line tool
      url: 'https://storage.googleapis.com/gcping-release/{{ .Name }}_{{ .OS }}_{{ .Arch }}_latest'
      templates:
        replacements:
          darwin: darwin # can replace "darwin" as you like!
      command:
        link:
        - from: gcping_*
          to: gcping
    ```

Key | Description
---|---
`.Name` | Same as `.name` (Package name)
`.OS` | [GOOS](https://go.dev/doc/install/source#environment)[^1] (e.g. `darwin` etc)
`.Arch` | [GOARCH](https://go.dev/doc/install/source#environment)[^1] (e.g. `amd64` etc)

[^1]: This can be overwritten by `templates.replacements`.

### templates.replacements

Type | Default
---|---
list | `[]`

In `.url` field, the template variables can be used. Also you can replace it with your own values. For more details, see also below page.

See [GitHub#release.asset.replacements](github.md#releaseassetreplacements)

### depends-on

See [GitHub#depends-on](github.md#depends-on) page. Same as that.

### command

See [Command](../command.md) page

### plugin

See [Plugin](../plugin.md) page
