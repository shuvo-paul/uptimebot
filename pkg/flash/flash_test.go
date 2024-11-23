package flash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndGetFlash(t *testing.T) {
	fs := NewFlashStore()
	flashID := "test-id"

	fs.SetFlash(flashID, "key1", "value1")

	value := fs.GetFlash(flashID, "key1")

	assert.Equal(t, value, "value1")

	value = fs.GetFlash(flashID, "key1")

	assert.Nil(t, value)
}

func TestMiddleware(t *testing.T) {

	handler := Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flashID := GetFlashIDFromContext(r.Context())
		assert.NotEmpty(t, flashID)
	}))

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	cookie := rr.Result().Cookies()[0]

	assert.Equal(t, cookie.Name, "flash_id")
	assert.NotEmpty(t, cookie.Value)
}
