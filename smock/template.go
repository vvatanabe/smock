package smock

const tpl = `
package {{.PackageName}}

import (
{{- range .Imports }}
	{{.}}
{{- end}}
)

{{ range $i, $s := .Structures }}
type {{$s.Name}}Mock struct {
{{- range $s.Methods }}
	{{.Name}}Func func({{ .Params }}) {{.Returns}}
{{- end }}
}

{{- range $s.Methods }}
func (m *{{$s.Name}}Mock) {{.Name}}({{ .Params }}) {{.Returns}} {
	if m.{{.Name}}Func == nil {
		panic("This method is not defined.")
	}
{{- if .Returns}}
	return m.{{.Name}}Func({{ .Params.Names }})
{{- else}}
	m.{{.Name}}Func({{ .Params.Names }})
{{- end}}
}

{{ end }}
{{ end }}

`
