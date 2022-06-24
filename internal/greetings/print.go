package greetings

import (
	"os"
	"text/template"
)

const greetingsTemplate = `
Build version: <{{.BuildVersion}}>
Build date: <{{.BuildDate}}>
Build commit: <{{.BuildCommit}}>
`

// Greetings represents data for template
type Greetings struct {
	BuildVersion string
	BuildDate    string
	BuildCommit  string
}

// Print prints template to stdout
func Print(buildVersion, buildDate, buildCommit string) error {
	greetings := Greetings{
		BuildVersion: buildVersion,
		BuildDate:    buildDate,
		BuildCommit:  buildCommit,
	}

	if greetings.BuildVersion == "" {
		greetings.BuildVersion = "N/A"
	}
	if greetings.BuildDate == "" {
		greetings.BuildDate = "N/A"
	}
	if greetings.BuildCommit == "" {
		greetings.BuildCommit = "N/A"
	}

	tmpl := template.Must(template.New("greetings").Parse(greetingsTemplate))

	return tmpl.Execute(os.Stdout, greetings)
}
