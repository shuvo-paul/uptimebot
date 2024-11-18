package flash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndGetFlash(t *testing.T) {
	fs := NewFlashStore()
	fs.generateFlashID = func() string { return "test-id" }
	flashID := "test-id"

	fs.SetFlash(flashID, "key1", "value1")

	value := fs.GetFlash(flashID, "key1")

	assert.Equal(t, value, "value1")

	value = fs.GetFlash(flashID, "key1")

	assert.Nil(t, value)
}

func TestMiddleware(t *testing.T) {
	fs := NewFlashStore()
	fs.generateFlashID = func() string { return "test-id" }

	handler := fs.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		flashID, ok := GetFlashIDFromContext(r.Context())
		assert.True(t, ok)
		assert.Equal(t, flashID, "test-id")
	}))

	req, _ := http.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	cookie := rr.Result().Cookies()[0]

	assert.Equal(t, cookie.Name, "flash_id")
	assert.Equal(t, cookie.Value, "test-id")
}
