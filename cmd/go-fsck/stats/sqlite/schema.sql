-- Table for packages
CREATE TABLE `packages` (
    `id` TEXT PRIMARY KEY,
    `name` TEXT,
    `import_path` TEXT,
    `path` TEXT,
    `test_package` INTEGER
);

-- Table for imports
CREATE TABLE `imports` (
    `package_id` TEXT,
    `path` TEXT,
    `file` TEXT,
    `import` TEXT
);
CREATE INDEX `idx_imports_pkg` ON `imports`(`package_id`);
CREATE INDEX `idx_imports_file` ON `imports`(`file`);
CREATE INDEX `idx_imports_path` ON `imports`(`path`);

-- Table for types
CREATE TABLE `types` (
    `package_id` TEXT,
    `path` TEXT,
    `file` TEXT,
    `name` TEXT,
    `kind` TEXT,
    `doc` TEXT,
    `signature` TEXT
);
CREATE INDEX `idx_types_pkg` ON `types`(`package_id`);
CREATE INDEX `idx_types_file` ON `types`(`file`);
CREATE INDEX `idx_types_path` ON `types`(`path`);
CREATE INDEX `idx_types_name` ON `types`(`name`);

-- Table for struct/interface fields
CREATE TABLE `fields` (
    `type_name` TEXT,
    `package_id` TEXT,
    `path` TEXT,
    `file` TEXT,
    `name` TEXT,
    `type` TEXT,
    `json_name` TEXT,
    `map_key` TEXT,
    `doc` TEXT,
    `comment` TEXT,
    `tag` TEXT
);
CREATE INDEX `idx_fields_pkg` ON `fields`(`package_id`);
CREATE INDEX `idx_fields_file` ON `fields`(`file`);
CREATE INDEX `idx_fields_type` ON `fields`(`type_name`);
CREATE INDEX `idx_fields_name` ON `fields`(`name`);

-- Table for functions
CREATE TABLE `funcs` (
    `package_id` TEXT,
    `path` TEXT,
    `file` TEXT,
    `name` TEXT,
    `type` TEXT,
    `receiver` TEXT,
    `signature` TEXT,
    `complexity_cognitive` INTEGER,
    `complexity_cyclomatic` INTEGER,
    `complexity_lines` INTEGER,
);
CREATE INDEX `idx_funcs_pkg` ON `funcs`(`package_id`);
CREATE INDEX `idx_funcs_file` ON `funcs`(`file`);
CREATE INDEX `idx_funcs_path` ON `funcs`(`path`);
CREATE INDEX `idx_funcs_name` ON `funcs`(`name`);

-- Table for variables
CREATE TABLE `vars` (
    `package_id` TEXT,
    `path` TEXT,
    `file` TEXT,
    `name` TEXT,
    `type` TEXT
);
CREATE INDEX `idx_vars_pkg` ON `vars`(`package_id`);
CREATE INDEX `idx_vars_file` ON `vars`(`file`);
CREATE INDEX `idx_vars_path` ON `vars`(`path`);
CREATE INDEX `idx_vars_name` ON `vars`(`name`);

-- Table for constants
CREATE TABLE `consts` (
    `package_id` TEXT,
    `path` TEXT,
    `file` TEXT,
    `name` TEXT,
    `type` TEXT
);
CREATE INDEX `idx_consts_pkg` ON `consts`(`package_id`);
CREATE INDEX `idx_consts_file` ON `consts`(`file`);
CREATE INDEX `idx_consts_path` ON `consts`(`path`);
CREATE INDEX `idx_consts_name` ON `consts`(`name`);

-- Table for references
CREATE TABLE `references` (
    `package_id` TEXT,
    `path` TEXT,
    `file` TEXT,
    `from_symbol` TEXT,
    `to_symbol` TEXT
);
CREATE INDEX `idx_refs_pkg` ON `references`(`package_id`);
CREATE INDEX `idx_refs_file` ON `references`(`file`);
CREATE INDEX `idx_refs_path` ON `references`(`path`);
CREATE INDEX `idx_refs_from` ON `references`(`from_symbol`);
CREATE INDEX `idx_refs_to` ON `references`(`to_symbol`);
