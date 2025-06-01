package sampler

import (
	"fmt"

	"github.com/alesr/platform-go-challenge/internal/users"
	"github.com/go-faker/faker/v4"
	"github.com/go-faker/faker/v4/pkg/options"
	"github.com/oklog/ulid/v2"
)

func SampleUsers(n int) ([]users.User, error) {
	samples := make([]users.User, n)
	for i := 0; i < n; i++ {
		var u users.User
		if err := faker.FakeData(&u, options.WithFieldsToIgnore("ID")); err != nil {
			return nil, fmt.Errorf("could not sample user with fake data: %w", err)
		}
		u.ID = ulid.Make()
		samples[i] = u
	}
	return samples, nil
}
