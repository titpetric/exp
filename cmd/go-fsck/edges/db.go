package edges

import (
	"database/sql"
	_ "embed"
	"fmt"

	_ "github.com/mattn/go-sqlite3"
)

//go:embed schema/edges.up.sql
var schemaUp string

// DB wraps a SQLite database for storing and querying edges and relationships.
type DB struct {
	conn *sql.DB
}

// NewDB creates a new database connection and initializes the schema.
// Pass ":memory:" for an in-memory database or a file path for persistent storage.
func NewDB(dsn string) (*DB, error) {
	conn, err := sql.Open("sqlite3", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test the connection
	if err := conn.Ping(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	db := &DB{conn: conn}

	// Initialize schema
	if err := db.Create(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to create schema: %w", err)
	}

	return db, nil
}

// Create initializes the database schema.
func (db *DB) Create() error {
	_, err := db.conn.Exec(schemaUp)
	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}
	return nil
}

// Close closes the database connection.
func (db *DB) Close() error {
	if db.conn != nil {
		return db.conn.Close()
	}
	return nil
}

// InsertSymbols inserts edges (symbol definitions) into the database.
func (db *DB) InsertSymbols(edges []*Edge) error {
	if len(edges) == 0 {
		return nil
	}

	stmt, err := db.conn.Prepare(`
		INSERT OR IGNORE INTO symbols 
		(import_path, symbol_name, receiver, symbol_kind, is_exported, file, line)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, edge := range edges {
		_, err := stmt.Exec(
			edge.ImportPath,
			edge.SymbolName,
			edge.Receiver,
			edge.SymbolKind,
			edge.IsExported,
			edge.File,
			edge.Line,
		)
		if err != nil {
			return fmt.Errorf("failed to insert symbol %s: %w", edge.SymbolID(), err)
		}
	}

	return nil
}

// InsertRelationships inserts relationships between symbols into the database.
func (db *DB) InsertRelationships(relationships []*Relationship) error {
	if len(relationships) == 0 {
		return nil
	}

	// First, fetch the IDs of all referenced symbols
	symbolIDMap := make(map[string]int64)

	stmt, err := db.conn.Prepare("SELECT id FROM symbols WHERE import_path = ? AND symbol_name = ? AND receiver = ?")
	if err != nil {
		return fmt.Errorf("failed to prepare symbol lookup statement: %w", err)
	}
	defer stmt.Close()

	for _, rel := range relationships {
		// Get From symbol ID
		if _, ok := symbolIDMap[rel.From.SymbolID()]; !ok {
			var id int64
			err := stmt.QueryRow(rel.From.ImportPath, rel.From.SymbolName, rel.From.Receiver).Scan(&id)
			if err != nil {
				if err == sql.ErrNoRows {
					// Symbol not found; skip relationship
					continue
				}
				return fmt.Errorf("failed to lookup symbol %s: %w", rel.From.SymbolID(), err)
			}
			symbolIDMap[rel.From.SymbolID()] = id
		}

		// Get To symbol ID
		if _, ok := symbolIDMap[rel.To.SymbolID()]; !ok {
			var id int64
			err := stmt.QueryRow(rel.To.ImportPath, rel.To.SymbolName, rel.To.Receiver).Scan(&id)
			if err != nil {
				if err == sql.ErrNoRows {
					// Symbol not found; skip relationship
					continue
				}
				return fmt.Errorf("failed to lookup symbol %s: %w", rel.To.SymbolID(), err)
			}
			symbolIDMap[rel.To.SymbolID()] = id
		}
	}

	// Insert relationships
	relStmt, err := db.conn.Prepare(`
		INSERT INTO relationships (from_id, to_id, relationship_type, details)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare relationship insert statement: %w", err)
	}
	defer relStmt.Close()

	for _, rel := range relationships {
		fromID, ok := symbolIDMap[rel.From.SymbolID()]
		if !ok {
			continue // Skip if symbol not found
		}

		toID, ok := symbolIDMap[rel.To.SymbolID()]
		if !ok {
			continue // Skip if symbol not found
		}

		_, err := relStmt.Exec(fromID, toID, rel.Type, rel.Details)
		if err != nil {
			return fmt.Errorf("failed to insert relationship %s: %w", rel.String(), err)
		}
	}

	return nil
}

// InsertAll inserts both symbols and relationships in a single transaction.
func (db *DB) InsertAll(edges []*Edge, relationships []*Relationship) error {
	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert symbols
	stmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO symbols 
		(import_path, symbol_name, receiver, symbol_kind, is_exported, file, line)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare insert statement: %w", err)
	}
	defer stmt.Close()

	for _, edge := range edges {
		_, err := stmt.Exec(
			edge.ImportPath,
			edge.SymbolName,
			edge.Receiver,
			edge.SymbolKind,
			edge.IsExported,
			edge.File,
			edge.Line,
		)
		if err != nil {
			return fmt.Errorf("failed to insert symbol %s: %w", edge.SymbolID(), err)
		}
	}

	// Insert relationships
	relStmt, err := tx.Prepare(`
		INSERT INTO relationships (from_id, to_id, relationship_type, details)
		VALUES (
			(SELECT id FROM symbols WHERE import_path = ? AND symbol_name = ? AND (receiver = ? OR (receiver IS NULL AND ? IS NULL))),
			(SELECT id FROM symbols WHERE import_path = ? AND symbol_name = ? AND (receiver = ? OR (receiver IS NULL AND ? IS NULL))),
			?,
			?
		)
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare relationship insert statement: %w", err)
	}
	defer relStmt.Close()

	for _, rel := range relationships {
		result, err := relStmt.Exec(
			rel.From.ImportPath, rel.From.SymbolName, rel.From.Receiver, rel.From.Receiver,
			rel.To.ImportPath, rel.To.SymbolName, rel.To.Receiver, rel.To.Receiver,
			rel.Type,
			rel.Details,
		)
		if err != nil {
			return fmt.Errorf("failed to insert relationship %s: %w", rel.String(), err)
		}

		// Check if any rows were inserted
		rows, _ := result.RowsAffected()
		if rows == 0 {
			// One or both symbols don't exist; skip silently
			continue
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// nullableString converts an empty string to nil for SQL NULL.
func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

// SymbolCount returns the total number of symbols in the database.
func (db *DB) SymbolCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM symbols").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count symbols: %w", err)
	}
	return count, nil
}

// RelationshipCount returns the total number of relationships in the database.
func (db *DB) RelationshipCount() (int, error) {
	var count int
	err := db.conn.QueryRow("SELECT COUNT(*) FROM relationships").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count relationships: %w", err)
	}
	return count, nil
}
