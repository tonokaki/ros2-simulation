-- +migrate Up
CREATE TABLE locations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    x FLOAT NOT NULL,
    y FLOAT NOT NULL,
    floor VARCHAR(10) NOT NULL DEFAULT '1F',
    location_type VARCHAR(50) NOT NULL DEFAULT 'room'
);

-- +migrate Down
DROP TABLE IF EXISTS locations;
