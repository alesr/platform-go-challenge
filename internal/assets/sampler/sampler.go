// Package generator provides a utility function to create sample assets for demonstration purposes.
// It generates random instances of different asset types (Chart, Insight, and Audience).
// To allow users to favorite assets during the Challenge's demonstration, we need existing assets,
// hence this generator. It is implemented as a subpackage under assets to 1) clearly separate
// non-business related concerns, and 2) allow easy removal when it becomes no longer necessary.
package sampler

import (
	"math/rand"
	"strconv"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/go-faker/faker/v4"
)

var (
	genders   = []string{"Male", "Female", "Other"}
	countries = []string{"Germany", "Italy", "Brazil", "Portugal", "Slovenia", "Sweden", "Romania", "Greece"}
)

func SampleAssets(n int) []assets.Asseter {
	samples := make([]assets.Asseter, n)
	factory := assets.NewAssetFactory()

	for i := 0; i < n; i++ {
		assetType := rand.Intn(3)

		switch assetType {
		case 0:
			data := make([]float64, rand.Intn(5)+1)
			for j := range data {
				data[j] = rand.Float64() * 100
			}
			samples[i] = factory.CreateChart(
				"Chart "+strconv.Itoa(i),
				"X Axis",
				"Y Axis",
				data,
			)

		case 1:
			samples[i] = factory.CreateInsight(faker.Sentence())

		case 2:
			samples[i] = factory.CreateAudience(
				genders[rand.Intn(len(genders))],
				countries[rand.Intn(len(countries))],
				rand.Intn(60),    // ageMin
				rand.Intn(20)+60, // ageMax
				rand.Intn(9000),  // socialMediaHours
				rand.Intn(100),   // lastMonthPurchases
			)
		}
	}
	return samples
}
