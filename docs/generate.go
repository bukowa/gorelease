package main

import (
	"github.com/bukowa/gorelease/cmd"
	"log"

	"github.com/bukforks/cobra/doc"
)

func main() {
	err := doc.GenMarkdownTree(cmd.RootCmd, "")
	if err != nil {
		log.Fatal(err)
	}
}
