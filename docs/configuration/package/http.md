# HTTP

## Example

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

Name | Type | Required | Description
---|---|---|---
name | string | yes | Package name (must be unique in all packages)
description | string | | A description of a package
url | string | yes | URL which can be downloaded.
output | string | | TBD
command | section | | See [Command](../command.md) page
plugin | section | | See [Plugin](../plugin.md) page
