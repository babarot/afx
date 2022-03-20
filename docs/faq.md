FAQ
===

## Debugging

In `afx`, it provides debugging feature by default. You can specify environment variables like:

```console
$ export AFX_LOG=debug
$ afx <sub-command>
```

It always shows debug messages while running afx, so it's ok to specify only oneline:

```console
$ AFX_LOG=debug afx ...
```

Currently log levels we can use are here:

Level | Message
---|---
INFO | Only show info message
WARN | Previous level, plus warn message
ERROR | Previous level, plus error message
DEBUG | Previous level, plug more detail log messages
TRACE | Previous level, plus all of log messages
