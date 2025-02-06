package main

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"
)

func main() {
	var file = path.Join("static", "test.txt")
	iterations := 10
	for i := range iterations {
		b, err := os.ReadFile(file)
		if err != nil {
			panic(err)
		}
		fmt.Printf("(%v) test.txt: %s\n", i, strings.TrimSpace(string(b)))
		time.Sleep(1 * time.Second)
	}
	os.Exit(99)
}
