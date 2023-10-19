package gohtmx

import (
	"context"
	"net/http"
)

type dataKeyType int

const dataKey dataKeyType = iota

type TemplateData struct {
	Path string
}

func TemplateDataOnContext(ctx context.Context, data TemplateData) context.Context {
	return context.WithValue(ctx, dataKey, data)
}

func TemplateDataFromRequestOnContext(ctx context.Context, r *http.Request) context.Context {
	return context.WithValue(ctx, dataKey, TemplateData{
		Path: r.URL.Path,
	})
}

func DataFromContext(ctx context.Context) any {
	if data, ok := ctx.Value(dataKey).(TemplateData); ok {
		return data
	}
	return TemplateData{}
}
