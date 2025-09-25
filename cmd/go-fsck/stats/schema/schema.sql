-- Table for packages
CREATE TABLE packages (
    id TEXT PRIMARY KEY,
    name TEXT,
    import_path TEXT,
    path TEXT,
    test_package INTEGER
);

-- Table for imports
CREATE TABLE imports (
    package_id TEXT,
    path TEXT,
    file TEXT,
    import TEXT,
    FOREIGN KEY(package_id) REFERENCES packages(id)
);

-- Table for functions
CREATE TABLE funcs (
    package_id TEXT,
    path TEXT,
    file TEXT,
    name TEXT,
    receiver TEXT,
    signature TEXT,
    complexity_cognitive INTEGER,
    complexity_cyclomatic INTEGER,
    FOREIGN KEY(package_id) REFERENCES packages(id)
);

-- Table for variables
CREATE TABLE vars (
    package_id TEXT,
    path TEXT,
    file TEXT,
    name TEXT,
    type TEXT,
    FOREIGN KEY(package_id) REFERENCES packages(id)
);

-- Table for constants
CREATE TABLE consts (
    package_id TEXT,
    path TEXT,
    file TEXT,
    name TEXT,
    type TEXT,
    FOREIGN KEY(package_id) REFERENCES packages(id)
);

-- Table for references
CREATE TABLE references (
    package_id TEXT,
    path TEXT,
    file TEXT,
    from_symbol TEXT,
    to_symbol TEXT,
    FOREIGN KEY(package_id) REFERENCES packages(id)
);
