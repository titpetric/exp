package generic_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/titpetric/exp/pkg/generic"
)

func TestTemplateExample(t *testing.T) {
	example := generic.NewTemplateExample(&templatesFS)

	req, err := http.NewRequestWithContext(t.Context(), "GET", "/", nil)
	assert.NoError(t, err)

	recorder := httptest.NewRecorder()
	example.ServeHTTP(recorder, req)

	resp := recorder.Result()
	assert.NotNil(t, resp)

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	bodystr := string(body)
	assert.Equal(t, 200, resp.StatusCode, "expected 200 OK response")
	assert.Contains(t, bodystr, "Alice Example", "rendered output should contain user name")
	assert.Contains(t, bodystr, "https://example.com/alice", "rendered output should contain user link")
	assert.Contains(t, bodystr, "<img ", "rendered output should contain img tag")
}
