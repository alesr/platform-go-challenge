package users

import (
	"context"
	"testing"
	"time"

	"github.com/alesr/platform-go-challenge/internal/pkg/logutil"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService(t *testing.T) {
	t.Parallel()

	logger := logutil.NewNoop()
	repo := repoMock{}

	got := NewService(logger, &repo)

	require.NotNil(t, got)

	assert.Equal(t, logger.WithGroup("users-service"), got.logger)
	assert.Equal(t, &repo, got.repository)
}

func TestService_ListUsers(t *testing.T) {
	t.Parallel()

	now := time.Time{}.Add(time.Hour)

	givenUsers := []User{
		{
			ID:        ulid.Make(),
			Name:      "Rigoletto",
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:        ulid.Make(),
			Name:      "Mary Jane",
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	testCases := []struct {
		name            string
		givenRepoResult []User
		expected        []User
	}{
		{
			name:            "return a non-empty list",
			givenRepoResult: givenUsers,
			expected:        givenUsers,
		},
		{
			name:            "return an empty list",
			givenRepoResult: []User{},
			expected:        []User{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			repo := repoMock{
				listUsersFunc: func(ctx context.Context) []User {
					return tc.givenRepoResult
				},
			}

			svc := Service{repository: &repo}

			got := svc.ListUsers(context.TODO())

			assert.ElementsMatch(t, tc.expected, got)
		})
	}
}

func TestService_FetchUser(t *testing.T) {
	t.Parallel()

	now := time.Time{}.Add(time.Hour)

	givenUser := User{
		ID:        ulid.Make(),
		Name:      "Rigoletto",
		CreatedAt: now,
		UpdatedAt: now,
	}

	testCases := []struct {
		name            string
		givenID         string
		givenRepoResult func() (*User, error)
		expected        *User
		expectedError   error
	}{
		{
			name:    "succesfully return a user",
			givenID: givenUser.ID.String(),
			givenRepoResult: func() (*User, error) {
				return &givenUser, nil
			},
			expected: &givenUser,
		},
		{
			name:    "repo returns some error",
			givenID: "whatever-id",
			givenRepoResult: func() (*User, error) {
				return nil, assert.AnError
			},
			expectedError: assert.AnError,
		},
		{
			name:    "repo returns user not found",
			givenID: "whatever-id",
			givenRepoResult: func() (*User, error) {
				return nil, ErrUserNotFound
			},
			expectedError: ErrUserNotFound,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var repoCalled bool
			repo := repoMock{
				fetchUserFunc: func(_ context.Context, id string) (*User, error) {
					repoCalled = true
					assert.Equal(t, tc.givenID, id)
					return tc.givenRepoResult()
				},
			}

			svc := Service{repository: &repo}

			got, err := svc.FetchUser(context.TODO(), tc.givenID)
			require.True(t, repoCalled)

			if tc.expectedError != nil {
				require.ErrorIs(t, err, tc.expectedError)
				return
			}
			assert.Equal(t, tc.expected, got)
		})
	}
}
