PRAGMA foreign_keys=off;

-- Rename legacy singular table to plural to match GORM naming
ALTER TABLE parts_by_model RENAME TO parts_by_models;

PRAGMA foreign_keys=on;
