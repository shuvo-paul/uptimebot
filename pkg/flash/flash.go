package flash

import (
	"context"
	"log/slog"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type FlashStoreInterface interface {
	SetFlash(flashID, key string, value any)
	GetFlash(flashID, key string) any
}

var _ FlashStoreInterface = (*FlashStore)(nil)

type contextKey string

const flashIdKey contextKey = "flashIdKey"

type FlashStore struct {
	mu      sync.RWMutex
	flashes map[string]map[string]any
}

func defaultFlashIdGenerator() string {
	return uuid.New().String()
}

func NewFlashStore() *FlashStore {
	return &FlashStore{
		flashes: make(map[string]map[string]any),
	}
}

func (fs *FlashStore) SetFlash(flashID, key string, value any) {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	if _, exists := fs.flashes[flashID]; !exists {
		fs.flashes[flashID] = make(map[string]any)
	}
	fs.flashes[flashID][key] = value
}

func (fs *FlashStore) GetFlash(flashID, key string) any {
	fs.mu.RLock()
	session, exists := fs.flashes[flashID]
	fs.mu.RUnlock()

	if !exists {
		return nil
	}

	fs.mu.Lock()
	defer fs.mu.Unlock()

	value, ok := session[key]
	if ok {
		delete(session, key)
		// Clean up empty sessions
		if len(session) == 0 {
			delete(fs.flashes, flashID)
		}
	}
	return value
}

func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("flash_id")

		if err == nil && cookie.Value != "" {
			ctx := context.WithValue(r.Context(), flashIdKey, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		flashID := defaultFlashIdGenerator()
		http.SetCookie(w, &http.Cookie{
			Name:  "flash_id",
			Value: flashID,
			Path:  "/",
		})

		ctx := context.WithValue(r.Context(), flashIdKey, flashID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetFlashIDFromContext(ctx context.Context) string {
	flashId, ok := ctx.Value(flashIdKey).(string)
	if !ok {
		slog.Error("Flash Store", "OK", ok)
	}
	return flashId
}
