-- For the sake of this challenge, I'm assuming a scenario where the number
-- of records for each different type can vary significantly.
-- I'm also assuming that some assets may be more frequently requested than others.
-- Spliting the assets can make the query for fetching data a bit more complex,
-- but it can also improve performance, especially when dealing with large datasets.

CREATE TABLE chart_assets (
    id VARCHAR(127) PRIMARY KEY,
    title VARCHAR(255) NOT NULL,
    x_axis VARCHAR(255) NOT NULL,
    y_axis VARCHAR(255) NOT NULL,
    data FLOAT[] NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE insight_assets (
    id VARCHAR(127) PRIMARY KEY,
    data TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);

CREATE TABLE audience_assets (
    id VARCHAR(127) PRIMARY KEY,
    gender VARCHAR(255) NOT NULL,
    birth_country VARCHAR(255) NOT NULL,
    age_min INTEGER NOT NULL,
    age_max INTEGER NOT NULL,
    social_media_hours INTEGER NOT NULL,
    last_month_purchases INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL
);
