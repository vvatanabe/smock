package smock

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io"
	"strings"
	"text/template"
)

type Data struct {
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
	var brackets bool
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

func Gen(pkg string, src io.Reader, dist io.Writer) error {

	mockTpl, err := template.New("smock").Parse(tpl)
	if err != nil {
		panic(err)
	}

	fileSet := token.NewFileSet()
	astFile, err := parser.ParseFile(fileSet, "", src, 0)
	if err != nil {
		panic(err)
	}

	data := &Data{}
	data.PackageName = pkg
	for _, v := range astFile.Imports {
		data.Imports = append(data.Imports, v.Path.Value)
	}

	for _, decl := range astFile.Decls {

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

			structure := &Structure{}
			structure.Name = typeSpec.Name.String()

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

			data.Structures = append(data.Structures, structure)
		}
	}

	var buf bytes.Buffer
	err = mockTpl.Execute(&buf, data)
	if err != nil {
		panic(err)
	}

	fmted, err := format.Source(buf.Bytes())
	if err != nil {
		panic(err)
	}

	_, err = dist.Write(fmted)
	return err
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
