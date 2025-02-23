package flash

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSetAndGetFlash(t *testing.T) {
	fs := NewFlashStore()
	flashID := "test-id"

	testValue := []string{"value1"}
	fs.setFlash(flashID, "key1", testValue)

	value := fs.getFlash(flashID, "key1")

	assert.Equal(t, value, testValue)
	assert.Equal(t, len(value), len(testValue))

	value = fs.getFlash(flashID, "key1")

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

func TestSetErrors(t *testing.T) {
	fs := NewFlashStore()
	ctx := context.Background()

	// Test with empty flash ID
	fs.SetErrors(ctx, []string{"error1", "error2"})
	assert.Nil(t, fs.getFlash("", FlashKeyErrors))

	// Test with valid flash ID
	flashID := "test-flash-id"
	ctx = context.WithValue(ctx, flashIdKey, flashID)
	errors := []string{"error1", "error2"}
	fs.SetErrors(ctx, errors)

	// Verify errors were stored
	value := fs.getFlash(flashID, FlashKeyErrors)
	assert.NotNil(t, value)
	assert.Equal(t, errors, value)
	assert.Equal(t, len(errors), len(value))

	// Verify errors were cleared after retrieval
	value = fs.getFlash(flashID, FlashKeyErrors)
	assert.Nil(t, value)
}

func TestSetSuccesses(t *testing.T) {
	fs := NewFlashStore()
	ctx := context.Background()

	// Test with empty flash ID
	fs.SetSuccesses(ctx, []string{"success1", "success2"})
	assert.Nil(t, fs.getFlash("", FlashKeySuccesses))

	// Test with valid flash ID
	flashID := "test-flash-id"
	ctx = context.WithValue(ctx, flashIdKey, flashID)
	successes := []string{"success1", "success2"}
	fs.SetSuccesses(ctx, successes)

	// Verify successes were stored
	value := fs.getFlash(flashID, FlashKeySuccesses)
	assert.NotNil(t, value)
	assert.Equal(t, successes, value)

	// Verify successes were cleared after retrieval
	value = fs.getFlash(flashID, FlashKeySuccesses)
	assert.Nil(t, value)
}

func TestGetSuccesses(t *testing.T) {
	fs := NewFlashStore()
	ctx := context.Background()

	// Test with empty flash ID
	value := fs.GetSuccesses(ctx)
	assert.Nil(t, value)

	// Test with valid flash ID
	flashID := "test-flash-id"
	ctx = context.WithValue(ctx, flashIdKey, flashID)
	successes := []string{"success1", "success2"}
	fs.SetSuccesses(ctx, successes)

	// Verify successes can be retrieved
	value = fs.GetSuccesses(ctx)
	assert.NotNil(t, value)
	assert.Equal(t, successes, value)

	// Verify successes were cleared after retrieval
	value = fs.GetSuccesses(ctx)
	assert.Nil(t, value)
}

func TestGetErrors(t *testing.T) {
	fs := NewFlashStore()
	ctx := context.Background()

	// Test with empty flash ID
	value := fs.GetErrors(ctx)
	assert.Nil(t, value)

	// Test with valid flash ID
	flashID := "test-flash-id"
	ctx = context.WithValue(ctx, flashIdKey, flashID)
	errors := []string{"error1", "error2"}
	fs.SetErrors(ctx, errors)

	// Verify errors can be retrieved
	value = fs.GetErrors(ctx)
	assert.NotNil(t, value)
	assert.Equal(t, errors, value)

	// Verify errors were cleared after retrieval
	value = fs.GetErrors(ctx)
	assert.Nil(t, value)
}
