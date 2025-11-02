package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/omkar273/nashikdarshan/internal/config"
)

// KeyMigration represents the migration from old to new key structure
type KeyMigration struct {
	OldConfig struct {
		URL        string `json:"url"`
		Key        string `json:"key"`
		ServiceKey string `json:"service_key"`
	} `json:"old_config"`
	NewConfig struct {
		URL            string `json:"url"`
		PublishableKey string `json:"publishable_key"`
		SecretKey      string `json:"secret_key"`
	} `json:"new_config"`
}

// LegacyConfig represents the old configuration structure
type LegacyConfig struct {
	Supabase struct {
		URL        string `mapstructure:"url"`
		Key        string `mapstructure:"key"`
		ServiceKey string `mapstructure:"service_key"`
	} `mapstructure:"supabase"`
}

func main() {
	fmt.Println("🔑 Supabase Key Migration Tool")
	fmt.Println("==============================")
	fmt.Println()

	// Load current config
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	fmt.Println("Current configuration:")
	fmt.Printf("  URL: %s\n", cfg.Supabase.URL)
	fmt.Printf("  Publishable Key: %s\n", maskKey(cfg.Supabase.PublishableKey))
	fmt.Printf("  Secret Key: %s\n", maskKey(cfg.Supabase.SecretKey))
	fmt.Println()

	// Check if we're already using new key structure
	if cfg.Supabase.PublishableKey != "" && cfg.Supabase.SecretKey != "" {
		fmt.Println("✅ Already using new key structure!")
		fmt.Printf("  Publishable Key: %s\n", maskKey(cfg.Supabase.PublishableKey))
		fmt.Printf("  Secret Key: %s\n", maskKey(cfg.Supabase.SecretKey))
		return
	}

	// Migration steps
	fmt.Println("📋 Migration Steps:")
	fmt.Println("1. Go to your Supabase Dashboard")
	fmt.Println("2. Navigate to Settings > API")
	fmt.Println("3. Create new keys:")
	fmt.Println("   - Publishable Key (for client-side use)")
	fmt.Println("   - Secret Key (for server-side use)")
	fmt.Println("4. Update your configuration with the new keys")
	fmt.Println()

	// Generate migration template
	migration := KeyMigration{
		OldConfig: struct {
			URL        string `json:"url"`
			Key        string `json:"key"`
			ServiceKey string `json:"service_key"`
		}{
			URL:        cfg.Supabase.URL,
			Key:        "old_anon_key_here",
			ServiceKey: "old_service_role_key_here",
		},
		NewConfig: struct {
			URL            string `json:"url"`
			PublishableKey string `json:"publishable_key"`
			SecretKey      string `json:"secret_key"`
		}{
			URL:            cfg.Supabase.URL,
			PublishableKey: "pk_your_new_publishable_key_here",
			SecretKey:      "sk_your_new_secret_key_here",
		},
	}

	// Save migration template
	migrationFile := "key_migration_template.json"
	migrationData, err := json.MarshalIndent(migration, "", "  ")
	if err != nil {
		log.Fatalf("Failed to marshal migration data: %v", err)
	}

	if err := os.WriteFile(migrationFile, migrationData, 0644); err != nil {
		log.Fatalf("Failed to write migration file: %v", err)
	}

	fmt.Printf("📄 Migration template saved to: %s\n", migrationFile)
	fmt.Println()

	// Generate config update commands
	fmt.Println("🔧 Configuration Update Commands:")
	fmt.Println()

	// Environment variables
	fmt.Println("Environment Variables:")
	fmt.Printf("export CAYGNUS_SUPABASE_PUBLISHABLE_KEY=\"pk_your_new_publishable_key_here\"\n")
	fmt.Printf("export CAYGNUS_SUPABASE_SECRET_KEY=\"sk_your_new_secret_key_here\"\n")
	fmt.Println()

	// Config file update
	fmt.Println("Config File Update (config.yaml):")
	fmt.Println("supabase:")
	fmt.Printf("  url: \"%s\"\n", cfg.Supabase.URL)
	fmt.Println("  publishable_key: \"pk_your_new_publishable_key_here\"  # For client-side use")
	fmt.Println("  secret_key: \"sk_your_new_secret_key_here\"            # For server-side use")
	fmt.Println("  jwks_url: \"https://your-project.supabase.co/auth/v1/.well-known/jwks.json\"")
	fmt.Println("  jwt_secret: \"your_jwt_secret_here\"  # Optional: only used for HMAC fallback")
	fmt.Println()

	// Client code update
	fmt.Println("📱 Client Code Update (Flutter/React/etc.):")
	fmt.Println("// OLD (using anon key):")
	fmt.Println("const supabase = createClient(url, anonKey)")
	fmt.Println()
	fmt.Println("// NEW (using publishable key):")
	fmt.Println("const supabase = createClient(url, publishableKey)")
	fmt.Println()

	// Server code update
	fmt.Println("🖥️  Server Code Update:")
	fmt.Println("// OLD (using service_role key):")
	fmt.Println("const supabase = createClient(url, serviceRoleKey)")
	fmt.Println()
	fmt.Println("// NEW (using secret key):")
	fmt.Println("const supabase = createClient(url, secretKey)")
	fmt.Println()

	// Cleanup steps
	fmt.Println("🧹 Cleanup Steps (after migration):")
	fmt.Println("1. Test your application with new keys")
	fmt.Println("2. Update all client applications")
	fmt.Println("3. Update all server applications")
	fmt.Println("4. Revoke old anon/service_role keys in Supabase Dashboard")
	fmt.Println("5. Remove old keys from your configuration")
	fmt.Println()

	fmt.Println("✅ Migration guide complete!")
	fmt.Println("   Follow the steps above to migrate from anon/service_role keys to publishable/secret keys.")
}

// maskKey masks a key for display purposes
func maskKey(key string) string {
	if len(key) < 8 {
		return "***"
	}
	return key[:4] + "..." + key[len(key)-4:]
}
