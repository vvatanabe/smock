package main

import (
	"log"
	"os"

	"flag"

	"fmt"

	"io/ioutil"
	"path/filepath"

	"strings"

	"github.com/vvatanabe/smock/smock"
)

var (
	typeNames   = flag.String("type", "", "comma-separated list of type names; must be set")
	output      = flag.String("output", "", "output directory; default process whole package in current directory")
	showVersion = flag.Bool("v", false, "show version")
)

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of smock:\n")
	fmt.Fprintf(os.Stderr, "\tsmock [flags] -type T [directory] # Default: process whole package in current directory\n")
	fmt.Fprintf(os.Stderr, "\tsmock [flags] -type T files... # Must be a single package\n")
	fmt.Fprintf(os.Stderr, "For more information, see:\n")
	fmt.Fprintf(os.Stderr, "\thttps://godoc.org/github.com/vvatanabe/smock\n")
	fmt.Fprintf(os.Stderr, "Flags:\n")
	flag.PrintDefaults()
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("smock: ")
	flag.Usage = Usage
	flag.Parse()

	if *showVersion {
		fmt.Println("version:", smock.FmtVersion())
		return
	}

	if len(*typeNames) == 0 {
		flag.Usage()
		os.Exit(2)
	}
	types := strings.Split(*typeNames, ",")

	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}

	var dir string
	var g smock.Generator
	if len(args) == 1 && isDirectory(args[0]) {
		dir = args[0]
		g.ParsePackageDir(args[0])
	} else {
		dir = filepath.Dir(args[0])
		g.ParsePackageFiles(args)
	}

	baseName := fmt.Sprintf("%s_mock.go", ToCamel(types[0]))

	var outputFile string
	if *output != "" {
		g.SetPackageName(filepath.Base(*output))
		outputFile = filepath.Join(*output, baseName)
	} else {
		outputFile = filepath.Join(dir, baseName)
	}

	for _, typeName := range types {
		g.Generate(typeName)
	}

	src := g.Format()

	err := ioutil.WriteFile(outputFile, src, 0644)
	if err != nil {
		log.Fatalf("writing output: %s", err)
	}
}

func ToCamel(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if i > 0 && (c >= 'A' && c <= 'Z') {
			b.WriteByte('_')
		}
		b.WriteByte(c)
	}
	return strings.ToLower(b.String())
}

func isDirectory(name string) bool {
	info, err := os.Stat(name)
	if err != nil {
		log.Fatal(err)
	}
	return info.IsDir()
}
