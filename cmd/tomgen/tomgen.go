package main

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strings"

	"github.com/thehowl/tomino/generator"
	"golang.org/x/tools/go/packages"
)

func main() {
	flag.Parse()

	args := flag.Args()
	if err := run(args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
	}
}

func run(args []string) error {
	// parse symbols
	qsym := make([]qualifiedSymbol, 0, len(args))
	var paths []string
	for _, arg := range args {
		lastSlash := strings.IndexByte(arg, '/') + 1
		lastPart := arg[lastSlash:]
		dot := strings.LastIndexByte(lastPart, '.')
		if dot < 0 {
			// TODO: support symbols in local dir.
			return fmt.Errorf("invalid argument: %q (need a qualified symbol, like 'net/url.URL')", arg)
		}
		pkg := arg[:lastSlash+dot]
		qsym = append(qsym, qualifiedSymbol{
			pkg:    pkg,
			symbol: lastPart[dot+1:],
		})
		if !slices.Contains(paths, pkg) {
			paths = append(paths, pkg)
		}
	}

	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedDeps | packages.NeedTypes | packages.NeedSyntax | packages.NeedTypesSizes,
	}, paths...)
	if err != nil {
		return fmt.Errorf("loading packages: %w", err)
	}

	for _, sym := range qsym {
		pos := slices.IndexFunc(pkgs, func(pkg *packages.Package) bool { return pkg.PkgPath == sym.pkg })
		if pos < 0 {
			panic("could not find explicitly requested package?")
		}
		pkg := pkgs[pos]
		obj := pkg.Types.Scope().Lookup(sym.symbol)
		rec, err := generator.Parse(obj)
		if err != nil {
			return err
		}
		fmt.Println(rec)
	}

	return nil
}

type qualifiedSymbol struct {
	pkg    string
	symbol string
}
