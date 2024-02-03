package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TheWozard/gohtmx"
	"github.com/TheWozard/gohtmx/example/forms/components"
	"github.com/gorilla/mux"
)

// go:embed assets/*
// var assets embed.FS

func main() {
	store := Store{Path: "./example/forms/store"}

	f := gohtmx.NewPage()
	f.Add(
		gohtmx.Document{
			Header: gohtmx.Fragment{
				gohtmx.Raw(`<meta charset="utf-8">`),
				gohtmx.Raw(`<meta name="viewport" content="width=device-width, initial-scale=1">`),
				gohtmx.Raw(`<title>Form Inputs</title>`),
				gohtmx.Raw(`<link rel="stylesheet" href="/assets/style.css">`),
				gohtmx.Raw(`<script src="https://unpkg.com/htmx.org@1.9.6/dist/htmx.min.js"></script>`),
				gohtmx.Raw(`<script src="https://unpkg.com/idiomorph/dist/idiomorph-ext.min.js"></script>`),
				gohtmx.Raw(`<script defer src="/assets/script.js"></script>`),
			},
			Body: Body(store),
		},
	)
	h, err := f.Build()
	if err != nil {
		log.Fatal(err)
	}

	mux := mux.NewRouter()
	// Embedding assets enables a single binary to be distributed, however during development, http.Dir enables updates without recompiling.
	// mux.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	mux.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./example/forms/assets"))))
	mux.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate Delay
		// time.Sleep(200 * time.Millisecond)
		h.ServeHTTP(w, r)
	})

	log.Default().Println("staring server at http://localhost:8080")
	log.Fatal((&http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}).ListenAndServe())
}

func Body(store Store) gohtmx.Component {

	return gohtmx.Div{
		Content: gohtmx.Fragment{
			&components.Counter{Min: 0, Max: 10, Step: 2, Count: 0},
		},
	}
}

// Defines a simple persistent file store for getting and setting json data.
type Store struct {
	Path string
}

func (s Store) File(name string) string {
	return s.Path + "/" + name + ".json"
}

func (s Store) Get(name string) (map[string]any, error) {
	data, err := os.ReadFile(s.File(name))
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	result := map[string]any{}
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal file: %w", err)
	}
	return result, nil
}

func (s Store) Set(name string, data map[string]any) error {
	raw, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal file: %w", err)
	}
	err = os.WriteFile(s.File(name), raw, os.ModePerm)
	if err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}
	return nil
}

func (s Store) List() []string {
	files, err := os.ReadDir(s.Path)
	if err != nil {
		return nil
	}
	result := []string{}
	for _, file := range files {
		if file.IsDir() {
			continue
		}
		name := file.Name()
		if strings.HasSuffix(name, ".json") {
			result = append(result, name[:len(name)-5])
		}
	}
	return result
}
