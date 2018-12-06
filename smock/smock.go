package smock

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/build"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"log"
	"path/filepath"
	"strings"
	"text/template"
)

type Model struct {
	PackageName string
	Imports     []string
	Structures  []*Structure
}

type Structure struct {
	Name    string
	Methods []*Method
}

type Method struct {
	Name    string
	Params  Params
	Returns Returns
}

func (m *Method) Args() string {
	return m.Params.Names()
}

type Params []*variable

func (ps Params) String() string {
	var a []string
	for _, p := range ps {
		a = append(a, p.String())
	}
	return strings.Join(a, ",")
}

func (ps Params) Names() string {
	var names []string
	for _, p := range ps {
		if p.Name != "" && strings.HasPrefix(p.Type, "...") {
			names = append(names, p.Name+"...")
		} else {
			names = append(names, p.Name)
		}
	}
	return strings.Join(names, ",")
}

type Returns []*variable

func (rs Returns) String() string {
	var a []string
	brackets := len(rs) > 1
	for _, r := range rs {
		a = append(a, r.String())
		if !brackets {
			brackets = r.Name != ""
		}
	}
	joined := strings.Join(a, ",")
	if brackets {
		return fmt.Sprintf("(%s)", joined)
	}
	return joined
}

type variable struct {
	Name string
	Type string
}

func (p *variable) String() string {
	return fmt.Sprintf("%s %s", p.Name, p.Type)
}

type File struct {
	pkg     *Package
	astFile *ast.File
}

type Package struct {
	name  string
	files []*File
}

type Generator struct {
	pkg   *Package
	model *Model
	buf   bytes.Buffer
}

func (g *Generator) ParsePackageFiles(names []string) {
	g.parsePackage(".", names, nil)
}

func (g *Generator) ParsePackageDir(directory string) {
	pkg, err := build.Default.ImportDir(directory, 0)
	if err != nil {
		log.Fatalf("cannot process directory %s: %s", directory, err)
	}
	var names []string
	names = append(names, pkg.GoFiles...)
	names = append(names, pkg.CgoFiles...)
	names = prefixDirectory(directory, names)
	g.parsePackage(directory, names, nil)
}

func prefixDirectory(directory string, names []string) []string {
	if directory == "." {
		return names
	}
	ret := make([]string, len(names))
	for i, name := range names {
		ret[i] = filepath.Join(directory, name)
	}
	return ret
}

func (g *Generator) parsePackage(directory string, names []string, src interface{}) {
	var files []*File
	g.pkg = new(Package)
	g.model = new(Model)
	fs := token.NewFileSet()
	if src != nil {
		astFile, err := parser.ParseFile(fs, "", src, parser.ParseComments)
		if err != nil {
			log.Fatalf("parsing package: src: %s", err)
		}
		files = append(files, &File{
			astFile: astFile,
			pkg:     g.pkg,
		})
	} else {
		for _, name := range names {
			if !strings.HasSuffix(name, ".go") {
				continue
			}
			astFile, err := parser.ParseFile(fs, name, src, parser.ParseComments)
			if err != nil {
				log.Fatalf("parsing package: %s: %s", name, err)
			}
			files = append(files, &File{
				astFile: astFile,
				pkg:     g.pkg,
			})
		}
	}
	if len(files) == 0 {
		log.Fatalf("%s: no buildable Go files", directory)
	}
	g.model.PackageName = files[0].astFile.Name.Name
	g.pkg.files = files
}

func (g *Generator) ParseReader(src io.Reader) {
	g.parsePackage("", []string{}, src)
}

func (g *Generator) SetPackageName(name string) {
	g.model.PackageName = name
}

func (g *Generator) Generate(typeName string) {
	for _, file := range g.pkg.files {
		for _, importSpec := range file.astFile.Imports {
			g.model.Imports = append(g.model.Imports, importSpec.Path.Value)
		}
		for _, decl := range file.astFile.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			if genDecl.Tok.String() != "type" {
				continue
			}
			for _, spec := range genDecl.Specs {
				typeSpec, ok := spec.(*ast.TypeSpec)
				if !ok {
					continue
				}
				interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
				if !ok {
					continue
				}
				structureName := typeSpec.Name.String()
				if structureName != typeName {
					continue
				}
				structure := &Structure{}
				structure.Name = structureName
				for _, method := range interfaceType.Methods.List {
					funcType, ok := method.Type.(*ast.FuncType)
					if !ok {
						continue
					}
					var funcName string
					for _, n := range method.Names {
						funcName = n.Name
					}
					params, returns := parseFunc(funcType)
					structure.Methods = append(structure.Methods, &Method{
						Name:    funcName,
						Params:  params,
						Returns: returns,
					})
				}
				g.model.Structures = append(g.model.Structures, structure)
			}
		}
	}
}

func (g *Generator) Format() []byte {
	mockTemplate, err := template.New("smock").Parse(codeTemplate)
	if err != nil {
		log.Fatalln("could not parse Go template")
	}
	if err = mockTemplate.Execute(&g.buf, g.model); err != nil {
		log.Fatalln("could not apply data to Go template")
	}
	src, err := format.Source(g.buf.Bytes())
	if err != nil {
		log.Printf("warning: internal error: invalid Go generated: %s", err)
		log.Printf("warning: compile the package to analyze the error")
		return g.buf.Bytes()
	}
	return src
}

func parseFunc(funcType *ast.FuncType) (Params, Returns) {
	var (
		params  Params
		returns Returns
	)
	for _, p := range funcType.Params.List {
		params = append(params, parseField(p))
	}
	if funcType.Results != nil {
		for _, r := range funcType.Results.List {
			returns = append(returns, parseField(r))
		}
	}
	return params, returns
}

func interfaceType(t *ast.InterfaceType) string {
	var s strings.Builder
	s.WriteString("interface {\n")
	for _, fl := range t.Methods.List {
		ft, ok := fl.Type.(*ast.FuncType)
		if !ok {
			continue
		}
		var funcName string
		for _, n := range fl.Names {
			funcName = n.Name
		}
		params, returns := parseFunc(ft)
		s.WriteString(fmt.Sprintf("%s(%s)%s", funcName, params, returns))
		s.WriteString("\n")
	}
	s.WriteString("}")
	return s.String()
}

func parseArrayLen(len ast.Expr) string {
	if len == nil {
		return ""
	}
	switch t := len.(type) {
	case *ast.BinaryExpr:
		x := parseArrayLen(t.X)
		y := parseArrayLen(t.Y)
		return fmt.Sprintf("%s %s %s", x, t.Op.String(), y)
	case *ast.ParenExpr:
		return fmt.Sprintf("(%s)", parseArrayLen(t.X))
	case *ast.BasicLit:
		return t.Value
	default:
		panic(fmt.Sprintf("Failed parse array length. Please report with github issue. type: %v", t))
	}
}

func ident(t *ast.Ident) string {
	return t.Name
}

func ellipsis(t *ast.Ellipsis) string {
	return fmt.Sprintf("...%s", ParseType(t.Elt))
}

func starExpr(t *ast.StarExpr) string {
	return fmt.Sprintf("*%s", ParseType(t.X))
}

func selectorExpr(t *ast.SelectorExpr) string {
	return fmt.Sprintf("%s.%s", ParseType(t.X), t.Sel.Name)
}

func arrayType(t *ast.ArrayType) string {
	return fmt.Sprintf("[%s]%s", parseArrayLen(t.Len), ParseType(t.Elt))
}

func mapType(t *ast.MapType) string {
	return fmt.Sprintf("map[%s]%s", ParseType(t.Key), ParseType(t.Value))
}

func structType(t *ast.StructType) string {
	var s strings.Builder
	s.WriteString("struct {\n")
	for _, fl := range t.Fields.List {
		s.WriteString(parseField(fl).String())
		s.WriteString("\n")
	}
	s.WriteString("}")
	return s.String()
}

func chanType(t *ast.ChanType) string {
	var s strings.Builder
	switch t.Dir {
	case ast.SEND:
		s.WriteString("chan<- ")
	case ast.RECV:
		s.WriteString("<-chan ")
	default:
		s.WriteString("chan ")
	}
	s.WriteString(ParseType(t.Value))
	return s.String()
}

func funcType(t *ast.FuncType) string {
	params, returns := parseFunc(t)
	return fmt.Sprintf("func (%s)%s", params, returns)
}

func ParseType(t interface{}) string {
	switch t := t.(type) {
	case *ast.Ident:
		return ident(t)
	case *ast.Ellipsis:
		return ellipsis(t)
	case *ast.StarExpr:
		return starExpr(t)
	case *ast.SelectorExpr:
		return selectorExpr(t)
	case *ast.ArrayType:
		return arrayType(t)
	case *ast.MapType:
		return mapType(t)
	case *ast.StructType:
		return structType(t)
	case *ast.ChanType:
		return chanType(t)
	case *ast.InterfaceType:
		return interfaceType(t)
	case *ast.FuncType:
		return funcType(t)
	default:
		panic(fmt.Sprintf("Failed parse type. Please report with github issue. type: %v", t))
	}
}

func parseField(pl *ast.Field) *variable {
	var names []string
	for _, n := range pl.Names {
		names = append(names, n.Name)
	}
	return &variable{
		Name: strings.Join(names, ","),
		Type: ParseType(pl.Type),
	}
}
