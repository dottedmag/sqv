package sqv

import (
	"context"
	"database/sql"
	"fmt"
)

// Apply applies schema upgrades to the SQLite database
//
// appID is the expected pragma `application_id` of the SQLite database.
//
// schema is the set of SQL statements to apply. Every statement increments
// pragma `user_version` of the database.
func Apply(ctx context.Context, db *sql.DB, appID int, schema []string) error {
	var schemaVer int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&schemaVer); err != nil {
		return fmt.Errorf("failed to query user_version: %w", err)
	}

	var existingAppID int
	if err := db.QueryRowContext(ctx, "PRAGMA application_id").Scan(&existingAppID); err != nil {
		return fmt.Errorf("failed to query application_id: %w", err)
	}

	switch {
	case existingAppID == appID:
		// This is fine
	case schemaVer == 0 && existingAppID == 0:
		if _, err := db.ExecContext(ctx, fmt.Sprintf("PRAGMA application_id = %d", appID)); err != nil {
			return fmt.Errorf("failed to set application_id: %w", err)
		}
	default:
		return fmt.Errorf("wrong application_id: expected %08x, got %08x", appID, existingAppID)
	}

	if int(schemaVer) > len(schema) {
		return fmt.Errorf("existing user_version is %d, while schema has only %d versions",
			schemaVer, len(schema))
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin tx to apply schema changes: %w", err)
	}

	for i := int(schemaVer); i < len(schema); i++ {
		if _, err := tx.ExecContext(ctx, schema[i]); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to apply schema version %d: %w", i+1, err)
		}
	}

	if _, err := tx.ExecContext(ctx, fmt.Sprintf("PRAGMA user_version = %d", len(schema))); err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update user_version: %w", err)
	}

	return tx.Commit()
}
