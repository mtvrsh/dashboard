// Package ui holds the shared dashboard assets, embedded once so both the
// server and the standalone demo binary render the same page.
package ui

import _ "embed"

//go:embed index.template
var Index string

//go:embed style.css
var CSS []byte
