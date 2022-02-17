# Gist

## Example

```yaml
gist:
- name: hoge.sh
  description: this is a test for gist
  owner: b4b4r07
  id: f26dd264f094e0ca834ce9feadc0c3f1
  command:
    link:
    - from: hoge.sh
      to: hoge
```

## Parameters

Name | Type | Required | Description
---|---|---|---
name | string | yes | Package name (must be unique in all packages)
description | string | | A description of a package
owner | string | yes | Gist owner
id | string | yes | Gist page id
command | section | | See [Command](../command.md) page
plugin | section | | See [Plugin](../plugin.md) page
depends-on | array | Dependency list (you can write package name here)
