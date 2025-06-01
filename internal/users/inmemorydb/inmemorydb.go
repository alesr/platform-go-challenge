package inmemorydb

import (
	"context"
	"sync"

	"github.com/alesr/platform-go-challenge/internal/users"
)

type Repository struct{ store *sync.Map }

func NewRepository(usersSample []users.User) *Repository {
	var store sync.Map
	for _, user := range usersSample {
		store.Store(user.ID.String(), user)
	}
	return &Repository{
		store: &store,
	}
}

func (r *Repository) ListUsers(ctx context.Context) []users.User {
	userStore := make([]users.User, 0)

	r.store.Range(func(key, value any) bool {
		select {
		case <-(ctx).Done():
			return false
		default:
			if user, ok := value.(users.User); ok {
				userStore = append(userStore, user)
			}
			return true
		}
	})

	select {
	case <-(ctx).Done():
		return []users.User{}
	default:
		return userStore
	}
}

func (r *Repository) FetchUser(_ context.Context, id string) (*users.User, error) {
	val, found := r.store.Load(id)
	if !found {
		return nil, users.ErrUserNotFound
	}
	u, ok := val.(users.User)
	if !ok {
		return nil, users.ErrInvalidUserValue
	}
	return &u, nil
}
