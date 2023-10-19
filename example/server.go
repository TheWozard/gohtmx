package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
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
			DefaultRedirect: "One",
			Tabs: []gohtmx.Tab{
				{
					Value: "One",
					Tag:   gohtmx.Raw("One"),
					Contents: gohtmx.Stream{
						ID: "stream",
						SSEEventGenerator: func(ctx context.Context, c chan gohtmx.SSEEvent) {
							t := time.NewTicker(1 * time.Second)
							c <- gohtmx.SSEEvent{
								Data: gohtmx.Raw(time.Now().Format(time.RFC3339)),
							}
							for {
								select {
								case <-ctx.Done():
									return
								case <-t.C:
									c <- gohtmx.SSEEvent{
										Data: gohtmx.Raw(time.Now().Format(time.RFC3339)),
									}
								}
							}
						},
						Content: gohtmx.StreamTarget{},
					},
				},
				{
					Value:    "Two",
					Tag:      gohtmx.Raw("Two"),
					Contents: gohtmx.Raw("Tab Two"),
				},
				{
					Value: "Three",
					Tag:   gohtmx.Raw("Three"),
					Contents: gohtmx.Tabs{
						ID:              "tabs2",
						Classes:         []string{"tabs", "is-centered"},
						ActiveClasses:   []string{"is-active"},
						DefaultRedirect: "Foo",
						Tabs: []gohtmx.Tab{
							{
								Value:    "Foo",
								Tag:      gohtmx.Raw("Foo"),
								Contents: gohtmx.Raw("Foo"),
							},
							{
								Value:    "Bar",
								Tag:      gohtmx.Raw("Bar"),
								Contents: gohtmx.Raw("Bar"),
							},
							{
								Value:    "Foobar",
								Tag:      gohtmx.Raw("Both"),
								Contents: gohtmx.Raw("Foobar"),
							},
						},
					},
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
