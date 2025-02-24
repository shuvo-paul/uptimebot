package flash

import "context"

// MockFlashStore is a mock implementation of flash.FlashStoreInterface
type MockFlashStore struct {
	flashes map[string]map[string][]string
}

// NewMockFlashStore creates a new instance of MockFlashStore
func NewMockFlashStore() *MockFlashStore {
	return &MockFlashStore{
		flashes: make(map[string]map[string][]string),
	}
}

func (m *MockFlashStore) setFlash(flashID, key string, value []string) {
	if _, exists := m.flashes[flashID]; !exists {
		m.flashes[flashID] = make(map[string][]string)
	}
	m.flashes[flashID][key] = value
}

func (m *MockFlashStore) getFlash(flashID, key string) []string {
	session, exists := m.flashes[flashID]
	if !exists {
		return nil
	}

	value, ok := session[key]
	if !ok {
		return nil
	}
	return value
}

func (m *MockFlashStore) SetErrors(ctx context.Context, errors []string) {
	flashId := GetFlashIDFromContext(ctx)
	if flashId == "" {
		return
	}
	m.setFlash(flashId, FlashKeyErrors, errors)
}

func (m *MockFlashStore) SetSuccesses(ctx context.Context, successes []string) {
	flashId := GetFlashIDFromContext(ctx)
	if flashId == "" {
		return
	}
	m.setFlash(flashId, FlashKeySuccesses, successes)
}

func (m *MockFlashStore) GetSuccesses(ctx context.Context) []string {
	flashId := GetFlashIDFromContext(ctx)
	if flashId == "" {
		return nil
	}
	return m.getFlash(flashId, FlashKeySuccesses)
}

func (m *MockFlashStore) GetErrors(ctx context.Context) []string {
	flashId := GetFlashIDFromContext(ctx)
	if flashId == "" {
		return nil
	}
	return m.getFlash(flashId, FlashKeyErrors)
}
