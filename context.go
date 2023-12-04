package gohtmx

import (
	"bytes"
	"context"
	"fmt"
	"html/template"
	"net/http"
	"strings"
)

type dataKeyType int

const dataKey dataKeyType = iota

var NoTemplateData = fmt.Errorf("no data to load")

// TemplateData defines data intended to be used to render templates defined by components.
type TemplateData map[string]any

// MergeDataToContext adds the passed map into the current data map on the context.
func MergeDataToContext(ctx context.Context, data TemplateData) context.Context {
	current := DataFromContext(ctx)
	if len(current) == 0 {
		current = data
	} else {
		for k, v := range data {
			current[k] = v
		}
	}
	return context.WithValue(ctx, dataKey, current)
}

// DataFromContext extracts the current available data from the context.
// This is used as the data for rendering templates
func DataFromContext(ctx context.Context) TemplateData {
	if data, ok := ctx.Value(dataKey).(TemplateData); ok {
		return data
	}
	return TemplateData{}
}

// TemplateDataFromRequest converts an *http.Request into the default TemplateData info.
func TemplateDataFromRequest(r *http.Request) TemplateData {
	return TemplateData{
		"path": r.URL.Path,
	}
}

type TemplateDataLoaderFunc func(r *http.Request) (TemplateData, error)

// TemplateDataLoader defines potential data to load to properly rendered a Component.
type TemplateDataLoader struct {
	Loaders []TemplateDataLoaderFunc
}

func (t *TemplateDataLoader) Add(f TemplateDataLoaderFunc) {
	t.Loaders = append(t.Loaders, f)
}

func (d *TemplateDataLoader) AddNamedAtPath(name, path string, load TemplateDataLoaderFunc) {
	d.Add(func(r *http.Request) (TemplateData, error) {
		if strings.HasPrefix(r.URL.Path, path) {
			data, err := load(r)
			return TemplateData{name: data}, err
		}
		return nil, NoTemplateData
	})
}

func (d *TemplateDataLoader) LoadContext(ctx context.Context, r *http.Request) context.Context {
	for _, loader := range d.Loaders {
		if data, err := loader(r); err == nil {
			ctx = MergeDataToContext(ctx, data)
		}
	}
	return ctx
}

func LoadTemplateDataFromQuery(params ...string) TemplateDataLoaderFunc {
	return func(r *http.Request) (TemplateData, error) {
		data := TemplateData{}
		for key, value := range r.URL.Query() {
			data[key] = value[0]
		}
		return data, nil
	}
}

func LoadTemplateDataFromForm(params ...string) TemplateDataLoaderFunc {
	return func(r *http.Request) (TemplateData, error) {
		data := TemplateData{}
		err := r.ParseForm()
		if err != nil {
			return data, NoTemplateData
		}
		for key, value := range r.Form {
			data[key] = value[0]
		}
		return data, nil
	}
}

// TemplateHandler wrapper for making a template into an http.Handler
type TemplateHandler struct {
	Template *template.Template
	Loader   *TemplateDataLoader
}

func (th TemplateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := th.Template.Execute(w, DataFromContext(th.Loader.LoadContext(r.Context(), r)))
	if err != nil {
		http.Error(w, "error rendering template", http.StatusInternalServerError)
		return
	}
}

func (th TemplateHandler) ServeHTTPWithExtraData(w http.ResponseWriter, r *http.Request, d TemplateData) {
	ctx := th.Loader.LoadContext(MergeDataToContext(r.Context(), d), r)
	buffer := bytes.Buffer{}
	err := th.Template.Execute(&buffer, DataFromContext(ctx))
	if err != nil {
		http.Error(w, fmt.Sprintf("error rendering template: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	_, _ = w.Write(buffer.Bytes())
}
