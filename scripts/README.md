# Database Migration Scripts

This directory contains SQL scripts for database setup and migrations, plus CLI tools for user management.

## Fresh Database Setup (Recommended)

If you're starting with a **fresh/cleared database**, follow this order:

1. **`install_extensions.sql`** - Install required PostgreSQL extensions (PostGIS)
   - Must be run as a database superuser
   - Run this FIRST before any other migrations

2. **Run Ent migrations** - `make migrate-ent` or `go run cmd/migrate/main.go`
   - Creates all tables based on Ent schema

3. **Onboard users** - `go run scripts/main.go -cmd onboard-user --email user@example.com --password pass123`
   - Creates users in Supabase and local database

## Migration from Old Schema

If you're **migrating from an existing database** with old schema:

1. **`install_extensions.sql`** - Install PostGIS extension

2. **`drop_event_hotel_tables.sql`** - Drop Event and Hotel tables (if they exist)
   - Only needed if you have existing Event/Hotel tables to remove

3. **`migrate_to_postgis.sql`** - Migrate Place table to use PostGIS geography
   - Converts latitude/longitude columns to PostGIS location column
   - Only needed if you have existing data with lat/lng columns

4. **Run Ent migrations** - `make migrate-ent` or `go run cmd/migrate/main.go`
   - Updates schema to match current Ent definitions

## User Onboarding Script

The `onboard-user` command automates user creation in both Supabase and your local database.

### Usage

```bash
# Onboard a new user (name will be auto-generated from email)
go run scripts/main.go -cmd onboard-user \
  --email user@example.com \
  --password securepass123

# Onboard with custom name
go run scripts/main.go -cmd onboard-user \
  --email user@example.com \
  --password securepass123 \
  --name "John Doe"
```

### How It Works

1. **Checks Supabase**: Tries to sign in with email/password
   - If user exists: Gets access token
   - If user doesn't exist: Creates new user in Supabase and gets access token

2. **Checks Local Database**: Verifies if user exists in local DB
   - If exists: Skips creation
   - If doesn't exist: Creates user record with Supabase ID

3. **Runs Onboarding**: Executes onboarding service (currently just validates)

4. **Returns Access Token**: Prints access token to stdout for use in API calls

### Requirements

- Supabase configuration must be set in `config.yaml` or environment variables
- Database must be accessible and migrations must be run first
- Supabase SecretKey must have admin privileges for user creation

## Running SQL Scripts

### Using psql command line:

```bash
# Install extensions (requires superuser) - REQUIRED for fresh setup
psql -U postgres -d your_database -f scripts/install_extensions.sql

# Drop old tables (only if migrating from old schema)
psql -U your_user -d your_database -f scripts/drop_event_hotel_tables.sql

# Migrate to PostGIS (only if migrating existing data)
psql -U your_user -d your_database -f scripts/migrate_to_postgis.sql
```

### Using database client:

1. Connect to your database
2. Open and execute each script in order
3. Verify extensions are installed: `SELECT * FROM pg_extension;`

## Extension Requirements

- **PostGIS**: Required for geospatial queries (geography(Point, 4326))
  - Version: 3.0+ recommended
  - Installation: Usually available via package manager or PostgreSQL contrib

## Troubleshooting

### PostGIS not found
- Ensure PostGIS is installed on your PostgreSQL server
- On Ubuntu/Debian: `sudo apt-get install postgresql-postgis`
- On macOS: `brew install postgis`
- On Windows: Install via PostGIS installer

### Permission denied
- Extensions must be installed by a superuser
- Use `postgres` user or a user with CREATE EXTENSION privileges

### Extension already exists
- Scripts use `CREATE EXTENSION IF NOT EXISTS` - safe to run multiple times
- No harm in running install_extensions.sql multiple times
