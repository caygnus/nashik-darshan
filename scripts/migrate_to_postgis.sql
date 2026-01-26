-- PostGIS Migration Script for Places Table
-- This script migrates the places table from separate latitude/longitude columns to PostGIS geography
-- 
-- IMPORTANT: 
-- 1. Run scripts/install_extensions.sql FIRST to install required extensions
-- 2. This script is ONLY needed if you have existing data with latitude/longitude columns
-- 3. If you're starting with a fresh database, you don't need this script - Ent migrations will create the location column directly
-- 
-- This script assumes PostGIS extension is already installed

-- Verify PostGIS extension is installed (will error if not)
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_extension WHERE extname = 'postgis') THEN
        RAISE EXCEPTION 'PostGIS extension is not installed. Please run scripts/install_extensions.sql first.';
    END IF;
END $$;

-- Step 2: Add location column to places table
-- Note: This assumes the places table already exists with latitude and longitude columns
ALTER TABLE places ADD COLUMN IF NOT EXISTS location geography(Point, 4326);

-- Step 3: Migrate existing data from latitude/longitude to location
-- Only migrate rows where at least one coordinate is non-zero
UPDATE places
SET location = ST_SetSRID(ST_MakePoint(longitude::numeric, latitude::numeric), 4326)::geography
WHERE (latitude != 0 OR longitude != 0)
  AND location IS NULL;

-- Step 4: Create GiST spatial index on location column for efficient geospatial queries
-- This index enables fast ST_DWithin and ST_Distance queries
CREATE INDEX IF NOT EXISTS places_location_gix ON places USING GIST (location);

-- Step 5: Verify migration (optional - can be run manually to check)
-- SELECT id, latitude, longitude, 
--        ST_X(location::geometry) as lng_from_location,
--        ST_Y(location::geometry) as lat_from_location
-- FROM places 
-- WHERE location IS NOT NULL 
-- LIMIT 10;

-- Step 6: Drop old latitude and longitude columns (UNCOMMENT AFTER VERIFICATION)
-- ALTER TABLE places DROP COLUMN IF EXISTS latitude;
-- ALTER TABLE places DROP COLUMN IF EXISTS longitude;

-- Note: Keep latitude and longitude columns until you've verified the migration is correct
-- Then uncomment Step 6 and run it manually
