# Gist

Gist type allows you to manage [Gist](https://gist.github.com/) pages as a plugin or command.

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

### name

Type | Default
---|---
string | (required)

Package name. Name it as you like. In Gist, there's a case that several files are attached in one Gist. So may be better to name considering it.

### description

Type | Default
---|---
string | `""`

Package description.

### owner

Type | Default
---|---
string | (required)

Gist owner.

### id

Type | Default
---|---
string | (required)

Gist page id.

### depends-on

See [GitHub#depends-on](github.md#depends-on) page. Same as that.

### command

See [Command](../command.md) page

### plugin

See [Plugin](../plugin.md) page
