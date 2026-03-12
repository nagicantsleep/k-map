// Package migrations embeds SQL migration files for use with golang-migrate.
package migrations

import "embed"

// FS contains all SQL migration files.
//
//go:embed *.sql
var FS embed.FS
