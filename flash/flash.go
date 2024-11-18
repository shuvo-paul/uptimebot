package flash

import (
	"context"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

type contextKey string

const flashIdKey contextKey = "flashIdKey"

type FlashStore struct {
	mu              sync.RWMutex
	flashes         map[string]map[string]any
	generateFlashID func() string
}

func defaultFlashIdGenerator() string {
	return uuid.New().String()
}

func NewFlashStore() *FlashStore {
	return &FlashStore{
		flashes:         make(map[string]map[string]any),
		generateFlashID: defaultFlashIdGenerator,
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

func (fs *FlashStore) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("flash_id")

		if err == nil && cookie.Value != "" {
			ctx := context.WithValue(r.Context(), flashIdKey, cookie.Value)
			next.ServeHTTP(w, r.WithContext(ctx))
			return
		}

		flashID := fs.generateFlashID()
		http.SetCookie(w, &http.Cookie{
			Name:  "flash_id",
			Value: flashID,
			Path:  "/",
		})

		ctx := context.WithValue(r.Context(), flashIdKey, flashID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetFlashIDFromContext(ctx context.Context) (string, bool) {
	flashId, ok := ctx.Value(flashIdKey).(string)
	return flashId, ok
}
