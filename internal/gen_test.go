package internal

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/luno/jettison/errors"
	"github.com/luno/jettison/jtest"
	"github.com/luno/jettison/log"
	"github.com/sebdah/goldie/v2"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v2"
)

//go:generate go test -update

// TestMain creates temporary copies of gomod and gosum so the tests will pass.
// Since resulting go.sum will change depending latest bitx dependencies, these copies are deleted afterwards.
// See testdata/example/doc.go for more details.
func TestMain(m *testing.M) {
	copies := map[string]string{
		"testdata/example/gomod": "testdata/example/go.mod",
		"testdata/example/gosum": "testdata/example/go.sum",
	}

	// Create copies
	for actual, expect := range copies {
		b, err := ioutil.ReadFile(actual)
		if err != nil {
			log.Error(nil, errors.Wrap(err, "reading file"))
			os.Exit(1)
		}

		err = ioutil.WriteFile(expect, b, 0o644)
		if err != nil {
			log.Error(nil, errors.Wrap(err, "writing file"))
			os.Exit(1)
		}
	}

	cmd := exec.Command("go", "mod", "tidy")
	cmd.Dir = "testdata/example"
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		log.Error(nil, errors.Wrap(err, "go mody tidy"))
		os.Exit(1)
	}

	m.Run()

	// Delete copies
	for _, expect := range copies {
		err := os.Remove(expect)
		if err != nil {
			log.Error(nil, errors.Wrap(err, "delete"))
			os.Exit(1)
		}
	}
}

func TestGenerate(t *testing.T) {
	tests := []struct {
		Name    string
		WorkDir string
		Tags    string
	}{
		{
			Name:    "identity",
			WorkDir: "example/identity/state",
			Tags:    "!dev",
		},
		{
			Name:    "exchange",
			WorkDir: "example/exchange/state",
			Tags:    "!dev",
		},
		{
			Name:    "dev_exchange",
			WorkDir: "example/exchange/state/devstate",
		},
		{
			Name:    "dev_identity",
			WorkDir: "example/identity/state/devstate",
		},
		{
			Name:    "empty",
			WorkDir: "example/empty/state",
			Tags:    "!dev",
		},
		{
			Name:    "no_err",
			WorkDir: "example/no_err/state",
			Tags:    "!dev",
		},
		{
			Name:    "duplicate",
			WorkDir: "example/duplicate/state",
			Tags:    "!dev",
		},
		{
			Name:    "transitive",
			WorkDir: "example/transitive/state",
			Tags:    "!dev",
		},
		{
			Name:    "samevar",
			WorkDir: "example/samevar/state",
		},
		{
			Name:    "sort_with_bind",
			WorkDir: "example/sort_with_bind/state",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			g := goldie.New(t)

			wd, err := os.Getwd()
			require.NoError(t, err)

			targetDir := filepath.Join(wd, "testdata", test.WorkDir)
			targetWeldFile := filepath.Join(targetDir, "weld_gen.go")
			targetBcksFile := filepath.Join(targetDir, "backends_gen.go")
			_ = os.Remove(targetWeldFile)
			_ = os.Remove(targetBcksFile)

			res, err := Generate(context.Background(), Args{
				InDir:   targetDir,
				OutDir:  targetDir,
				Env:     nil,
				Verbose: true,
				Tags:    test.Tags,
			})
			jtest.Require(t, nil, err)
			require.Empty(t, res.Errors)

			var graph bytes.Buffer
			printNode(&graph, res.Root, 0, true)
			g.Assert(t, test.Name+"_"+"graph", graph.Bytes())

			var specBack bytes.Buffer
			printBackends(&specBack, res.SpecBackends)
			g.Assert(t, test.Name+"_"+"specBack", specBack.Bytes())

			var transBack bytes.Buffer
			for _, b := range res.TransBackends {
				printBackends(&transBack, b)
			}
			g.Assert(t, test.Name+"_"+"transBack", transBack.Bytes())

			var selected bytes.Buffer
			for _, node := range res.SelectedNodes {
				printNode(&selected, &node, 0, false)
			}
			g.Assert(t, test.Name+"_"+"selected", selected.Bytes())

			tpldata, err := yaml.Marshal(res.TplData)
			require.NoError(t, err)
			g.Assert(t, test.Name+"_"+"tpldata", tpldata)

			g.Assert(t, test.Name+"_"+"weldoutput", res.WeldOutput)
			g.Assert(t, test.Name+"_"+"bcksoutput", res.BackendsOutput)

			err = ioutil.WriteFile(targetWeldFile, res.WeldOutput, 0o644)
			require.NoError(t, err)

			if len(res.BackendsOutput) > 0 {
				err = ioutil.WriteFile(targetBcksFile, res.BackendsOutput, 0o644)
				require.NoError(t, err)
			}
		})
	}
}

func printBackends(w io.Writer, b Backends) {
	fmt.Fprintf(w, "%s[%d]: ", b.Type.String(), len(b.Deps))
	var deps []string
	for _, dep := range b.Deps {
		deps = append(deps, dep.Type.String())
	}
	sort.Strings(deps)
	fmt.Fprintf(w, "%s\n", strings.Join(deps, ", "))
}

func printNode(w io.Writer, node *Node, depth int, recurse bool) {
	fmt.Fprintf(w, "%s%s[%d", strings.Repeat("  ", depth), node.Type, len(node.Deps))
	if node.HasDups {
		fmt.Fprint(w, ",dups")
	}
	fmt.Fprint(w, "]: ")
	if node.Type == NodeTypeFunc {
		fmt.Fprintf(w, "%s\n", node.FuncObj)
	} else if node.Type == NodeTypeSet && node.VarName != "" {
		fmt.Fprintf(w, "var %s.%s\n", node.VarPkg, node.VarName)
	} else if node.Type == NodeTypeSet {
		fmt.Fprintf(w, "(inline)\n")
	} else if node.Type == NodeTypeBind {
		fmt.Fprintf(w, "%s(%s)\n", node.BindInterface, node.BindImpl)
	}
	if !recurse {
		return
	}
	for _, child := range node.Children {
		printNode(w, child, depth+1, true)
	}
}
