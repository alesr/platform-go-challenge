CREATE TABLE user_favorites (
    id VARCHAR(127) PRIMARY KEY,
    user_id VARCHAR(127) NOT NULL,
    asset_id VARCHAR(127) NOT NULL,
    asset_type VARCHAR(50) NOT NULL,
    description TEXT,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    CONSTRAINT asset_type_check CHECK (asset_type IN ('CHART', 'INSIGHT', 'AUDIENCE')),
    CONSTRAINT unique_user_asset UNIQUE (user_id, asset_id)
);

-- Supports GetUserFavorites operation which filters by user_id and orders by created_at DESC
-- Slso covers queries filtering by user_id alone
CREATE INDEX idx_user_favorites_user_created ON user_favorites(user_id, created_at DESC);

--Supports UpdateFavorite and DeleteFavorite operations which filter by both id and user_id
CREATE INDEX idx_user_favorites_id_user ON user_favorites(id, user_id);
