package ui

import (
	"bytes"
	"fmt"
	"html/template"
	"io"

	"github.com/mtvrsh/dashboard/api"
)

var page = template.Must(template.New("index.template").Parse(Index))

const styleLink = `<link rel="stylesheet" href="/style.css">`

type pageData struct {
	api.SystemInfo
	StyleTag template.HTML
}

// Render writes the dashboard page for info to w.
func Render(w io.Writer, info api.SystemInfo) error {
	return page.Execute(w, pageData{SystemInfo: info, StyleTag: template.HTML(styleLink)})
}

// RenderStandalone returns a complete HTML document with the stylesheet
// inlined, so it can be opened directly from disk.
func RenderStandalone(info api.SystemInfo) ([]byte, error) {
	tag := template.HTML("<style>\n" + string(CSS) + "\n</style>")
	var b bytes.Buffer
	if err := page.Execute(&b, pageData{SystemInfo: info, StyleTag: tag}); err != nil {
		return nil, fmt.Errorf("render page: %w", err)
	}
	return b.Bytes(), nil
}
