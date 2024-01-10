package core

import (
	"context"
)

type dataKeyType int

const dataKey dataKeyType = iota

// TemplateData defines data intended to be used to render templates defined by components.
type TemplateData map[string]any

// DataFromContext extracts the current available data from the context.
// This is used as the data for rendering templates.
func DataFromContext(ctx context.Context) TemplateData {
	if data, ok := ctx.Value(dataKey).(TemplateData); ok {
		return data
	}
	return TemplateData{}
}

// Context adds the passed map into the current data map on the context.
func (t TemplateData) Context(ctx context.Context) context.Context {
	return context.WithValue(ctx, dataKey, DataFromContext(ctx).Merge(t))
}

// Merge merges two TemplateData maps together. The addition map will overwrite any existing keys.
func (t TemplateData) Merge(addition TemplateData) TemplateData {
	if len(t) == 0 {
		return addition
	}
	for k, v := range addition {
		t[k] = v
	}
	return t
}
