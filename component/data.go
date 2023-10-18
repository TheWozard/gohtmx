package component

import "context"

type dataKeyType int

const dataKey dataKeyType = iota

func DataOnContext(ctx context.Context, data any) context.Context {
	return context.WithValue(ctx, dataKey, data)
}

func DataFromContext(ctx context.Context) any {
	return ctx.Value(dataKey)
}
