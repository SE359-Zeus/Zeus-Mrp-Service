package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"database/sql"

	_ "modernc.org/sqlite"
)

func main() {
	var (
		dbPath        = flag.String("db", filepath.Join("configs", "sales.db"), "sqlite database path")
		migrationsDir = flag.String("dir", "migrations", "migrations directory")
		direction     = flag.String("direction", "up", "up or down")
	)
	flag.Parse()
	if err := run(context.Background(), *dbPath, *migrationsDir, *direction); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(ctx context.Context, dbPath, migrationsDir, direction string) error {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}
	defer db.Close()
	if err := db.PingContext(ctx); err != nil {
		return err
	}
	if _, err := db.ExecContext(ctx, `PRAGMA foreign_keys = ON`); err != nil {
		return err
	}
	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		return err
	}
	var migrationFiles []string
	for _, file := range files {
		name := file.Name()
		if strings.HasSuffix(name, ".up.sql") && direction == "up" {
			migrationFiles = append(migrationFiles, filepath.Join(migrationsDir, name))
		}
		if strings.HasSuffix(name, ".down.sql") && direction == "down" {
			migrationFiles = append(migrationFiles, filepath.Join(migrationsDir, name))
		}
	}
	sort.Strings(migrationFiles)
	for _, path := range migrationFiles {
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if _, err := db.ExecContext(ctx, string(content)); err != nil {
			return fmt.Errorf("apply %s: %w", path, err)
		}
	}
	return nil
}
