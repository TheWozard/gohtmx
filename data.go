package gohtmx

import (
	"net/http"
	"net/url"
)

// GetDataFromRequest returns Data with the values from the request. Handles both GET and POST requests.
func GetDataFromRequest(params ...string) func(r *http.Request) Data {
	return func(r *http.Request) Data {
		data := Data{}
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

// GetAllDataFromRequest returns all Data from the request. Handles both GET and POST requests.
func GetAllDataFromRequest(r *http.Request) Data {
	data := Data{}
	if r.Method == http.MethodGet {
		query := r.URL.Query()
		for key := range query {
			data[key] = query.Get(key)
		}
	} else {
		err := r.ParseForm()
		if err != nil {
			return data
		}
		for key := range r.Form {
			data[key] = r.Form.Get(key)
		}
	}
	return data
}

func UpdateParams(names ...string) Middleware {
	loader := GetDataFromRequest(names...)
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			loader(r).SetInResponse(w, r)
			h.ServeHTTP(w, r)
		})
	}
}

// Data represents a neutral representation of data that is passed through any method HTMX requests.
type Data map[string]any

// Merge merges two Data maps together. The addition map will overwrite any existing keys.
func (d Data) Merge(a Data) Data {
	if len(d) == 0 {
		return a
	}
	for k, v := range a {
		d[k] = v
	}
	return d
}

// SetValuesInResponse sets the data Data in the response. Handles both GET and POST requests.
func (d Data) SetInResponse(w http.ResponseWriter, r *http.Request) {
	current, err := url.Parse(r.Header.Get("HX-Current-URL"))
	if err != nil {
		return
	}
	query := current.Query()
	for key, value := range d {
		query.Set(key, value.(string))
	}
	current.RawQuery = query.Encode()
	w.Header().Set("HX-Push-Url", current.String())
}

// Subset creates a new Data map with only the passed keys
func (d Data) Subset(keys ...string) Data {
	result := Data{}
	for _, key := range keys {
		if value, ok := d[key]; ok {
			result[key] = value
		}
	}
	return result
}
