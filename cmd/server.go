package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/TheWozard/gojsox/component"
	"gopkg.in/yaml.v2"
)

func main() {
	// store := FileStore{BasePath: "./data/"}

	mux := http.NewServeMux()
	err := component.LoadComponent(mux, component.Page{
		Header: component.Fragment{
			component.Raw(`<script src="https://unpkg.com/htmx.org@1.9.6"></script>`),
			component.Raw(`<script src="https://unpkg.com/htmx.org/dist/ext/sse.js"></script>`),
		},
		Body: nil,
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
	return filepath.Join(fs.BasePath, name+".yaml")
}

func (fs FileStore) Get(name string) (any, error) {
	filename := fs.Path(name)

	raw, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	var data any
	err = yaml.Unmarshal(raw, &data)
	if err != nil {
		return nil, fmt.Errorf("file store: failed to decode yaml %s: %w", filename, err)
	}
	return data, nil
}

func (fs FileStore) Set(name string, data any) error {
	filename := fs.Path(name)

	raw, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Errorf("file store: failed to encode yaml %s: %w", filename, err)
	}
	err = os.WriteFile(filename, raw, os.ModePerm)
	if err != nil {
		return fmt.Errorf("file store: failed to write file %s: %w", filename, err)
	}
	return nil
}
