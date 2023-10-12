package gojsox

import (
	"encoding/json"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// func ParseGlob(pattern string) error {
// 	filenames, err := filepath.Glob(pattern)
// 	if err != nil {
// 		return err
// 	}
// 	if len(filenames) == 0 {
// 		return fmt.Errorf("gojsox: pattern matches no files: %#q", pattern)
// 	}
// 	return ParseFiles(filenames...)
// }

// func ParseFiles(filenames ...string) error {
// 	for _, filename := range filenames {
// 		err := ParseFile(filename)
// 	}
// 	return nil
// }

func ParseFile(template *template.Template, filename string) (string, *template.Template, error) {
	ext := filepath.Ext(filename)
	name := filepath.Base(filename)
	raw, err := os.ReadFile(filename)
	if err != nil {
		return name, template, fmt.Errorf("gojsox: failed reading file %s: %w", filename, err)
	}
	var temp string
	switch ext {
	case ".json":
		temp, err = processJson(name, raw)
	case ".yaml", ".yml":
		temp, err = processYaml(name, raw)
	default:
		return name, template, fmt.Errorf("gojsox: unknown extension %s: %w", ext, err)
	}
	if err != nil {
		return name, template, fmt.Errorf("gojsox: failed to process file %s: %w", filename, err)
	}

	os.WriteFile("template.html", []byte(temp), os.ModePerm)

	template, err = template.Clone()
	if err != nil {
		return name, template, fmt.Errorf("gojsox: failed to copy initial template: %w", err)
	}

	template, err = template.Parse(temp)
	if err != nil {
		return name, template, fmt.Errorf("gojsox: failed to parse generated template data %s: %w", filename, err)
	}
	return name, template, nil
}

func processJson(name string, raw []byte) (string, error) {
	var parsed map[string]any
	err := json.Unmarshal(raw, &parsed)
	if err != nil {
		return "", err
	}
	return createTemplate(name, parsed)
}

func processYaml(name string, raw []byte) (string, error) {
	var parsed map[string]any
	err := yaml.Unmarshal(raw, &parsed)
	if err != nil {
		return "", err
	}
	return createTemplate(name, parsed)
}

func createTemplate(name string, data map[string]any) (string, error) {
	var w strings.Builder

	w.WriteString(`{{define "content"}}`)
	w.WriteString(`<div id="content">`)
	if title, ok := data["title"].(string); ok {
		w.WriteString(`<h2>`)
		w.WriteString(title)
		w.WriteString(`</h2>`)
	}

	if description, ok := data["description"].(string); ok {
		w.WriteString(`<h4>`)
		w.WriteString(description)
		w.WriteString(`</h4>`)
	}

	w.WriteString(`<form hx-target="#response" hx-post="/`)
	w.WriteString(name)
	w.WriteString(`">`)
	err := processNative(&w, data, processingContext{IsRoot: true})
	w.WriteString(`<button class="btn btn-default">Submit</button>`)
	w.WriteString(`</form>`)
	w.WriteString(`<div id="response"></div>`)
	w.WriteString(`</div>`)
	w.WriteString(`{{end}}`)
	return w.String(), err
}

type processingContext struct {
	Name   string
	Prefix []string
	IsRoot bool
}

func (p processingContext) Path() []string {
	if p.IsRoot {
		return p.Prefix
	}
	return append(p.Prefix, p.Name)
}

func (p processingContext) JoinedPath(sep string) string {
	return strings.Join(p.Path(), sep)
}

// processNative accepts a go native map implementation of a json schema untyped object.
func processNative(w *strings.Builder, data map[string]any, ctx processingContext) error {
	typ, ok := data["type"].(string)
	if !ok {
		return fmt.Errorf("unexpected/missing type definition")
	}

	switch typ {
	case "object":
		return processObject(w, data, ctx)
	case "string":
		return processString(w, data, ctx)
	default:
		return fmt.Errorf("unexpected type %s", typ)
	}
}

// processObject converts a json schema object to template.
func processObject(w *strings.Builder, data map[string]any, ctx processingContext) error {
	properties, ok := data["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("unexpected/missing properties definition")
	}

	keys := make([]string, len(properties))
	i := 0
	for key := range properties {
		keys[i] = key
		i++
	}

	// TODO: sorting
	sort.Strings(keys)

	w.WriteString(`<div class="form-group pure-form-stacked">`)
	if !ctx.IsRoot {
		w.WriteString("<fieldset>")
	}
	if ctx.Name != "" {
		w.WriteString(`<legend>`)
		w.WriteString(strings.ToUpper(ctx.Name))
		w.WriteString(`</legend>`)
	}
	for _, key := range keys {
		typed, ok := properties[key].(map[string]any)
		if !ok {
			return fmt.Errorf("failed to process %s", key)
		}
		err := processNative(w, typed, processingContext{Name: key, Prefix: ctx.Path()})
		if err != nil {
			return fmt.Errorf("failed to process %s: %w", key, err)
		}
	}
	if !ctx.IsRoot {
		w.WriteString("</fieldset>")
	}
	w.WriteString(`</div>`)

	return nil
}

// processString converts a json schema string to template.
func processString(w *strings.Builder, data map[string]any, ctx processingContext) error {
	formInput{
		label:       strings.ToUpper(ctx.Name),
		placeholder: ctx.Name,
		path:        strings.Join(ctx.Path(), "."),
		typ:         "text",
	}.Write(w)

	return nil
}
