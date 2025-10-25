package mocks

import "context"

type MockUserProvider struct {
	UserID string
	Ok     bool
}

func (m *MockUserProvider) GetUserID(ctx context.Context) (string, bool) {
	return m.UserID, m.Ok
}

func NewMockUserProvider(userID string, ok bool) *MockUserProvider {
	return &MockUserProvider{
		UserID: userID,
		Ok:     ok,
	}
}
