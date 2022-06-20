# sqv
[![Go Reference](https://pkg.go.dev/badge/github.com/dottedmag/sqv.svg)](https://pkg.go.dev/github.com/dottedmag/sqv)

`sqv` applies DB schema to SQLite database.

	const appID = 0x1234

	db, err := sql.Open("sqlite", filename)
	err := sqv.Apply(db, appID, []string{
		"CREATE TABLE mytable (colmn INT)",
		"ALTER TABLE mytable RENAME colmn TO column",
	})

Schema version is stored via `user_version` pragma. Application ID is stored and checked via `application_id` pragma.

## Legal

Author and copyright: Mikhail Gusarov <dottedmag@dottedmag.net>

Licensed under [Apache 2.0](LICENSE) license.
