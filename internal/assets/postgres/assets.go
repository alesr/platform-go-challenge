package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alesr/platform-go-challenge/internal/assets"
)

func (r *Repository) StoreAsset(ctx context.Context, asset assets.Asseter) error {
	switch v := asset.(type) {
	case assets.ChartAsset:
		return r.storeChartAsset(ctx, v)
	case assets.InsightAsset:
		return r.storeInsightAsset(ctx, v)
	case assets.AudienceAsset:
		return r.storeAudienceAsset(ctx, v)
	default:
		return errors.New("unsupported asset type")
	}
}

func (r *Repository) ListAssets(ctx context.Context, params *assets.ListAssetsParams) ([]assets.Asseter, string, error) {
	var lastID string
	if params.PageToken != "" {
		lastID = params.PageToken
	}

	query := `
    SELECT * FROM (
        (SELECT
            id,
            'CHART' as asset_type,
            title,
            x_axis,
            y_axis,
            data,
            NULL as insight_data,
            NULL as gender,
            NULL as birth_country,
            NULL::integer as age_min,
            NULL::integer as age_max,
            NULL::integer as social_media_hours,
            NULL::integer as last_month_purchases,
            created_at,
            updated_at
        FROM chart_assets)
        UNION ALL
        (SELECT
            id,
            'INSIGHT' as asset_type,
            NULL as title,
            NULL as x_axis,
            NULL as y_axis,
            NULL::float[] as data,
            data as insight_data,
            NULL as gender,
            NULL as birth_country,
            NULL::integer as age_min,
            NULL::integer as age_max,
            NULL::integer as social_media_hours,
            NULL::integer as last_month_purchases,
            created_at,
            updated_at
        FROM insight_assets)
        UNION ALL
        (SELECT
            id,
            'AUDIENCE' as asset_type,
            NULL as title,
            NULL as x_axis,
            NULL as y_axis,
            NULL::float[] as data,
            NULL as insight_data,
            gender,
            birth_country,
            age_min,
            age_max,
            social_media_hours,
            last_month_purchases,
            created_at,
            updated_at
        FROM audience_assets)
    ) combined
    WHERE ($1 = '' OR combined.id > $1)
    ORDER BY combined.id
    LIMIT $2`

	rows, err := r.db.Query(ctx, query, lastID, params.PageSize)
	if err != nil {
		return nil, "", fmt.Errorf("could not query assets: %w", err)
	}
	defer rows.Close()

	factory := assets.NewAssetFactory()
	var (
		result     []assets.Asseter
		lastIDSeen string
	)

	for rows.Next() {
		var (
			id                 string
			assetType          string
			title              sql.NullString
			xAxis              sql.NullString
			yAxis              sql.NullString
			data               []float64
			insightData        sql.NullString
			gender             sql.NullString
			birthCountry       sql.NullString
			ageMin             sql.NullInt32
			ageMax             sql.NullInt32
			socialMediaHours   sql.NullInt32
			lastMonthPurchases sql.NullInt32
			createdAt          time.Time
			updatedAt          time.Time
		)

		if err := rows.Scan(
			&id,
			&assetType,
			&title,
			&xAxis,
			&yAxis,
			&data,
			&insightData,
			&gender,
			&birthCountry,
			&ageMin,
			&ageMax,
			&socialMediaHours,
			&lastMonthPurchases,
			&createdAt,
			&updatedAt,
		); err != nil {
			return nil, "", fmt.Errorf("could not scan asset: %w", err)
		}

		var asset assets.Asseter
		switch assets.AssetType(assetType) {
		case assets.TypeAssetChart:
			chart := factory.CreateChart(title.String, xAxis.String, yAxis.String, data)
			chart.CreatedAt = createdAt
			chart.UpdatedAt = updatedAt
			chart.ID = id
			asset = chart

		case assets.TypeAssetInsight:
			insight := factory.CreateInsight(insightData.String)
			insight.CreatedAt = createdAt
			insight.UpdatedAt = updatedAt
			insight.ID = id
			asset = insight

		case assets.TypeAssetAudience:
			audience := factory.CreateAudience(
				gender.String,
				birthCountry.String,
				int(ageMin.Int32),
				int(ageMax.Int32),
				int(socialMediaHours.Int32),
				int(lastMonthPurchases.Int32),
			)
			audience.CreatedAt = createdAt
			audience.UpdatedAt = updatedAt
			audience.ID = id
			asset = audience
		}
		result = append(result, asset)
		lastIDSeen = id
	}

	if err := rows.Err(); err != nil {
		return nil, "", fmt.Errorf("could not iterate over rows: %w", err)
	}
	return result, lastIDSeen, nil
}

// Internal

func (r *Repository) storeChartAsset(ctx context.Context, asset assets.ChartAsset) error {
	if _, err := r.db.Exec(ctx, `
        INSERT INTO chart_assets (id, title, x_axis, y_axis, data, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		asset.ID,
		asset.Data.Title,
		asset.Data.XAxis,
		asset.Data.YAxis,
		asset.Data.Data,
		asset.CreatedAt,
		asset.UpdatedAt,
	); err != nil {
		return fmt.Errorf("could not insert chart asset: %w", err)
	}
	return nil
}

func (r *Repository) storeInsightAsset(ctx context.Context, asset assets.InsightAsset) error {
	if _, err := r.db.Exec(ctx, `
        INSERT INTO insight_assets (id, data, created_at, updated_at)
        VALUES ($1, $2, $3, $4)`,
		asset.ID,
		asset.Data.Insight,
		asset.CreatedAt,
		asset.UpdatedAt,
	); err != nil {
		return fmt.Errorf("could not insret insight asset: %w", err)
	}
	return nil
}

func (r *Repository) storeAudienceAsset(ctx context.Context, asset assets.AudienceAsset) error {
	if _, err := r.db.Exec(ctx, `
        INSERT INTO audience_assets (
            id, gender, birth_country, age_min, age_max,
            social_media_hours, last_month_purchases, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		asset.ID,
		asset.Data.Gender,
		asset.Data.BirthCountry,
		asset.Data.AgeMin,
		asset.Data.AgeMax,
		asset.Data.SocialMediaHours,
		asset.Data.LastMonthPurchases,
		asset.CreatedAt,
		asset.UpdatedAt,
	); err != nil {
		return fmt.Errorf("could not insert audience asset: %w", err)
	}
	return nil
}
