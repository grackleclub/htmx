# example

This example program provides a 10 second running program for testing [`go-run`](../bin/go-run). It prints the contents of [static/test.txt](../static/test.txt) every second for 10 seconds, but is relaunced if the contents of the file changes.

See `go-run` update if the file contents of static are changed, as below.
```sh
$ echo "new contents" > test.txt
```
