package stats

import (
	"context"
	"errors"
	"os"

	"github.com/go-bridget/mig/db"

	"github.com/titpetric/exp/cmd/go-fsck/stats/sqlite"
)

func storeDefinitions(cfg *options) error {
	defs, err := getDefinitions(cfg)
	if err != nil {
		return err
	}

	ctx := context.Background()

	_, err = os.Stat("go-fsck.db")
	create := errors.Is(err, os.ErrNotExist)

	// Aggregations are easier in SQL... the following block of
	// code uses an sqlite in-memory database to do some math.
	conn, err := db.ConnectWithOptions(ctx, &db.Options{
		Credentials: db.Credentials{
			DSN:    "file:go-fsck.db",
			Driver: "sqlite",
		},
	})

	if err != nil {
		return err
	}

	if create {
		for _, stmt := range sqlite.Statements() {
			conn.MustExec(stmt)
		}

		for _, def := range defs {
			if err := sqlite.Store(conn, def); err != nil {
				return err
			}
		}
	}

	if err := sqlite.Stats(conn); err != nil {
		return err
	}

	return nil
}
