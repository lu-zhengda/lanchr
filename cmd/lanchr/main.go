//go:build darwin

package main

import (
	"fmt"
	"os"

	"github.com/zhengda-lu/lanchr/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
