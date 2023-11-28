package gohtmx

import (
	"context"
	"net/http"
)

type dataKeyType int

// TemplateData defines data intended to be used to render templates defined by components.
type TemplateData map[string]any

const dataKey dataKeyType = iota

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
