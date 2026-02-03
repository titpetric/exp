-- Symbols table: stores all symbol definitions
CREATE TABLE IF NOT EXISTS symbols (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  import_path TEXT NOT NULL,
  symbol_name TEXT NOT NULL,
  receiver TEXT,                    -- NULL for non-methods, receiver type for methods (e.g., "MyType" or "*MyType")
  symbol_kind TEXT NOT NULL,        -- one of: type, var, const, func
  is_exported BOOLEAN NOT NULL,     -- true for exported symbols, false for unexported
  file TEXT NOT NULL,               -- source file (relative path with .go extension)
  line INTEGER NOT NULL,            -- line number where symbol is defined
  UNIQUE(import_path, symbol_name, receiver)
);

-- Relationships table: captures how symbols relate to each other
CREATE TABLE IF NOT EXISTS relationships (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  from_id INTEGER NOT NULL,
  to_id INTEGER NOT NULL,
  relationship_type TEXT NOT NULL,  -- one of: receiver, argument, return, uses, test
  details TEXT,                     -- JSON metadata (e.g., {"index": 0} for argument position)
  FOREIGN KEY(from_id) REFERENCES symbols(id) ON DELETE CASCADE,
  FOREIGN KEY(to_id) REFERENCES symbols(id) ON DELETE CASCADE
);

-- Indexes for efficient queries
CREATE INDEX IF NOT EXISTS idx_symbols_import_path ON symbols(import_path);
CREATE INDEX IF NOT EXISTS idx_symbols_name ON symbols(symbol_name);
CREATE INDEX IF NOT EXISTS idx_symbols_kind ON symbols(symbol_kind);
CREATE INDEX IF NOT EXISTS idx_symbols_path_name ON symbols(import_path, symbol_name);
CREATE INDEX IF NOT EXISTS idx_symbols_receiver ON symbols(receiver);

CREATE INDEX IF NOT EXISTS idx_relationships_from ON relationships(from_id);
CREATE INDEX IF NOT EXISTS idx_relationships_to ON relationships(to_id);
CREATE INDEX IF NOT EXISTS idx_relationships_type ON relationships(relationship_type);
CREATE INDEX IF NOT EXISTS idx_relationships_from_type ON relationships(from_id, relationship_type);
CREATE INDEX IF NOT EXISTS idx_relationships_to_type ON relationships(to_id, relationship_type);
