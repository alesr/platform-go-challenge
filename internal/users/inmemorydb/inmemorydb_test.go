package inmemorydb

import (
	"context"
	"testing"
	"time"

	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createTestUser(t *testing.T, name string) users.User {
	t.Helper()
	return users.User{
		ID:        ulid.Make(),
		Name:      name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestListUsers(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		testUsers := []users.User{
			createTestUser(t, "Rigoletto"),
			createTestUser(t, "Mary Jane"),
		}

		repo := NewRepository(testUsers)

		got := repo.ListUsers(context.TODO())

		assert.Len(t, got, 2)
		assert.ElementsMatch(t, testUsers, got)
	})

	t.Run("return empty slice when context is canceled", func(t *testing.T) {
		t.Parallel()

		testUsers := []users.User{
			createTestUser(t, "Rigoletto"),
		}

		repo := NewRepository(testUsers)

		ctx, cancel := context.WithCancel(context.Background())
		cancel() // cancel context immediately

		got := repo.ListUsers(ctx)
		assert.Empty(t, got)
	})

	t.Run("return empty slice for empty repository", func(t *testing.T) {
		t.Parallel()

		repo := NewRepository([]users.User{})

		got := repo.ListUsers(context.TODO())
		assert.Empty(t, got)
	})

	t.Run("context times out during listing", func(t *testing.T) {
		// create a large dataset to force things to take a little longer
		givenUsers := make([]users.User, 1000)
		for i := 0; i < 1000; i++ {
			givenUsers[i] = createTestUser(t, "Test User")
		}

		repo := NewRepository(givenUsers)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond) // quick timeout
		defer cancel()

		// give context time to timeout
		time.Sleep(2 * time.Millisecond)

		got := repo.ListUsers(ctx)
		assert.Empty(t, got)
	})
}

func TestFetchUser(t *testing.T) {
	t.Parallel()

	givenUsers := []users.User{
		createTestUser(t, "Rigoletto"),
		createTestUser(t, "Mary Jane"),
	}

	repo := NewRepository(givenUsers)

	// invalid user to test type casting.
	invalidID := ulid.Make().String()
	repo.store.Store(invalidID, "not a user struct")

	testCases := []struct {
		name          string
		givenUserID   string
		expectedUser  *users.User
		expectedError error
	}{
		{
			name:         "successfully fetchs user",
			givenUserID:  givenUsers[0].ID.String(),
			expectedUser: &givenUsers[0],
		},
		{
			name:          "user not found",
			givenUserID:   "foo",
			expectedError: users.ErrUserNotFound,
		},
		{
			name:          "stored value cannot be cast to user",
			givenUserID:   invalidID,
			expectedError: users.ErrInvalidUserValue,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := repo.FetchUser(context.TODO(), tc.givenUserID)
			require.Equal(t, tc.expectedError, err)

			assert.Equal(t, tc.expectedUser, got)
		})
	}
}
