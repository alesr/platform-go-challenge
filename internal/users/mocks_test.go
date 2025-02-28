package users

import "context"

// Repository mock

var _ repository = &repoMock{}

type repoMock struct {
	listUsersFunc func(ctx context.Context) []User
	fetchUserFunc func(_ context.Context, id string) (*User, error)
}

func (m *repoMock) ListUsers(ctx context.Context) []User {
	return m.listUsersFunc(ctx)
}

func (m *repoMock) FetchUser(ctx context.Context, id string) (*User, error) {
	return m.fetchUserFunc(ctx, id)
}
