-- Install Required PostgreSQL Extensions
-- This script should be run FIRST before any other migrations
-- Run this as a database superuser (typically the postgres user)

-- PostGIS extension for geospatial data types and functions
-- Required for: geography(Point, 4326) location fields
CREATE EXTENSION IF NOT EXISTS postgis;

-- Optional but recommended extensions:

-- UUID extension for generating UUIDs (if not using application-level UUID generation)
-- CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- pg_trgm extension for trigram-based text search (if using text search features)
-- CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Verify extensions are installed
SELECT 
    extname as "Extension Name",
    extversion as "Version"
FROM pg_extension
WHERE extname IN ('postgis')
ORDER BY extname;

-- Note: If you need additional extensions, uncomment them above and add to the WHERE clause
