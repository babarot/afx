# AFX Concepts

AFX is a command-line package manager. afx can allow us to manage almost all things available on GitHub, Gist and so on. Before, we needed to trawl web pages to download each package one by one. It's very annoying every time we set up new machine and also it's difficult to get how many commands/plugins we installed.

So afx's motivation is coming from that and to manage them with YAML files (as a code).

```console
$ afx help
Package manager for CLI

Usage:
  afx [flags]
  afx [command]

Available Commands:
  check       Check new updates on each package
  completion  Generate completion script
  help        Help about any command
  init        Initialize installed packages
  install     Resume installation from paused part (idempotency)
  self-update Update afx itself to latest version
  show        Show packages managed by afx
  uninstall   Uninstall installed packages
  update      Update installed package if version etc is changed

Flags:
  -h, --help      help for afx
  -v, --version   version for afx

Use "afx [command] --help" for more information about a command.
```
