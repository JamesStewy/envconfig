package docs_test

import (
	"github.com/JamesStewy/envconfig"
	"github.com/JamesStewy/envconfig/docs"
	"html/template"
	"os"
)

func ExampleTextTable() {
	var conf struct {
		Protocol   string `envconfig:"default=https,note=Protocol to be used"`
		RemoteHost string `envconfig:"note=Remote hostname"`
		Port       int    `envconfig:"default=443"`
	}

	os.Setenv("REMOTE_HOST", "localhost")
	os.Setenv("PORT", "80")

	cinfo, err := envconfig.Parse(&conf)
	if err != nil {
		panic(err)
	}

	if err = cinfo.Read(); err != nil {
		panic(err)
	}

	docs.TextTable(os.Stdout, cinfo)
	// Output:
	// +------------------------+-----------+---------+---------------------+
	// |          KEYS          |   VALUE   | DEFAULT |        NOTE         |
	// +------------------------+-----------+---------+---------------------+
	// | PROTOCOL               | https     | https   | Protocol to be used |
	// +------------------------+-----------+---------+---------------------+
	// | REMOTEHOST             | localhost |         | Remote hostname     |
	// | REMOTE_HOST            |           |         |                     |
	// +------------------------+-----------+---------+---------------------+
	// | PORT                   | 80        | 443     |                     |
	// +------------------------+-----------+---------+---------------------+
}

func ExampleHTMLTable() {
	var conf struct {
		Protocol   string `envconfig:"default=https,note=Protocol to be used"`
		RemoteHost string `envconfig:"note=Remote hostname"`
		Port       int    `envconfig:"default=443"`
	}

	os.Setenv("REMOTE_HOST", "localhost")
	os.Setenv("PORT", "80")

	cinfo, err := envconfig.Parse(&conf)
	if err != nil {
		panic(err)
	}

	if err = cinfo.Read(); err != nil {
		panic(err)
	}

	if err = docs.HTMLTable(os.Stdout, cinfo); err != nil {
		panic(err)
	}
	// Output:
	// <table>
	// 	<thead>
	// 		<tr>
	// 			<th>Keys</th>
	// 			<th>Value</th>
	// 			<th>Default</th>
	// 			<th>Note</th>
	// 		</tr>
	// 	</thead>
	// 	<tbody>
	// 		<tr>
	// 			<th>PROTOCOL</th>
	// 			<th>https</th>
	// 			<th>https</th>
	// 			<th>Protocol to be used</th>
	// 		</tr>
	// 		<tr>
	// 			<th>REMOTEHOST<br>REMOTE_HOST</th>
	// 			<th>localhost</th>
	// 			<th></th>
	// 			<th>Remote hostname</th>
	// 		</tr>
	// 		<tr>
	// 			<th>PORT</th>
	// 			<th>80</th>
	// 			<th>443</th>
	// 			<th></th>
	// 		</tr>
	// 	</tbody>
	// </table>
}

func ExampleHTMLTableWithTemplate() {
	var conf struct {
		Protocol   string `envconfig:"default=https,note=Protocol to be used"`
		RemoteHost string `envconfig:"note=Remote hostname"`
		Port       int    `envconfig:"default=443"`
	}

	os.Setenv("REMOTE_HOST", "localhost")
	os.Setenv("PORT", "80")

	cinfo, err := envconfig.Parse(&conf)
	if err != nil {
		panic(err)
	}

	if err = cinfo.Read(); err != nil {
		panic(err)
	}

	base_tmpl := template.Must(template.New("base").Parse(`<html>
	<head>
		<title>Test Page</title>
	</head>
	<body>
{{template "envconfig" .}}
	</body>
</html>`))

	t, err := docs.HTMLTableWithTemplate(base_tmpl)
	if err != nil {
		panic(err)
	}

	err = t.Execute(os.Stdout, cinfo)
	if err != nil {
		panic(err)
	}
	// Output:
	// <html>
	// 	<head>
	// 		<title>Test Page</title>
	// 	</head>
	// 	<body>
	// <table>
	// 	<thead>
	// 		<tr>
	// 			<th>Keys</th>
	// 			<th>Value</th>
	// 			<th>Default</th>
	// 			<th>Note</th>
	// 		</tr>
	// 	</thead>
	// 	<tbody>
	// 		<tr>
	// 			<th>PROTOCOL</th>
	// 			<th>https</th>
	// 			<th>https</th>
	// 			<th>Protocol to be used</th>
	// 		</tr>
	// 		<tr>
	// 			<th>REMOTEHOST<br>REMOTE_HOST</th>
	// 			<th>localhost</th>
	// 			<th></th>
	// 			<th>Remote hostname</th>
	// 		</tr>
	// 		<tr>
	// 			<th>PORT</th>
	// 			<th>80</th>
	// 			<th>443</th>
	// 			<th></th>
	// 		</tr>
	// 	</tbody>
	// </table>
	// 	</body>
	// </html>
}
