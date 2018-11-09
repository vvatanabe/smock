package main

import (
	"os"

	"flag"

	"fmt"

	"github.com/vvatanabe/smock/smock"
)

func NewCLI() *CLI {
	return &CLI{}
}

type CLI struct {
}

func (cli *CLI) Run(argv []string) {
	var (
		pkg         string
		in          string
		out         string
		showVersion bool
	)

	flag.StringVar(&pkg, "pkg", "mock", "package name")
	flag.StringVar(&in, "in", "", "input file path")
	flag.StringVar(&out, "out", "", "output file path")
	flag.BoolVar(&showVersion, "v", false, "show version")
	flag.Parse()

	if showVersion {
		fmt.Println("version:", smock.FmtVersion())
		return
	}

	var (
		src *os.File
		err error
	)
	if in != "" {
		src, err = os.Open(in)
		if err != nil {
			os.Exit(1)
		}
	} else {
		src = os.Stdin
	}

	var dist *os.File
	if out != "" {
		dist, err = os.Open(out)
		if err != nil {
			os.Exit(1)
		}
	} else {
		dist = os.Stdout
	}

	smock.Gen(pkg, src, dist)
}

func main() {
	NewCLI().Run(os.Args)
}
