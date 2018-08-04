package main

import (
	"fmt"
	"io/ioutil"
	"log"

	"github.com/kazu/ifacemaker/maker"
	"github.com/mkideal/cli"
)

type cmdlineArgs struct {
	cli.Helper
	Files          []string `cli:"*f,file" usage:"Go source file to read"`
	StructType     string   `cli:"*s,struct" usage:"Generate an interface for this structure name"`
	IfaceName      string   `cli:"*i,iface" usage:"Name of the generated interface"`
	PkgName        string   `cli:"*p,pkg" usage:"Package name for the generated interface"`
	ExcludeMethods []string `cli:"e,exclude" usage:"Method name for not generation " `
	CopyDocs       bool     `cli:"d,doc" usage:"Copy docs from methods" dft:"true"`
	Output         string   `cli:"o,output" usage:"Output file name. If not provided, result will be printed to stdout."`
}

func run(args *cmdlineArgs) {
	allMethods := []string{}
	allImports := []string{}
	mset := make(map[string]struct{})
	iset := make(map[string]struct{})
	parseds := make(map[string]*maker.StructData)
	for _, f := range args.Files {
		src, err := ioutil.ReadFile(f)
		if err != nil {
			log.Fatal(err.Error())
		}
		sparsed := maker.ParseStruct(src, args.CopyDocs, args.ExcludeMethods)
		for s, parsed := range sparsed {
			//parseds[s] = parsed
			if parseds[s] != nil {
				if parseds[s].Embedded != nil {
					parseds[s].Embedded = append(parseds[s].Embedded, parsed.Embedded...)
				}

				parseds[s].Methods = append(parseds[s].Methods, parsed.Methods...)
				parseds[s].Imports = append(parseds[s].Imports, parsed.Imports...)
			} else {
				parseds[s] = parsed
			}

		}
	}
	var methods []maker.Method
	var imports []string

	if parseds[args.StructType] != nil {
		methods = append(methods, parseds[args.StructType].Methods...)
		imports = parseds[args.StructType].Imports

		if len(parseds[args.StructType].Embedded) > 0 {
			for _, embed := range parseds[args.StructType].Embedded {
				if parseds[embed] != nil {
					methods = append(methods, parseds[embed].Methods...)
				}
			}
		}
	}

	for _, m := range methods {
		if _, ok := mset[m.Code]; !ok {
			allMethods = append(allMethods, m.Lines()...)
			mset[m.Code] = struct{}{}
		}
	}
	for _, i := range imports {
		if _, ok := iset[i]; !ok {
			allImports = append(allImports, i)
			iset[i] = struct{}{}
		}
	}

	result, err := maker.MakeInterface(args.PkgName, args.IfaceName, allMethods, allImports)
	if err != nil {
		log.Fatal(err.Error())
	}

	if args.Output == "" {
		fmt.Println(string(result))
	} else {
		ioutil.WriteFile(args.Output, result, 0644)
	}

}

func main() {
	cli.Run(&cmdlineArgs{}, func(ctx *cli.Context) error {
		argv := ctx.Argv().(*cmdlineArgs)
		run(argv)
		return nil
	})
}
