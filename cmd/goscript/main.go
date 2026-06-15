package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gomazing/goscript/cmd/goscript/cli"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		return
	}

	command := strings.ToLower(os.Args[1])
	args := os.Args[2:]

	var err error
	switch command {
	case "add":
		if len(args) == 0 {
			usage()
			return
		}
		err = cli.AddComponent(args[0])
	case "fmt", "format":
		err = cli.FormatTargets(args)
	case "check", "doctor":
		err = cli.CheckTargets(args)
	case "index":
		err = cli.IndexTargets(args)
	case "watch", "dev":
		err = cli.WatchTargets(args)
	default:
		usage()
		return
	}

	if err != nil {
		log.Fatal(err)
	}
}

func usage() {
	fmt.Println("GoScript CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  goscript add <component-name>")
	fmt.Println("  goscript fmt [path ...]")
	fmt.Println("  goscript check [path ...]")
	fmt.Println("  goscript index [path ...]")
	fmt.Println("  goscript watch [path ...]")
}

