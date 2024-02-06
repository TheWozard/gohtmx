package gohtmx

import (
	"fmt"
)

// Generator used to generate content for building templates.
type Generator interface {
	NewID(group string) string
}

func NewDefaultGenerator() Generator {
	return &IterGenerator{}
}

type IterGenerator struct {
	Index map[string]int64
}

func (b *IterGenerator) NewID(prefix string) string {
	if b.Index == nil {
		b.Index = map[string]int64{}
	}
	value := b.Index[prefix]
	b.Index[prefix] = value + 1
	return fmt.Sprintf("%s_%d", prefix, value)
}
