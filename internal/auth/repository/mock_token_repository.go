package repository

import (
	"github.com/shuvo-paul/uptimebot/internal/auth/model"
)

type TokenRepositoryMock struct {
	SaveTokenFunc                func(token *model.AccountToken) (*model.AccountToken, error)
	GetTokenByValueFunc          func(token string) (*model.AccountToken, error)
	MarkTokenUsedFunc            func(tokenID int) error
	GetTokensByUserIDFunc        func(userID int) ([]*model.AccountToken, error)
	InvalidateExistingTokensFunc func(userID int, tokenType model.TokenType) error
}

func (m *TokenRepositoryMock) SaveToken(token *model.AccountToken) (*model.AccountToken, error) {
	return m.SaveTokenFunc(token)
}

func (m *TokenRepositoryMock) GetTokenByValue(token string) (*model.AccountToken, error) {
	return m.GetTokenByValueFunc(token)
}

func (m *TokenRepositoryMock) MarkTokenUsed(tokenID int) error {
	return m.MarkTokenUsedFunc(tokenID)
}

func (m *TokenRepositoryMock) GetTokensByUserID(userID int) ([]*model.AccountToken, error) {
	return m.GetTokensByUserIDFunc(userID)
}

func (m *TokenRepositoryMock) InvalidateExistingTokens(userID int, tokenType model.TokenType) error {
	return m.InvalidateExistingTokensFunc(userID, tokenType)
}
