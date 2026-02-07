package config

import (
	"fmt"
	"math"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Theme defines a named color theme loaded from themes.yaml.
type Theme struct {
	ID     string      `yaml:"id"`
	Name   string      `yaml:"name"`
	Scheme string      `yaml:"scheme"` // "dark" or "light" — sets browser color-scheme
	Logo   string      `yaml:"logo"`   // "light" = white logo, "dark" = black logo
	Colors ThemeColors `yaml:"colors"`
}

// ThemeColors holds the 6 core colors that define a theme.
type ThemeColors struct {
	Background string `yaml:"background"` // page background
	Surface    string `yaml:"surface"`    // card/panel backgrounds
	Navbar     string `yaml:"navbar"`     // navigation bar background
	Primary    string `yaml:"primary"`    // buttons, links, active accents
	Accent     string `yaml:"accent"`     // highlights, secondary accents
	Text       string `yaml:"text"`       // main body text
}

type themesFile struct {
	Themes []Theme `yaml:"themes"`
}

// LoadThemes reads themes from a YAML file. Falls back to defaults if file missing.
func LoadThemes(path string) ([]Theme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return DefaultThemes(), nil
		}
		return nil, fmt.Errorf("read themes file: %w", err)
	}

	var f themesFile
	if err := yaml.Unmarshal(data, &f); err != nil {
		return nil, fmt.Errorf("parse themes file: %w", err)
	}

	if len(f.Themes) == 0 {
		return DefaultThemes(), nil
	}
	return f.Themes, nil
}

// DefaultThemes returns the built-in themes used when no themes.yaml is present.
func DefaultThemes() []Theme {
	return []Theme{
		{
			ID: "pastel-earth", Name: "Pastel Earth", Scheme: "light", Logo: "light",
			Colors: ThemeColors{
				Background: "#E0DDCE", Surface: "#CDE4E1", Navbar: "#57664A",
				Primary: "#AACF9F", Accent: "#BCE6DC", Text: "#2D323B",
			},
		},
		{
			ID: "light-and-fresh", Name: "Light and Fresh", Scheme: "light", Logo: "light",
			Colors: ThemeColors{
				Background: "#d1d8da", Surface: "#d3d9d9", Navbar: "#2D323B",
				Primary: "#3c6c60", Accent: "#668d7d", Text: "#2D323B",
			},
		},
		{
			ID: "soft-dark", Name: "Soft Dark", Scheme: "dark", Logo: "dark",
			Colors: ThemeColors{
				Background: "#222b35", Surface: "#222b35", Navbar: "#222b35",
				Primary: "#447163", Accent: "#567E75", Text: "#e9efea",
			},
		},
		{
			ID: "bright-ocean", Name: "Bright Ocean", Scheme: "light", Logo: "light",
			Colors: ThemeColors{
				Background: "#eeebf0", Surface: "#e2e7f0", Navbar: "#6F94C0",
				Primary: "#a6c8e8", Accent: "#a4d6ee", Text: "#222b35",
			},
		},
	}
}

// ResolveThemeCSS takes a Theme and returns a CSS string of custom property declarations
// suitable for injection inside :root { ... }.
func ResolveThemeCSS(t Theme) string {
	c := t.Colors
	bg := mustParseHex(c.Background)
	surface := mustParseHex(c.Surface)
	navbar := mustParseHex(c.Navbar)
	text := mustParseHex(c.Text)

	// Derive secondary colors.
	// For hover/input, use lighten/darken when surface and bg are the same.
	var surfaceHover, inputBg rgb
	if t.Scheme == "dark" {
		surfaceHover = lighten(surface, 0.08)
		inputBg = lighten(bg, 0.04)
	} else {
		surfaceHover = darken(surface, 0.06)
		inputBg = lighten(bg, 0.03)
	}
	border := blendColors(surface, text, 0.20)
	textMuted := blendColors(text, bg, 0.40)
	textHeading := contrastPush(text, bg, 0.10)

	// Shadows based on background luminance
	shadowColor := darken(bg, 0.5)
	shadowAlpha := 0.15
	shadowAlphaLg := 0.20
	if t.Scheme == "dark" {
		shadowAlpha = 0.30
		shadowAlphaLg = 0.35
	}

	var b strings.Builder
	writeProp := func(name, value string) {
		fmt.Fprintf(&b, "--%s: %s; ", name, value)
	}

	writeProp("bg", c.Background)
	writeProp("bg-surface", c.Surface)
	writeProp("bg-surface-hover", hexString(surfaceHover))
	writeProp("bg-input", hexString(inputBg))
	writeProp("bg-navbar", c.Navbar)
	writeProp("border", hexString(border))
	writeProp("border-focus", c.Primary)
	writeProp("primary", c.Primary)
	writeProp("secondary", c.Accent)
	writeProp("text", c.Text)
	writeProp("text-muted", hexString(textMuted))
	writeProp("text-heading", hexString(textHeading))

	// Navbar text color: pick white or black based on navbar luminance
	navbarTextColor := "#ffffff"
	if luminance(navbar) > 0.5 {
		navbarTextColor = "#2D323B"
	}
	writeProp("navbar-text", navbarTextColor)

	writeProp("shadow", fmt.Sprintf("0 2px 4px rgba(%d, %d, %d, %.2f), 0 8px 24px rgba(0, 0, 0, %.2f)",
		shadowColor.r, shadowColor.g, shadowColor.b, shadowAlpha, shadowAlpha))
	writeProp("shadow-lg", fmt.Sprintf("0 4px 8px rgba(%d, %d, %d, %.2f), 0 16px 40px rgba(0, 0, 0, %.2f)",
		shadowColor.r, shadowColor.g, shadowColor.b, shadowAlphaLg, shadowAlphaLg))
	writeProp("shadow-nav", fmt.Sprintf("0 2px 12px rgba(0, 0, 0, %.2f)", shadowAlpha))

	writeProp("card-radius", "12px")
	writeProp("btn-radius", "8px")
	writeProp("input-radius", "8px")
	fmt.Fprintf(&b, "color-scheme: %s;", t.Scheme)

	return b.String()
}

// --- Color math helpers ---

type rgb struct {
	r, g, b uint8
}

func mustParseHex(hex string) rgb {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	var r, g, b uint8
	fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	return rgb{r, g, b}
}

func hexString(c rgb) string {
	return fmt.Sprintf("#%02x%02x%02x", c.r, c.g, c.b)
}

func blendColors(c1, c2 rgb, ratio float64) rgb {
	return rgb{
		r: uint8(float64(c1.r)*(1-ratio) + float64(c2.r)*ratio),
		g: uint8(float64(c1.g)*(1-ratio) + float64(c2.g)*ratio),
		b: uint8(float64(c1.b)*(1-ratio) + float64(c2.b)*ratio),
	}
}

func luminance(c rgb) float64 {
	// Relative luminance (simplified sRGB)
	r := float64(c.r) / 255.0
	g := float64(c.g) / 255.0
	b := float64(c.b) / 255.0
	return 0.2126*r + 0.7152*g + 0.0722*b
}

func darken(c rgb, amount float64) rgb {
	return rgb{
		r: uint8(math.Max(0, float64(c.r)*(1-amount))),
		g: uint8(math.Max(0, float64(c.g)*(1-amount))),
		b: uint8(math.Max(0, float64(c.b)*(1-amount))),
	}
}

func lighten(c rgb, amount float64) rgb {
	return rgb{
		r: uint8(math.Min(255, float64(c.r)+amount*255)),
		g: uint8(math.Min(255, float64(c.g)+amount*255)),
		b: uint8(math.Min(255, float64(c.b)+amount*255)),
	}
}

// contrastPush moves text further from background for headings.
func contrastPush(text, bg rgb, amount float64) rgb {
	if luminance(bg) > 0.5 {
		// Light bg: darken text
		return darken(text, amount)
	}
	// Dark bg: lighten text
	return rgb{
		r: uint8(math.Min(255, float64(text.r)*(1+amount))),
		g: uint8(math.Min(255, float64(text.g)*(1+amount))),
		b: uint8(math.Min(255, float64(text.b)*(1+amount))),
	}
}
