# GitHub

This `github` allows you to get GitHub repositories and releases. To get releases, you need to specify `release` field.

## Examples

=== "Repository"
    ```yaml
    github:
    - name: kubectx
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
    - name: jq
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

Name | Type | Required | Description
---|---|---|---
name | string | yes | Package name (must be unique in all packages)
description | string | | A description of a package
owner | string | yes | GitHub owner
repo | string | yes | GitHub repo
with.depth | int | no | Fetch commit depth (default: 0). N>0 means shallow clone
release.name | string | yes (in `release`) | GitHub release name
release.tag | string | | GitHub release tag
command | section | | See [Command](../command.md) page
plugin | section | | See [Plugin](../plugin.md) page
