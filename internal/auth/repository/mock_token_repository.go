package repository

import (
	"github.com/shuvo-paul/uptimebot/internal/auth/model"
)

type AccountTokenRepositoryMock struct {
	SaveTokenFunc                func(token *model.AccountToken) (*model.AccountToken, error)
	GetTokenByValueFunc          func(token string) (*model.AccountToken, error)
	MarkTokenUsedFunc            func(tokenID int) error
	GetTokensByUserIDFunc        func(userID int) ([]*model.AccountToken, error)
	InvalidateExistingTokensFunc func(userID int, tokenType model.TokenType) error
}

func (m *AccountTokenRepositoryMock) SaveToken(token *model.AccountToken) (*model.AccountToken, error) {
	return m.SaveTokenFunc(token)
}

func (m *AccountTokenRepositoryMock) GetTokenByValue(token string) (*model.AccountToken, error) {
	return m.GetTokenByValueFunc(token)
}

func (m *AccountTokenRepositoryMock) MarkTokenUsed(tokenID int) error {
	return m.MarkTokenUsedFunc(tokenID)
}

func (m *AccountTokenRepositoryMock) GetTokensByUserID(userID int) ([]*model.AccountToken, error) {
	return m.GetTokensByUserIDFunc(userID)
}

func (m *AccountTokenRepositoryMock) InvalidateExistingTokens(userID int, tokenType model.TokenType) error {
	return m.InvalidateExistingTokensFunc(userID, tokenType)
}
