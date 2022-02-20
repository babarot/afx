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

### depends-on

See [GitHub#depends-on](github.md#depends-on) page. Same as that.

### command

See [Command](../command.md) page

### plugin

See [Plugin](../plugin.md) page
