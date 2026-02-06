package kibble

import "embed"

//go:embed web/static
var StaticFS embed.FS

//go:embed web/templates
var TemplateFS embed.FS
