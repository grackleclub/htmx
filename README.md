# htmx

Go htmx helpers and other frontend tools.

## `go-run`
[`go-run`](./bin/go-run) is a wrapper for `go run` that watches for changes, and initiates an automatic relaunch and rerun if any files are changed, if there is an interrupt, or if the program exits, propagating any exit code. Arguments should be preserved.

Run bare to target the current directory.
```sh
$ bin/go-run
```

Run with arguments and path for more advanced uses.
```sh
$ bin/go-run ./cmd/server arg1 arg2
```
