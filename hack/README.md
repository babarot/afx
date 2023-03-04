## Installation script

Run script from local:

```console
$ cat hack/install | bash
```

Run script via curl (when not cloning repo):

```console
$ curl -sL https://raw.githubusercontent.com/b4b4r07/afx/HEAD/hack/install | AFX_VERSION=v0.1.24 bash
```

env | description | default
---|---|---
`AFX_VERSION` | afx version, available versions are on [releases](https://github.com/b4b4r07/afx/releases) | `latest` 
`AFX_BIN_DIR` | Path to install | `~/bin` 

