package mock

import (
	"sync"

	"github.com/shuvo-paul/uptimebot/internal/email"
)

type MailServiceMock struct {
	mutex sync.Mutex

	SetToFunc      func(to string) error
	SetSubjectFunc func(subject string) error
	SetBodyFunc    func(body string) error
	SendEmailFunc  func() error

	calls struct {
		SetTo      []string
		SetSubject []string
		SetBody    []string
		SendEmail  int
	}
}

// Verify MailServiceMock implements email.Mailer interface
var _ email.Mailer = (*MailServiceMock)(nil)

func (m *MailServiceMock) SetTo(to string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.calls.SetTo = append(m.calls.SetTo, to)
	if m.SetToFunc != nil {
		return m.SetToFunc(to)
	}
	return nil
}

func (m *MailServiceMock) SetSubject(subject string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.calls.SetSubject = append(m.calls.SetSubject, subject)
	if m.SetSubjectFunc != nil {
		return m.SetSubjectFunc(subject)
	}
	return nil
}

func (m *MailServiceMock) SetBody(body string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.calls.SetBody = append(m.calls.SetBody, body)
	if m.SetBodyFunc != nil {
		return m.SetBodyFunc(body)
	}
	return nil
}

func (m *MailServiceMock) SendEmail() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.calls.SendEmail++
	if m.SendEmailFunc != nil {
		return m.SendEmailFunc()
	}
	return nil
}

// Helper methods for tests

func (m *MailServiceMock) GetSetToCalls() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.calls.SetTo...)
}

func (m *MailServiceMock) GetSetSubjectCalls() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.calls.SetSubject...)
}

func (m *MailServiceMock) GetSetBodyCalls() []string {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return append([]string{}, m.calls.SetBody...)
}

func (m *MailServiceMock) GetSendEmailCallCount() int {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	return m.calls.SendEmail
}
