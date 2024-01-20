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
	"github.com/gorilla/mux"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// go:embed assets/*
// var assets embed.FS

func main() {
	store := Store{Path: "./example/forms/store"}

	f := gohtmx.NewDefaultFramework()
	err := f.AddInteraction(
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
	h, err := f.Build()
	if err != nil {
		log.Fatal(err)
	}

	mux := mux.NewRouter()
	// mux.PathPrefix("/assets/").Handler(http.FileServer(http.FS(assets)))
	mux.PathPrefix("/assets/").Handler(http.StripPrefix("/assets/", http.FileServer(http.Dir("./example/forms/assets"))))
	mux.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate Delay
		time.Sleep(200 * time.Millisecond)
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
		Attrs: gohtmx.Attrs().Value("hx-ext", "morph"),
		Content: gohtmx.Fragment{
			gohtmx.Div{ID: "error"},
			Header(),
			gohtmx.Path{
				ID:          "body",
				DefaultPath: "search",
				Paths: map[string]gohtmx.Component{
					"search":  Search(store),
					"confirm": gohtmx.Raw("foo"),
					"bar":     gohtmx.Raw("bar"),
				},
			},
		},
	}
}

func Header() gohtmx.Component {
	return gohtmx.Tag{Name: "header",
		Content: gohtmx.Div{
			UpdateWith: []string{"/search", "/confirm", "/bar"},
			Content: gohtmx.Fragment{
				TabSelector("Search", "/search", "#body"),
				TabSelector("Confirm", "/confirm", "#body"),
				TabSelector("Bar", "/bar", "#body"),
			},
		},
	}
}

func Search(store Store) gohtmx.Component {
	return gohtmx.Fragment{
		gohtmx.Form{
			ID: "search",
			Content: AsCard(gohtmx.TWith{
				Func: gohtmx.GetDataFromRequest("search"),
				Content: gohtmx.Fragment{
					gohtmx.InputSearch{
						Placeholder: "Search", Name: "search", Value: "{{.search}}", Classes: []string{"input-group"},
						Options: func(r *http.Request) []any {
							search, ok := gohtmx.GetDataFromRequest("search")(r)["search"].(string)
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
						AutoFocus:   true,
					},
				},
			}),
			UpdateParams: []string{"search"},
			UpdateForm:   true,
			Target:       "#search-results",
			Error:        AsCard(gohtmx.Raw("{{.error}}")),
			Success:      Form(store),
		},
		gohtmx.Div{
			ID: "search-results",
			// An alternate option to AutoComplete is to write it on initial load.
			Content: gohtmx.TCondition{
				Condition: func(r *http.Request) bool {
					data := gohtmx.GetDataFromRequest("search")(r)
					search, ok := data["search"].(string)
					return ok && search != ""
				},
				// Interaction is disabled to prevent duplicate interactions.
				Content: gohtmx.MetaDisableInteraction{
					Content: Form(store),
				},
			},
		},
	}
}

func Form(store Store) gohtmx.Component {
	return gohtmx.Fragment{
		gohtmx.Form{
			ID: "document",
			Submit: func(data gohtmx.Data) (gohtmx.Data, error) {
				document, ok := data["document"].(string)
				if !ok {
					return nil, fmt.Errorf("no document found")
				}
				return gohtmx.Data{
					"time": time.Now().Format(time.RFC3339Nano),
				}, store.Set(document, data)
			},
			Content: AsCard(gohtmx.TWith{
				Func: func(r *http.Request) gohtmx.Data {
					data := gohtmx.GetDataFromRequest("search")(r)
					var document any
					search, ok := data["search"].(string)
					if ok {
						search = strings.ToLower(search)
						data["search"] = search
						document, _ = store.Get(search)
					}
					return data.Merge(gohtmx.Data{"document": document})
				},
				Content: gohtmx.Fragment{
					gohtmx.InputHidden{Name: "document", Value: "{{.search}}"},
					gohtmx.InputText{
						Label: "First", Name: "first", Value: "{{.document.first}}",
						OnChange: func(r *http.Request) gohtmx.Data {
							data := gohtmx.GetDataFromRequest("first")(r)
							first, _ := data["first"].(string)
							return gohtmx.Data{
								"modified": true,
								"document": gohtmx.Data{
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
			Target:  "#document-result",
			Error:   AsCard(gohtmx.Raw("{{.error}}")),
			Success: AsCard(gohtmx.Raw("Success at {{.time}}!")),
		},
		gohtmx.Div{ID: "document-result"},
	}
}

func TabSelector(text, path, target string) gohtmx.Component {
	return gohtmx.Button{
		Attr: gohtmx.Attrs().Value("hx-get", path).Value("hx-target", target).
			Condition(gohtmx.IsRequestAtPath(path), gohtmx.Attrs().Value("class", "active")),
		Content: gohtmx.Raw(text),
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
