# AFX Concepts

AFX is a command-line package manager. afx can allow us to manage almost all things available on GitHub, Gist and so on. Before, we nee ded to trawl web pages to download each package one by one. It's very annoying every time we set up new machine and also it's difficult to get how many commands/plugins we installed.

So afx's motivation is coming from that and to manage them with YAML files (as a code).

```console
$ afx help
Package manager for everything

Usage:
  afx [command]

Available Commands:
  help        Help about any command
  init        Initialize installed packages
  install     Resume installation from paused part (idempotency)
  uninstall   Uninstall installed packages

Flags:
  -h, --help   help for afx

Use "afx [command] --help" for more information about a command.
```
