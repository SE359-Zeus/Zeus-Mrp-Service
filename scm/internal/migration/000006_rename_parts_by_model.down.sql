PRAGMA foreign_keys=off;

-- Rollback rename if needed
ALTER TABLE parts_by_models RENAME TO parts_by_model;

PRAGMA foreign_keys=on;
