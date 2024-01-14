package gohtmx_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheWozard/gohtmx/v2"
	"github.com/TheWozard/gohtmx/v2/core"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTemplateHandler_ServeHTTPExtraData(t *testing.T) {
	testCases := []struct {
		desc     string
		content  gohtmx.Component
		request  *http.Request
		extra    core.TemplateData
		expected string
	}{
		{
			desc:     "Empty Template",
			content:  nil,
			request:  httptest.NewRequest("GET", "/", nil),
			expected: "",
		},
		{
			desc:     "Template with Content",
			content:  gohtmx.Raw("content"),
			request:  httptest.NewRequest("GET", "/", nil),
			expected: "content",
		},
		{
			desc:     "Template with Request Data",
			content:  gohtmx.Raw("{{.request.Method}} {{.request.URL.Path}}"),
			request:  httptest.NewRequest("GET", "/test", nil),
			expected: "GET /test",
		},
		{
			desc:    "Template with Context Data",
			content: gohtmx.Raw("{{.data}}"),
			request: httptest.NewRequest("GET", "/test", nil),
			extra: core.TemplateData{
				"data": "context data",
			},
			expected: "context data",
		},
	}

	for _, tC := range testCases {
		t.Run(tC.desc, func(t *testing.T) {
			handler, err := gohtmx.NewTemplateHandler(gohtmx.NewDefaultFramework(), tC.content)
			require.NoError(t, err)
			recorder := httptest.NewRecorder()
			handler.ServeHTTPWithExtraData(recorder, tC.request, tC.extra)
			assert.Equal(t, tC.expected, recorder.Body.String())
		})
	}
}
