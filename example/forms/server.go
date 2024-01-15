package main

import (
	"embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/TheWozard/gohtmx/v2"
	"github.com/TheWozard/gohtmx/v2/core"
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

//go:embed assets/*
var assets embed.FS

func main() {
	store := Store{Path: "./example/forms/store"}

	f := gohtmx.NewDefaultFramework()
	err := f.AddTemplateInteraction(
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
	if err != nil {
		log.Fatal(err)
	}

	mux := mux.NewRouter()
	mux.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	mux.PathPrefix("/").Handler(f)

	log.Default().Println("staring server at http://localhost:8080")
	log.Fatal((&http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}).ListenAndServe())
}

func Body(store Store) gohtmx.Component {
	return gohtmx.Div{
		Attr: []gohtmx.Attr{{Name: "hx-ext", Value: "morph"}},
		Content: gohtmx.Fragment{
			gohtmx.Div{ID: "error"},
			gohtmx.Path{
				ID:          "body",
				DefaultPath: "form",
				Paths: map[string]gohtmx.Component{
					"form": Search(store),
					"foo":  gohtmx.Raw("foo"),
					"bar":  gohtmx.Raw("bar"),
				},
			},
		},
	}
}

func Search(store Store) gohtmx.Component {
	return gohtmx.Form{
		ID: "search",
		Content: AsCard(gohtmx.TWith{
			Func: gohtmx.LoadData("search"),
			Content: gohtmx.InputSearch{
				Placeholder: "Search", Name: "search", Value: "{{.search}}", Classes: []string{"input-group"},
				Options: func(r *http.Request) []any {
					data := gohtmx.LoadData("search")(r)
					search, ok := data["search"].(string)
					if !ok {
						return nil
					}
					options := []any{}
					for _, name := range store.List() {
						if len(options) >= 5 {
							break
						}
						if strings.HasPrefix(strings.ToLower(name), strings.ToLower(search)) && name != search {
							options = append(options, name)
						}
					}
					return options
				},
				PrePopulate: true,
				Additional:  gohtmx.InputSubmit{Text: "Search"},
				Target:      "#search-results",
			},
		}),
		Action: func(w http.ResponseWriter, r *http.Request) (core.TemplateData, error) {
			gohtmx.AddValuesToQuery("search")(w, r)
			return nil, nil
		},
		CanAutoComplete: func(r *http.Request) bool {
			data := gohtmx.LoadData("search")(r)
			search, ok := data["search"].(string)
			return ok && search != ""
		},
		Error:   AsCard(gohtmx.Raw("{{.error}}")),
		Success: Form(store),
	}
}

func Form(store Store) gohtmx.Component {
	return gohtmx.Form{
		ID: "document",
		Action: func(w http.ResponseWriter, r *http.Request) (core.TemplateData, error) {
			data := gohtmx.LoadData("document", "first", "last", "title")(r)
			document, ok := data["document"].(string)
			if !ok {
				return nil, fmt.Errorf("no document found")
			}
			return core.TemplateData{
				"time": time.Now().Format(time.RFC3339Nano),
			}, store.Set(document, data)
		},
		Content: AsCard(gohtmx.TWith{
			Func: func(r *http.Request) core.TemplateData {
				data := gohtmx.LoadData("search")(r)
				var document any
				search, ok := data["search"].(string)
				if ok {
					search = strings.ToLower(search)
					data["search"] = search
					document, _ = store.Get(search)
				}
				return data.Merge(core.TemplateData{"document": document})
			},
			Content: gohtmx.Fragment{
				gohtmx.InputHidden{Name: "document", Value: "{{.search}}"},
				gohtmx.InputText{
					Label: "First", Name: "first", Value: "{{.document.first}}",
					Validate: func(r *http.Request) core.TemplateData {
						data := gohtmx.LoadData("first")(r)
						first, _ := data["first"].(string)
						return core.TemplateData{
							"modified": true,
							"document": core.TemplateData{
								"first": cases.Title(language.English).String(first),
							},
						}
					},
					Classes: []string{"input-group", "{{if .modified}}modified{{end}}"},
				},
				gohtmx.InputText{Label: "Last", Name: "last", Value: "{{.document.last}}", Classes: []string{"input-group"}},
				gohtmx.InputText{Label: "Title", Name: "title", Value: "{{.document.title}}", Classes: []string{"input-group"}},
				gohtmx.InputSubmit{Text: "Submit"},
			},
		}),
		Error:   AsCard(gohtmx.Raw("{{.error}}")),
		Success: AsCard(gohtmx.Raw("Success at {{.time}}!")),
	}
}

func AsCard(content gohtmx.Component) gohtmx.Component {
	return gohtmx.Div{
		Classes: []string{"card centered vertical-space"},
		Content: content,
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
