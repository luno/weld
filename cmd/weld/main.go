// Command weld is a compile time code generation tool for Luno style
// state and backends dependency injection
//
// See github.com/luno/weld godoc for more details.
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/luno/weld/internal"
)

var (
	verbose = flag.Bool("verbose", false, "Be verbose")
	tags    = flag.String("tags", "", "Build tags to include in generated file")
)

func fatal(err error) {
	fmt.Printf("%+v\n", err)
	os.Exit(1)
}

func getArgs() (internal.Args, error) {
	wd, err := os.Getwd()
	if err != nil {
		return internal.Args{}, err
	}

	pkgs := flag.Args()
	if len(pkgs) == 0 {
		pkgs = []string{"."}
	}

	return internal.Args{
		Dir:     wd,
		Env:     os.Environ(),
		Verbose: *verbose,
		Pkgs:    pkgs,
		Tags:    *tags,
	}, nil
}

func main() {
	flag.Usage = func() { fmt.Println("Usage: weld [-verbose] pkgs...") }
	flag.Parse()

	args, err := getArgs()
	if err != nil {
		fatal(err)
	}

	if err := internal.Run(context.Background(), args); err != nil {
		fatal(err)
	}
}
