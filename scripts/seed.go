// seed.go creates a test tenant and API key for local development.
//
// Usage:
//
//	go run scripts/seed.go [-dsn <postgres-dsn>]
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/nagicantsleep/k-map/internal/auth"
	"github.com/nagicantsleep/k-map/internal/storage"
	"github.com/nagicantsleep/k-map/migrations"
)

func main() {
	dsn := flag.String("dsn", "postgres://kmap:kmap@localhost:5432/kmap?sslmode=disable", "Postgres DSN")
	flag.Parse()

	db, err := storage.NewPostgresPool(*dsn)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := storage.RunMigrations(db, migrations.FS); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	repo := auth.NewRepository(db)
	ctx := context.Background()

	tenant, err := repo.CreateTenant(ctx, "dev-tenant", "free")
	if err != nil {
		log.Fatalf("create tenant: %v", err)
	}

	rawKey, err := auth.GenerateRawKey()
	if err != nil {
		log.Fatalf("generate key: %v", err)
	}

	keyHash := auth.HashKey(rawKey)

	apiKey, err := repo.CreateAPIKey(ctx, tenant.ID, keyHash)
	if err != nil {
		log.Fatalf("create api key: %v", err)
	}

	fmt.Fprintf(os.Stdout, "Tenant ID:  %s\n", tenant.ID)
	fmt.Fprintf(os.Stdout, "API Key ID: %s\n", apiKey.ID)
	fmt.Fprintf(os.Stdout, "Raw API Key (use in X-API-Key header): %s\n", rawKey)
}
