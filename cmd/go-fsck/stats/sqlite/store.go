package sqlite

import (
	"fmt"

	"github.com/jmoiron/sqlx"

	"github.com/titpetric/exp/cmd/go-fsck/model"
)

// Store stores a model.Definition into the SQLite database using sqlx.DB
func Store(db *sqlx.DB, def *model.Definition) error {
	tx, err := db.Beginx()
	if err != nil {
		return err
	}

	// Insert package
	_, err = tx.Exec(`
		INSERT OR IGNORE INTO `+"`packages`"+`(
			`+"`id`, `name`, `import_path`, `path`, `test_package`"+`
		) VALUES (?, ?, ?, ?, ?)`,
		def.ID, def.Package.Package, def.Package.ImportPath, def.Package.Path, def.Package.TestPackage,
	)
	if err != nil {
		tx.Rollback()
		return err
	}

	// Insert imports
	for _, imports := range def.Imports {
		for _, imp := range imports {
			_, err = tx.Exec(`
				INSERT INTO `+"`imports`"+`(
					`+"`package_id`, `path`, `file`, `import`"+`
				) VALUES (?, ?, ?, ?)`,
				def.ID, def.Package.Path, "", imp,
			)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	// Insert vars
	for _, v := range def.Vars {
		_, err = tx.Exec(`
			INSERT INTO `+"`vars`"+`(
				`+"`package_id`, `path`, `file`, `name`, `type`"+`
			) VALUES (?, ?, ?, ?, ?)`,
			def.ID, v.File, v.File, v.Name, v.Type,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		// References for var
		for from, tos := range v.References {
			for _, to := range tos {
				_, err = tx.Exec(`
					INSERT INTO `+"`references`"+`(
						`+"`package_id`, `path`, `file`, `from_symbol`, `to_symbol`"+`
					) VALUES (?, ?, ?, ?, ?)`,
					def.ID, v.File, v.File, from, to,
				)
				if err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	// Insert consts
	for _, c := range def.Consts {
		_, err = tx.Exec(`
			INSERT INTO `+"`consts`"+`(
				`+"`package_id`, `path`, `file`, `name`, `type`"+`
			) VALUES (?, ?, ?, ?, ?)`,
			def.ID, c.File, c.File, c.Name, c.Type,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		// References for const
		for from, tos := range c.References {
			for _, to := range tos {
				_, err = tx.Exec(`
					INSERT INTO `+"`references`"+`(
						`+"`package_id`, `path`, `file`, `from_symbol`, `to_symbol`"+`
					) VALUES (?, ?, ?, ?, ?)`,
					def.ID, c.File, c.File, from, to,
				)
				if err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	// Insert funcs
	for _, f := range def.Funcs {
		_, err = tx.Exec(`
			INSERT INTO `+"`funcs`"+`(
				`+"`package_id`, `path`, `file`, `name`, `receiver`, `signature`, `complexity_cognitive`, `complexity_cyclomatic`, `complexity_coverage`, `complexity_lines`"+`
			) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			def.ID, f.File, f.File, f.Name, f.Receiver, f.Signature, f.Complexity.Cognitive, f.Complexity.Cyclomatic, f.Complexity.Coverage, f.Complexity.Lines,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		// References for func
		for from, tos := range f.References {
			for _, to := range tos {
				_, err = tx.Exec(`
					INSERT INTO `+"`references`"+`(
						`+"`package_id`, `path`, `file`, `from_symbol`, `to_symbol`"+`
					) VALUES (?, ?, ?, ?, ?)`,
					def.ID, f.File, f.File, from, to,
				)
				if err != nil {
					tx.Rollback()
					return err
				}
			}
		}
	}

	// Insert types and fields
	for _, t := range def.Types {
		_, err = tx.Exec(`
			INSERT INTO `+"`types`"+`(
				`+"`package_id`, `path`, `file`, `name`, `kind`, `doc`, `signature`"+`
			) VALUES (?, ?, ?, ?, ?, ?, ?)`,
			def.ID, t.File, t.File, t.Name, t.Kind.String(), t.Doc, t.Signature,
		)
		if err != nil {
			tx.Rollback()
			return err
		}

		// Insert fields for the type
		for _, f := range t.Fields {
			_, err = tx.Exec(`
				INSERT INTO `+"`fields`"+`(
					`+"`type_name`, `package_id`, `path`, `file`, `name`, `type`, `json_name`, `map_key`, `doc`, `comment`, `tag`"+`
				) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				t.Name, def.ID, t.File, t.File, f.Name, f.Type, f.JSONName, f.MapKey, f.Doc, f.Comment, f.Tag,
			)
			if err != nil {
				tx.Rollback()
				return err
			}
		}
	}

	return tx.Commit()
}

// Stats prints the count of records in each table using sqlx.DB
func Stats(db *sqlx.DB) error {
	tables := []string{"`packages`", "`imports`", "`types`", "`fields`", "`funcs`", "`vars`", "`consts`", "`references`"}
	for _, table := range tables {
		var count int
		err := db.Get(&count, fmt.Sprintf("SELECT COUNT(*) FROM %s", table))
		if err != nil {
			return err
		}
		fmt.Printf("%s: %d\n", table, count)
	}
	return nil
}
