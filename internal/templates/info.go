package templates

import (
	"html/template"
	"io"
	"net/http"
)

// InfoData ...
type InfoData struct {
	ServerHostname string
	RemoteAddr     string
	Host           string
	Headers        map[string][]string
}

// FromRequest ...
func (d *InfoData) FromRequest(r *http.Request) {
	d.Headers = r.Header
	d.RemoteAddr = r.RemoteAddr
	d.Host = r.Host
}

// infoTemplate ...
type infoTemplate struct{ *template.Template }

// Render ...
func (t infoTemplate) Render(data InfoData, w io.Writer) {
	t.Execute(w, data)
}

var (
	// Info ..
	Info infoTemplate
)

const (
	body = `
<h1>Info for {{ .RemoteAddr }}</h1>
<table>
	<tr>
		<td>Server Hostname</td>
		<td>{{ .ServerHostname }}</td>
	</tr>
	
</table>
<h3>Headers</h3>
<table>
	
	<tr>
		<td>Host</td>
		<td>{{ .Host }}</td>
	</tr>
	{{ range $k, $v := .Headers }}
	<tr>
		<td>{{ $k }}</td>
		<td>{{index $v 0 }}</td>
	</tr>
	{{ end }}
</table>`
)

func init() {
	Info = infoTemplate{mustParse("info", body)}
}
