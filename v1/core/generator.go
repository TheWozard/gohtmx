package core

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"
)

// Generator used to generate content for building templates.
type Generator interface {
	NewGroupID(g string) string
	NewFunctionID(f any) string
}

func NewDefaultGenerator() Generator {
	return &IterGenerator{index: map[string]int{}}
}

type IterGenerator struct {
	index map[string]int
}

func (b *IterGenerator) NewFunctionID(f any) string {
	group := strings.Split(runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name(), "/")
	name := group[len(group)-1]
	return b.NewGroupID(strings.NewReplacer(".", "_", "-", "_").Replace(name))
}

func (b *IterGenerator) NewGroupID(group string) string {
	num := b.index[group]
	b.index[group] = num + 1
	return fmt.Sprintf("%s_%d", group, num)
}
