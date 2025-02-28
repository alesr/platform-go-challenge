package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/alesr/platform-go-challenge/internal/assets"
	"github.com/alesr/platform-go-challenge/internal/assets/favorites"
	"github.com/jackc/pgx/v5"
	"github.com/oklog/ulid/v2"
)

// StoreFavoriteAsset stores a favorite asset in the database.
// It does in a transaction to guarantee the asset is not removed while the user is favoriting it.
func (r *Repository) StoreFavoriteAsset(ctx context.Context, params *favorites.FavoriteAssetParams) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("could not begin transaction: %w", err)
	}
	defer func() {
		if err != nil {
			if rollbackErr := tx.Rollback(ctx); rollbackErr != nil {
				err = fmt.Errorf("could not rollback transaction: %v (original error: %w)", rollbackErr, err)
			}
			return
		}
		if commitErr := tx.Commit(ctx); commitErr != nil {
			err = fmt.Errorf("committing transaction: %w", commitErr)
		}
	}()

	var assetType sql.NullString
	err = tx.QueryRow(ctx, `
        SELECT
            CASE
                WHEN EXISTS (SELECT 1 FROM chart_assets WHERE id = $1) THEN 'CHART'
                WHEN EXISTS (SELECT 1 FROM insight_assets WHERE id = $1) THEN 'INSIGHT'
                WHEN EXISTS (SELECT 1 FROM audience_assets WHERE id = $1) THEN 'AUDIENCE'
            END
        `, params.AssetID).Scan(&assetType)

	if err != nil {
		return fmt.Errorf("could not get asset type: %w", err)
	}

	if !assetType.Valid || assetType.String == "" {
		return assets.ErrAssetNotFound
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO user_favorites (
            id, user_id, asset_id, asset_type, description, created_at, updated_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $6)
        ON CONFLICT (user_id, asset_id) DO UPDATE SET
            description = $5,
            updated_at = $6`,
		ulid.Make().String(),
		params.UserID,
		params.AssetID,
		assetType.String,
		params.Description,
		time.Now(),
	)

	if err != nil {
		return fmt.Errorf("could not insert favorite: %w", err)
	}
	return nil
}

func (r *Repository) GetUserFavorites(ctx context.Context, userID string) ([]favorites.FavoriteAsset, error) {
	rows, err := r.db.Query(ctx, `
        SELECT id, user_id, asset_id, asset_type, description, created_at, updated_at
        FROM user_favorites
        WHERE user_id = $1
        ORDER BY created_at DESC`,
		userID,
	)
	if err != nil {
		return nil, fmt.Errorf("could not run query: %w", err)
	}
	defer rows.Close()

	var result []favorites.FavoriteAsset
	for rows.Next() {
		var f favorites.FavoriteAsset
		if err := rows.Scan(
			&f.ID,
			&f.UserID,
			&f.AssetID,
			&f.AssetType,
			&f.Description,
			&f.CreatedAt,
			&f.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("could not scan favorite: %w", err)
		}
		result = append(result, f)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("could not iterate over rows: %w", err)
	}
	return result, nil
}

func (r *Repository) UpdateFavorite(ctx context.Context, favID, userID string, params *favorites.UpdateFavoriteParams) (*favorites.FavoriteAsset, error) {
	var favorite favorites.FavoriteAsset
	err := r.db.QueryRow(ctx, `
        UPDATE user_favorites
        SET description = $1, updated_at = $2
        WHERE id = $3 AND user_id = $4
        RETURNING id, user_id, asset_id, asset_type, description, created_at, updated_at`,
		params.Description, time.Now(), favID, userID,
	).Scan(
		&favorite.ID,
		&favorite.UserID,
		&favorite.AssetID,
		&favorite.AssetType,
		&favorite.Description,
		&favorite.CreatedAt,
		&favorite.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, favorites.ErrFavoriteAssetNotFound
		}
		return nil, fmt.Errorf("couild not execute update: %w", err)
	}
	return &favorite, nil
}

func (r *Repository) DeleteFavorite(ctx context.Context, favoriteID, userID string) error {
	result, err := r.db.Exec(ctx, `
        DELETE FROM user_favorites
        WHERE id = $1 AND user_id = $2`,
		favoriteID, userID,
	)
	if err != nil {
		return fmt.Errorf("deleting favorite: %w", err)
	}
	if result.RowsAffected() == 0 {
		return favorites.ErrFavoriteAssetNotFound
	}
	return nil
}
