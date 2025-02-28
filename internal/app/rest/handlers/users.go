package handlers

import (
	"net/http"
	"time"

	"github.com/alesr/platform-go-challenge/internal/pkg/httputil"
	"github.com/alesr/platform-go-challenge/internal/users"
)

type ListUsersResponse struct {
	Data []UserResponse `json:"data"`
}

type UserResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ListUsers defines a handler for listing users.
// NOTE: I'm not implementing pagination here for brevity.
// List assets implements pagination.
func (h *Handler) ListUsers() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		users := h.usersSvc.ListUsers(r.Context())
		items := make([]UserResponse, 0, len(users))
		for _, user := range users {
			items = append(items, toTransportUser(user))
		}
		httputil.RespondWithJSON(w, http.StatusOK, ListUsersResponse{Data: items})
	}
}

func toTransportUser(u users.User) UserResponse {
	// TODO(alesr): doing this because of ulid zero values (000....)
	if u == (users.User{}) {
		return UserResponse{}
	}
	return UserResponse{
		ID:        u.ID.String(),
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}
