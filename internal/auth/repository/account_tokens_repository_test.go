package repository

import (
	"testing"
	"time"

	"github.com/shuvo-paul/uptimebot/internal/auth/model"
	"github.com/shuvo-paul/uptimebot/internal/testutil"
	"github.com/stretchr/testify/assert"
)

func TestVerificationTokenRepository(t *testing.T) {
	db := testutil.NewInMemoryDB()
	defer db.Close()

	repo := NewVerificationTokenRepository(db)

	t.Run("SaveToken", func(t *testing.T) {
		token := &model.AccountToken{
			UserID:    1,
			Token:     "test-token",
			Type:      model.TokenTypeEmailVerification,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Used:      false,
		}

		savedToken, err := repo.SaveToken(token)
		assert.NoError(t, err)
		assert.NotZero(t, savedToken.ID)
		assert.Equal(t, token.UserID, savedToken.UserID)
		assert.Equal(t, token.Token, savedToken.Token)
	})

	t.Run("GetTokenByValue", func(t *testing.T) {
		token, err := repo.GetTokenByValue("test-token")
		assert.NoError(t, err)
		assert.NotNil(t, token)
		assert.Equal(t, "test-token", token.Token)
		assert.Equal(t, 1, token.UserID)
	})

	t.Run("GetTokenByValue_NotFound", func(t *testing.T) {
		token, err := repo.GetTokenByValue("non-existent-token")
		assert.NoError(t, err)
		assert.Nil(t, token)
	})

	t.Run("MarkTokenUsed", func(t *testing.T) {
		err := repo.MarkTokenUsed(1)
		assert.NoError(t, err)

		token, err := repo.GetTokenByValue("test-token")
		assert.NoError(t, err)
		assert.True(t, token.Used)
	})

	t.Run("GetTokensByUserID", func(t *testing.T) {
		// Add another token for the same user
		newToken := &model.AccountToken{
			UserID:    1,
			Token:     "test-token-2",
			Type:      model.TokenTypePasswordReset,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Used:      false,
		}
		_, err := repo.SaveToken(newToken)
		assert.NoError(t, err)

		tokens, err := repo.GetTokensByUserID(1)
		assert.NoError(t, err)
		assert.Len(t, tokens, 2)
	})

	t.Run("InvalidateExistingTokens", func(t *testing.T) {
		token := &model.AccountToken{
			UserID:    2,
			Token:     "test-token-3",
			Type:      model.TokenTypeEmailVerification,
			ExpiresAt: time.Now().Add(24 * time.Hour),
			Used:      false,
		}
		_, err := repo.SaveToken(token)
		assert.NoError(t, err)

		err = repo.InvalidateExistingTokens(2, model.TokenTypeEmailVerification)
		assert.NoError(t, err)

		retrievedToken, err := repo.GetTokenByValue(token.Token)
		assert.NoError(t, err)
		assert.NotNil(t, retrievedToken)
		assert.True(t, retrievedToken.Used)
	})
}
