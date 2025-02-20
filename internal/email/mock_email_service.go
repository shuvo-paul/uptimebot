package email

type EmailServiceMock struct {
	SetToFunc      func(to string) error
	SetSubjectFunc func(subject string) error
	SetBodyFunc    func(body string) error
	SendEmailFunc  func() error
}

func (m *EmailServiceMock) SetTo(to string) error {
	return m.SetToFunc(to)
}

func (m *EmailServiceMock) SetSubject(subject string) error {
	return m.SetSubjectFunc(subject)
}

func (m *EmailServiceMock) SetBody(body string) error {
	return m.SetBodyFunc(body)
}

func (m *EmailServiceMock) SendEmail() error {
	return m.SendEmailFunc()
}
