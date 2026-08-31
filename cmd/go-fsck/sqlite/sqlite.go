package sqlite

import (
	"context"
	"errors"
	"os"

	"github.com/go-bridget/mig/db"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/tools/splint/model"
)

func getDefinitions(cfg *options) (model.DefinitionList, error) {
	return internal.DefinitionsFrom(cfg.inputFile, internal.Options(".", false, false, false, cfg.verbose))
}

func sqliteRun(cfg *options) error {
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
		Credentials: db.NewCredentials("sqlite://file:go-fsck.db"),
	})

	if err != nil {
		return err
	}

	if create {
		for _, stmt := range Statements() {
			conn.MustExec(stmt)
		}

		for _, def := range defs {
			if err := Store(conn, def); err != nil {
				return err
			}
		}
	}

	if err := Stats(conn); err != nil {
		return err
	}

	return nil
}
