package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/TheWozard/gojsox"
	"github.com/nqd/flat"
	"gopkg.in/yaml.v3"
)

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

func main() {
	tmp, err := template.ParseGlob("./template/*.html")
	if err != nil {
		log.Fatal(err)
	}

	store := FileStore{BasePath: "./data/"}

	name, tmp, err := gojsox.ParseFile(tmp, "./schemas/example.yaml")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		value, _ := store.Get("default")
		err := tmp.ExecuteTemplate(w, "body", value)
		if err != nil {
			fmt.Printf("failed to serve: %s\n", err)
		}
	})
	mux.HandleFunc(fmt.Sprintf("/%s", name), func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		result := make(map[string]any, len(r.PostForm))
		for key, value := range r.PostForm {
			result[key] = value[0]
		}
		result, _ = flat.Unflatten(result, nil)
		store.Set("default", result)

		w.Write([]byte("<div>Success - "))
		w.Write([]byte(time.Now().Format(time.RFC3339)))
		w.Write([]byte("</div>"))
	})

	fmt.Printf("staring server at http://localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
}
