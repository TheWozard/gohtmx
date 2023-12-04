package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/TheWozard/gohtmx"
	"github.com/gorilla/mux"
)

func Body(store Store) gohtmx.Component {
	return gohtmx.Fragment{
		gohtmx.Div{ID: "error"},
		gohtmx.Form{
			ID:               "search",
			LoadTemplateData: gohtmx.LoadTemplateDataFromQuery("document"),
			LoadSubmitData:   gohtmx.LoadTemplateDataFromForm("document"),
			Content: AsCard(gohtmx.Fragment{
				AsFieldAddonInput(0,
					gohtmx.InputText{Classes: []string{"input"}, Placeholder: "Document", Name: "document"},
					gohtmx.InputSubmit{Classes: []string{"button", "is-info"}, Text: "Submit"},
				),
			}),
			Error:   AsCard(gohtmx.Raw("{{.error}}")),
			Success: Form(store),
		},
	}
}

func Form(store Store) gohtmx.Component {
	return gohtmx.Form{
		ID: "document",
		LoadTemplateData: func(r *http.Request) (gohtmx.TemplateData, error) {
			data := gohtmx.DataFromContext(r.Context())
			document := ""
			if search, ok := data["search"].(gohtmx.TemplateData); ok {
				document, _ = search["document"].(string)
			}
			data, err := store.Get(document)
			return data, err
		},
		LoadSubmitData: func(r *http.Request) (gohtmx.TemplateData, error) {
			data := gohtmx.TemplateData{}
			for key, value := range r.Form {
				data[key] = value[0]
			}
			return nil, store.Set(r.Form["document"][0], data)
		},
		Content: AsCard(gohtmx.Fragment{
			gohtmx.InputHidden{Name: "document", Value: `{{or .search.document ""}}`},
			LabeledField("First", gohtmx.InputText{Classes: []string{"input"}, Placeholder: "First", Name: "first"}),
			LabeledField("Last", gohtmx.InputText{Classes: []string{"input"}, Placeholder: "Last", Name: "last"}),
			LabeledField("Title", gohtmx.InputText{Classes: []string{"input"}, Placeholder: "Title", Name: "title"}),
			gohtmx.InputSubmit{Text: "Submit"},
		}),
		Error:   AsCard(gohtmx.Raw("{{.error}}")),
		Success: AsCard(gohtmx.Raw("Success!")),
	}
}

func AsCard(content gohtmx.Component) gohtmx.Component {
	return gohtmx.Div{
		Classes: []string{"card centered"},
		Content: gohtmx.Div{
			Classes: []string{"card-content"},
			Content: content,
		},
	}
}

func AsFieldAddonInput(fill int, contents ...gohtmx.Component) gohtmx.Component {
	for i, content := range contents {
		div := gohtmx.Div{
			Classes: []string{"control"},
			Content: content,
		}
		if i == fill {
			div.Classes = append(div.Classes, "fill")
		}
		contents[i] = div
	}
	return gohtmx.Div{
		Classes: []string{"field", "has-addons"},
		Content: gohtmx.Fragment(contents),
	}
}

func LabeledField(label string, content gohtmx.Component) gohtmx.Component {
	return gohtmx.Div{
		Classes: []string{"field"},
		Content: gohtmx.Fragment{
			gohtmx.Label{Classes: []string{"label"}, Text: label},
			gohtmx.Div{
				Classes: []string{"control", "fill"},
				Content: content,
			},
		},
	}
}

func main() {
	store := Store{Path: "./example/forms/store"}

	mux := mux.NewRouter()
	mux.PathPrefix("/assets/").Handler(http.FileServer(http.Dir("./example/forms")))
	gohtmx.Document{
		Header: gohtmx.Fragment{
			gohtmx.Raw(`<meta charset="utf-8">`),
			gohtmx.Raw(`<meta name="viewport" content="width=device-width, initial-scale=1">`),
			gohtmx.Raw(`<title>Form Inputs</title>`),
			gohtmx.Raw(`<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bulma@0.9.4/css/bulma.min.css">`),
			gohtmx.Raw(`<link rel="stylesheet" href="/assets/style.css">`),
			gohtmx.Raw(`<script src="https://unpkg.com/htmx.org@1.9.6/dist/htmx.min.js"></script>`),
			gohtmx.Raw(`<script defer src="/assets/script.js"></script>`),
		},
		Body: Body(store),
	}.Mount("/", mux)

	log.Default().Println("staring server at http://localhost:8080")
	log.Fatal((&http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}).ListenAndServe())
}

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
