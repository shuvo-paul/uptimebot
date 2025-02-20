package flash

// MockFlashStore is a mock implementation of flash.FlashStoreInterface
type MockFlashStore struct {
}

func (m *MockFlashStore) SetFlash(flashID, key string, value any) {

}

func (m *MockFlashStore) GetFlash(flashID, key string) any {
	return nil
}
