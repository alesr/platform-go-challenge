package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alesr/platform-go-challenge/internal/pkg/httputil"
	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/oklog/ulid/v2"
	"github.com/stretchr/testify/assert"
)

func TestToTransportUser(t *testing.T) {
	t.Parallel()

	id := ulid.Make()

	testCases := []struct {
		name   string
		given  users.User
		expect UserResponse
	}{
		{
			name: "valid user",
			given: users.User{
				ID:        id,
				Name:      "Xablau",
				CreatedAt: time.Time{}.Add(time.Hour * 48),
				UpdatedAt: time.Time{}.Add(time.Hour * 24),
			},
			expect: UserResponse{
				ID:        id.String(),
				Name:      "Xablau",
				CreatedAt: time.Time{}.Add(time.Hour * 48),
				UpdatedAt: time.Time{}.Add(time.Hour * 24),
			},
		},
		{
			name:   "empty user",
			given:  users.User{},
			expect: UserResponse{},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := toTransportUser(tc.given)
			assert.Equal(t, tc.expect, got)
		})
	}
}

func TestHandler_ListUsers(t *testing.T) {
	t.Parallel()

	usrSvc := usersSvcMock{
		listUsersFunc: func(ctx context.Context) []users.User {
			return []users.User{}
		},
	}

	handler := Handler{
		usersSvc: &usrSvc,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.ListUsers().ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var resp httputil.Response[ListUsersResponse]
	err := json.NewDecoder(rec.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.Equal(t, "success", resp.Status)
	assert.Empty(t, resp.Data.Data)
}
