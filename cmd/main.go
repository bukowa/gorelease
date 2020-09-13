package main

import (
	. "github.com/bukowa/gorelease"
	"log"
)

var Version string

func main() {
	log.Printf("hello world from gorelease%s", Version)
	release := FromFile(".gorelease.yaml")
	handler := NewHandler()
	if err := handler.Prepare(release); err != nil {
		log.Fatal(err)
	}
	for _, t := range release.Targets {
		if err := handler.Build(&t); err != nil {
			log.Fatal(err)
		}
	}
}
