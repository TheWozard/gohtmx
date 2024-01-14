package gohtmx

import (
	"net/http"
	"net/url"

	"github.com/TheWozard/gohtmx/v2/core"
)

func LoadData(params ...string) func(r *http.Request) core.TemplateData {
	return func(r *http.Request) core.TemplateData {
		data := core.TemplateData{}
		if r.Method == http.MethodGet {
			query := r.URL.Query()
			for _, key := range params {
				data[key] = query.Get(key)
			}
		} else {
			err := r.ParseForm()
			if err != nil {
				return data
			}
			for _, key := range params {
				data[key] = r.Form.Get(key)
			}
		}
		return data
	}
}

func AddValuesToQuery(params ...string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		r.ParseForm()
		current, err := url.Parse(r.Header.Get("HX-Current-URL"))
		if err != nil {
			return
		}
		query := current.Query()
		for _, key := range params {
			query.Set(key, r.FormValue(key))
		}
		current.RawQuery = query.Encode()
		w.Header().Set("HX-Push-Url", current.String())
	}
}
