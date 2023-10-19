package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/TheWozard/gohtmx"
)

func main() {
	// store := FileStore{BasePath: "./data/"}

	mux := http.NewServeMux()
	err := gohtmx.ServeComponent("/", mux, gohtmx.Document{
		Header: gohtmx.Fragment{
			gohtmx.Raw(`<meta charset="utf-8">`),
			gohtmx.Raw(`<meta name="viewport" content="width=device-width, initial-scale=1">`),
			gohtmx.Raw(`<title>Example</title>`),
			gohtmx.Raw(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">`),
			gohtmx.Raw(`<script src="https://unpkg.com/htmx.org@1.9.6/dist/htmx.min.js"></script>`),
			gohtmx.Raw(`<script src="https://unpkg.com/htmx.org@1.9.6/dist/ext/sse.js"></script>`),
		},
		Body: gohtmx.Tabs{
			ID:              "tabs",
			Classes:         []string{"tabs", "is-centered"},
			ActiveClasses:   []string{"is-active"},
			DefaultRedirect: "stream",
			Tabs: []gohtmx.Tab{
				{
					Value:   "stream",
					Tag:     gohtmx.Raw("Stream"),
					Content: StreamTab(),
				},
				{
					Value:   "form",
					Tag:     gohtmx.Raw("Form"),
					Content: Box(gohtmx.Raw("TODO")),
				},
				{
					Value:   "recursive",
					Tag:     gohtmx.Raw("Recursive"),
					Content: RecursiveTab(),
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("staring server at http://localhost:8080")
	log.Fatal((&http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}).ListenAndServe())
}

func StreamTab() gohtmx.Component {
	start := time.Now()
	return gohtmx.Stream{
		ID: "stream",
		SSEEventGenerator: func(ctx context.Context, c chan gohtmx.SSEEvent) {
			rollTicker := time.NewTicker(5 * time.Second)
			uptimeTicker := time.NewTicker(1 * time.Second)
			memTicker := time.NewTicker(3 * time.Second)
			r := rand.New(rand.NewSource(time.Now().Unix()))
			roll := func() {
				roll := r.Intn(20) + 1
				color := "light"
				if roll == 20 {
					color = "success"
				} else if roll == 1 {
					color = "danger"
				}
				c <- gohtmx.SSEEvent{
					Event: "roll",
					Data:  Tag("Roll", fmt.Sprintf("%.d", roll), color),
				}
			}
			tick := func() {
				c <- gohtmx.SSEEvent{
					Event: "uptime",
					Data:  Tag("Uptime", fmt.Sprintf("%.fs", time.Since(start).Seconds()), "light"),
				}
			}
			mem := func() {
				var m runtime.MemStats
				runtime.ReadMemStats(&m)

				c <- gohtmx.SSEEvent{
					Event: "memory",
					Data:  Tag("Memory", fmt.Sprintf("%.4f MiB", float64(m.Alloc)/1024./1024.), "light"),
				}
			}
			roll()
			tick()
			mem()
			for {
				select {
				case <-ctx.Done():
					return
				case <-rollTicker.C:
					roll()
				case <-uptimeTicker.C:
					tick()
				case <-memTicker.C:
					mem()
				}
			}
		},
		Content: Box(Content(gohtmx.Fragment{
			gohtmx.Tag{Name: "h3", Content: gohtmx.Raw("Streams")},
			gohtmx.Tag{Name: "p", Content: gohtmx.Raw("The following is an example of using Server Side Event streams to create self updating UIs. Each of the following tags are updated independently and push as a SSE from the server.")},
			gohtmx.Tag{Name: "div", Attributes: []gohtmx.Attribute{
				{Name: "class", Value: "field is-grouped is-grouped-multiline"},
			},
				Content: gohtmx.Fragment{
					gohtmx.StreamTarget{Events: []string{"roll"}, Classes: []string{"control"}, Content: Tag("Loading", "0", "light")},
					gohtmx.StreamTarget{Events: []string{"uptime"}, Classes: []string{"control"}, Content: Tag("Loading", "0", "light")},
					gohtmx.StreamTarget{Events: []string{"memory"}, Classes: []string{"control"}, Content: Tag("Loading", "0", "light")},
				}},
		})),
	}
}

func RecursiveTab() gohtmx.Component {
	return gohtmx.Tabs{
		ID:              "tabs-recursive",
		Classes:         []string{"tabs", "is-centered"},
		ActiveClasses:   []string{"is-active"},
		DefaultRedirect: "foo",
		Tabs: []gohtmx.Tab{
			{
				Value:   "foo",
				Tag:     gohtmx.Raw("Foo"),
				Content: Box(gohtmx.Raw("Foo")),
			},
			{
				Value:   "bar",
				Tag:     gohtmx.Raw("Bar"),
				Content: Box(gohtmx.Raw("Bar")),
			},
			{
				Value:   "foobar",
				Tag:     gohtmx.Raw("Both"),
				Content: Box(gohtmx.Raw("Foobar")),
			},
		},
	}
}

func Box(c gohtmx.Component) gohtmx.Component {
	return gohtmx.Tag{
		Name: "div",
		Attributes: []gohtmx.Attribute{
			{Name: "class", Value: "box"},
			{Name: "style", Value: "max-width: 500px;margin: auto;"},
		},
		Content: c,
	}
}

func Content(c gohtmx.Component) gohtmx.Component {
	return gohtmx.Tag{
		Name: "div",
		Attributes: []gohtmx.Attribute{
			{Name: "class", Value: "content"},
		},
		Content: c,
	}
}

func Tag(prefix, suffix, color string) gohtmx.Component {
	return gohtmx.Tag{
		Name: "div", Attributes: []gohtmx.Attribute{
			{Name: "class", Value: "tags has-addons"},
		},
		Content: gohtmx.Fragment{
			gohtmx.Tag{
				Name:       "span",
				Attributes: []gohtmx.Attribute{{Name: "class", Value: "tag is-dark"}},
				Content:    gohtmx.Raw(prefix),
			},
			gohtmx.Tag{
				Name:       "span",
				Attributes: []gohtmx.Attribute{{Name: "class", Value: "tag is-" + color}},
				Content:    gohtmx.Raw(suffix),
			},
		},
	}
}

type Store interface {
	Get(name string) (any, error)
	Set(name string, data any) error
}

type FileStore struct {
	BasePath string
}

func (fs FileStore) Path(name string) string {
	return filepath.Join(fs.BasePath, name+".json")
}

func (fs FileStore) Get(name string) (any, error) {
	filename := fs.Path(name)

	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var data any
	err = json.Unmarshal(raw, &data)
	if err != nil {
		return nil, fmt.Errorf("file store: failed to decode yaml %s: %w", filename, err)
	}
	return data, nil
}

func (fs FileStore) Set(name string, data any) error {
	filename := fs.Path(name)

	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("file store: failed to encode yaml %s: %w", filename, err)
	}
	err = os.WriteFile(filename, raw, os.ModePerm)
	if err != nil {
		return fmt.Errorf("file store: failed to write file %s: %w", filename, err)
	}
	return nil
}
