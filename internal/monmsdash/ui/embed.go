package ui

import "embed"

//go:embed static/monms-dash.css static/components.css static/alpine.min.js static/htmx.min.js static/fonts static/js
var StaticFS embed.FS

//go:embed templates/*
var TemplatesFS embed.FS
