package httpjson

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHTTPJSON(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		JSONError(w, http.StatusInternalServerError, "test error %s", "foo")
	}
	req := httptest.NewRequest("GET", "http://localhost", nil)
	w := httptest.NewRecorder()
	handler(w, req)

	resp := w.Result()
	body, _ := ioutil.ReadAll(resp.Body)

	assert.Equal(t, 500, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))
	assert.Equal(t, "{\"error\":\"test error foo\"}\n", string(body))
}
