package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRespondWithJSON(t *testing.T) {
	t.Parallel()

	type TestData struct {
		Message string `json:"message"`
	}

	t.Run("success response", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		data := TestData{Message: "foo"}

		RespondWithJSON(rec, http.StatusOK, data)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var response Response[TestData]
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "success", response.Status)
		assert.Equal(t, "foo", response.Data.Message)
	})

	t.Run("error response", func(t *testing.T) {
		t.Parallel()

		rec := httptest.NewRecorder()
		RespondWithError[TestData](rec, http.StatusBadRequest, "something went wrong")

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.Equal(t, "application/json", rec.Header().Get("Content-Type"))

		var response Response[TestData]
		err := json.Unmarshal(rec.Body.Bytes(), &response)
		require.NoError(t, err)

		assert.Equal(t, "error", response.Status)
		assert.Equal(t, "something went wrong", response.Message)
	})
}
