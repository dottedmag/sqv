package sqv

import (
	"context"
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"
	_ "modernc.org/sqlite"
)

func requirePragmaInt(t *testing.T, db *sql.DB, name string, expectedVal int) {
	var val int
	require.NoError(t, db.QueryRow("PRAGMA "+name).Scan(&val))
	require.Equal(t, expectedVal, val)
}

func requireVer(t *testing.T, db *sql.DB, expectedVer int) {
	requirePragmaInt(t, db, "user_version", expectedVer)
}

func requireAppID(t *testing.T, db *sql.DB, expectedAppID int) {
	requirePragmaInt(t, db, "application_id", expectedAppID)
}

func hasTable(t *testing.T, db *sql.DB, name string) bool {
	var count int
	require.NoError(t, db.QueryRow("SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name=?1", name).Scan(&count))
	return count == 1
}

var testSchema = []string{
	`CREATE TABLE a (colA INT)`,
	`ALTER TABLE a ADD COLUMN colB STRING`,
	`ALTER TABLE a RENAME colB TO colC`,
	`ALTER TABLE a DROP colC`,
	`CREATE TABLE b (colB INT)`,
}

func TestOK(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	requireVer(t, db, 0)

	require.NoError(t, Apply(context.Background(), db, 0x1, testSchema))
	requireVer(t, db, len(testSchema))
	requireAppID(t, db, 0x1)
	require.True(t, hasTable(t, db, "a"))
	require.True(t, hasTable(t, db, "b"))

	// Second application is no-op

	require.NoError(t, Apply(context.Background(), db, 0x1, testSchema))
	requireVer(t, db, len(testSchema))
}

func TestPiecemeal(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	for i := 1; i < len(testSchema); i++ {
		require.NoError(t, Apply(context.Background(), db, 0x1, testSchema[:i]))
		requireVer(t, db, i)
	}
}

var failSchema = []string{
	`CREATE TABLE a (colA INT)`, // first Apply, should succeed
	`CREATE TABLE b (colB INT)`, // second Apply, should revert
	`WHOAAAAA`,                  // second Apply, should revert
}

func TestFail(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, Apply(context.Background(), db, 1, failSchema[:1]))
	requireVer(t, db, 1)
	require.True(t, hasTable(t, db, "a"))
	require.False(t, hasTable(t, db, "b"))

	require.Error(t, Apply(context.Background(), db, 1, failSchema))
	requireVer(t, db, 1)
	require.True(t, hasTable(t, db, "a"))
	require.False(t, hasTable(t, db, "b"))
}

func TestOKAppID(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, Apply(context.Background(), db, 1, nil))
	require.NoError(t, Apply(context.Background(), db, 1, nil))
	require.NoError(t, Apply(context.Background(), db, 1, testSchema))
}

func TestWrongAppID(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, Apply(context.Background(), db, 1, testSchema))
	require.Error(t, Apply(context.Background(), db, 0xdead, testSchema))
}

func TestWrongAppIDEmptyDB(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, Apply(context.Background(), db, 1, nil))
	require.Error(t, Apply(context.Background(), db, 0xdead, nil))
}

func TestShortSchema(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	require.NoError(t, err)
	defer db.Close()

	require.NoError(t, Apply(context.Background(), db, 1, testSchema))
	require.Error(t, Apply(context.Background(), db, 1, testSchema[:2]))
}
