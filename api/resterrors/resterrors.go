// Package resterrors handles the mapping of service and handler errors
// to errors that are want to send to the client.
//
// Errors that are not explicitly mapped in this package are returned
// to the client as internal errors (HTTP 500), so we don't expose internal details.
//
// In an ideal setup, the transport layer should handle syntax errors,
// while the service layer (domain/business logic) should handle semantic errors.
//
// (I hope I haven't missed any error =])
package resterrors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/alesr/platform-go-challenge/internal/app/rest/handlers"
	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/assets/favorites"
	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/alesr/resterr"
)

// ErrorMap maps all possible errors returned by the Platform Go Challenge (PGC) REST API.
var ErrorMap = map[error]resterr.RESTErr{
	// From users service
	users.ErrUserNotFound: e(http.StatusNotFound, "User resource was not found"),

	// From assets service
	assets.ErrAssetNotFound: e(http.StatusNotFound, "Asset resource was not found"),

	// From favorites service

	favorites.ErrInvalidAssetID:        e(http.StatusBadRequest, "Invalid asset ID"),
	favorites.ErrFavoriteAssetNotFound: e(http.StatusNotFound, "Favorite asset not found"),

	// From transport handlers

	handlers.ErrInvalidPageSize:             e(http.StatusBadRequest, "Invalid page size"),
	handlers.ErrInvalidPageMaxResults:       e(http.StatusBadRequest, "Invalid page max results"),
	handlers.ErrInvalidPageToken:            e(http.StatusBadRequest, "Invalid page token"),
	handlers.ErrInvalidFavoriteAssetPayload: e(http.StatusBadRequest, "Invalid request payload to favorite assets"),
	handlers.ErrUserIDRequired:              e(http.StatusBadRequest, "User ID is required"),
	handlers.ErrFavoriteIDRequired:          e(http.StatusBadRequest, "Favorite ID is required"),
	handlers.ErrInvalidUserID:               e(http.StatusBadRequest, "Invalid user ID"),
	handlers.ErrDescriptionMaxLen: e(
		http.StatusBadRequest,
		fmt.Sprintf("Description for favorite asset is too long (max length '%d')", handlers.MaxDescriptionLength),
	),
}

func e(code int, message string) resterr.RESTErr {
	return resterr.RESTErr{StatusCode: code,
		Message: message,
	}
}

// ValidateRestErr validates a RESTErr and returns an error if it is invalid.
// It is used by the resterr error handler initialized in the main package.
func ValidateRestErr(restErr resterr.RESTErr) error {
	if restErr.StatusCode < 400 || restErr.StatusCode >= 600 {
		return errors.New("invalid status code")
	}
	if restErr.Message == "" {
		return errors.New("at least one of the messages is empty")
	}
	return nil
}
