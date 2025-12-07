package sqlite

import (
	"context"
	"errors"
	"os"

	"github.com/go-bridget/mig/db"

	"github.com/titpetric/exp/cmd/go-fsck/internal"
	"github.com/titpetric/exp/cmd/go-fsck/model"
	"github.com/titpetric/exp/cmd/go-fsck/model/loader"
)

func getDefinitions(cfg *options) ([]*model.Definition, error) {
	// Read the exported go-fsck.json data.
	defs, err := loader.ReadFile(cfg.inputFile)
	if err == nil {
		return defs, nil
	}

	// list current local packages
	packages, err := internal.ListPackages(".", ".")
	if err != nil {
		return nil, err
	}

	defs = []*model.Definition{}

	for _, pkg := range packages {
		d, err := loader.Load(pkg, false, cfg.verbose)
		if err != nil {
			return nil, err
		}

		defs = append(defs, d...)
	}

	return defs, nil
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
		Credentials: db.Credentials{
			DSN:    "file:go-fsck.db",
			Driver: "sqlite",
		},
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
