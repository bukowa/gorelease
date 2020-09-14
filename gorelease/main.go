package main

import (
	. "github.com/bukowa/gorelease/cmd"
	"os"
)

func main() {

	if err := RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
