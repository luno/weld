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
	"path/filepath"

	"github.com/luno/jettison/errors"

	"github.com/luno/weld/internal"
)

var (
	in      = flag.String("in", "", "Path to package with weld spec (default current path)")
	out     = flag.String("out", "", "Path to write output to (default is value of -in)")
	verbose = flag.Bool("verbose", false, "Be verbose")
	tags    = flag.String("tags", "", "Build tags to include in generated file")
	testing = flag.Bool("testing", false, "Generate the testing variant that allows for test case dependency management")
)

func fatal(err error) {
	fmt.Printf("%+v\n", err)
	os.Exit(1)
}

func getArgs() (internal.Args, error) {
	var err error
	var inDir string
	if *in != "" {
		inDir, err = filepath.Abs(*in)
		if err != nil {
			return internal.Args{}, err
		}
	} else {
		inDir, err = os.Getwd()
		if err != nil {
			return internal.Args{}, err
		}
	}

	outDir := inDir
	if *out != "" {
		outDir, err = filepath.Abs(*out)
		if err != nil {
			return internal.Args{}, err
		}
	}

	return internal.Args{
		InDir:      inDir,
		OutDir:     outDir,
		Env:        os.Environ(),
		Verbose:    *verbose,
		Tags:       *tags,
		ForTesting: *testing,
	}, nil
}

func main() {
	flag.Parse()

	args, err := getArgs()
	if err != nil {
		// NoReturnErr: fatal
		fatal(err)
	}

	if err := run(context.Background(), args); err != nil {
		// NoReturnErr: fatal
		fatal(err)
	}
}

func run(ctx context.Context, args internal.Args) error {
	err := internal.RemoveGenFiles(args.OutDir)
	if err != nil {
		return err
	}

	res, err := internal.Generate(ctx, args)
	if err != nil {
		return err
	} else if len(res.Errors) > 0 {
		for _, e := range res.Errors {
			fmt.Printf("%+v\n", e)
		}
		return errors.New("generate error")
	}

	return internal.WriteGenFiles(res, args.OutDir, args.Verbose)
}
