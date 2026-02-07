package kibble

import "embed"

//go:embed web/static
var StaticFS embed.FS

//go:embed web/templates
var TemplateFS embed.FS

//go:embed themes.yaml
var ThemesYAML []byte
