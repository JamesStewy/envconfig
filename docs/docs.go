/*
Package docs implements helper functions for automatically generating documentation from envconfig configuration structs.
*/
package docs

import (
	"bytes"
	"github.com/JamesStewy/envconfig"
	"github.com/olekukonko/tablewriter"
	"html/template"
	"io"
	"strings"
)

func noteOptional(fld *envconfig.Field) string {
	note := fld.Note()
	if !fld.Optional() {
		return note
	}
	if note == "" {
		return "Optional."
	}
	return "Optional. " + note
}

func keysUpper(fld *envconfig.Field) []string {
	keys := fld.Keys()
	return keys[:len(keys)/2]
}

// TextTable writes each field in the configuration struct as a row in a text table.
func TextTable(w io.Writer, cinfo *envconfig.ConfInfo) {
	TextTableWithWidth(w, cinfo, tablewriter.MAX_ROW_WIDTH)
}

// TextTableString writes each field in the configuration struct as a row in a text table.
// TextTableString returns the table in a string.
func TextTableString(cinfo *envconfig.ConfInfo) string {
	buf := new(bytes.Buffer)
	TextTable(buf, cinfo)
	return buf.String()
}

// TextTableWithWidth writes each field in the configuration struct as a row in a text table.
// maxwidth sets the maximum number of charaters wide each column in the table can be.
func TextTableWithWidth(w io.Writer, cinfo *envconfig.ConfInfo, maxwidth int) {
	table := tablewriter.NewWriter(w)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetRowLine(true)
	TextTableWithOptions(w, cinfo, table, maxwidth)
}

// TextTableWithOptions writes each field in the configuration struct as a row in a text table.
// maxwidth sets the maximum number of charaters wide each column in the table can be.
func TextTableWithOptions(w io.Writer, cinfo *envconfig.ConfInfo, table *tablewriter.Table, maxwidth int) {
	table.SetAutoWrapText(false)
	table.SetHeader([]string{"Keys", "Value", "Default", "Note"})

	for _, fld := range *cinfo {
		value, _ := tablewriter.WrapString(fld.Value(), maxwidth)
		deflt, _ := tablewriter.WrapString(fld.Default(), maxwidth)
		note, _ := tablewriter.WrapString(noteOptional(fld), maxwidth)
		table.Append([]string{strings.Join(keysUpper(fld), "\n"), strings.Join(value, "\n"), strings.Join(deflt, "\n"), strings.Join(note, "\n")})
	}

	table.Render()
}

var base_tmpl *template.Template

func init() {
	base_tmpl = template.Must(template.New("base").Parse(`{{template "envconfig" .}}`))
}

// HTMLTableWithTemplate defines a new template named "envconfig".
// The "envconfig" template must be executed with a ConfInfo object.
// Use HTMLTableWithTemplate if you want to embed the table within a custom template.
func HTMLTableWithTemplate(t *template.Template) (*template.Template, error) {
	funcmap := template.FuncMap{
		"envconfigNoteOptional": noteOptional,
		"envconfigKeysUpper":    keysUpper,
	}
	return t.Funcs(funcmap).Parse(tmpl_src)
}

// HTMLTable writes each field in the configuration struct as a row in an HTML table.
func HTMLTable(w io.Writer, cinfo *envconfig.ConfInfo) error {
	if t, err := HTMLTableWithTemplate(base_tmpl); err == nil {
		return t.Execute(w, cinfo)
	} else {
		return err
	}
}

// HTMLTableString writes each field in the configuration struct as a row in an HTML table.
// HTMLTableString returns the table in a string.
func HTMLTableString(cinfo *envconfig.ConfInfo) (string, error) {
	buf := new(bytes.Buffer)
	if err := HTMLTable(buf, cinfo); err != nil {
		return "", err
	}
	return buf.String(), nil
}

var tmpl_src = `{{define "envconfig"}}<table>
	<thead>
		<tr>
			<th>Keys</th>
			<th>Value</th>
			<th>Default</th>
			<th>Note</th>
		</tr>
	</thead>
	<tbody>{{range .}}
		<tr>
			<th>{{range $index, $element := envconfigKeysUpper .}}{{if ne $index 0}}<br>{{end}}{{$element}}{{end}}</th>
			<th>{{.Value}}</th>
			<th>{{.Default}}</th>
			<th>{{envconfigNoteOptional .}}</th>
		</tr>{{end}}
	</tbody>
</table>{{end}}`
