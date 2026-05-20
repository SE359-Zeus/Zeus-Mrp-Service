package main

import (
	"fmt"
	"log"

	"zeus-scm-service/internal/repository/sqlite"
)

func main() {
	db, err := sqlite.NewDB("scm.db")
	if err != nil {
		log.Fatal(err)
	}

	// find a catalog id for a known part
	var id string
	row := db.Raw(`SELECT id FROM part_catalogs WHERE part_number = ? AND mfg_number = ? LIMIT 1`, "5W10V25844", "SBB1G28019").Row()
	if err := row.Scan(&id); err != nil {
		log.Fatalf("failed to get catalog id: %v", err)
	}
	fmt.Println("found catalog id:", id)

	// try inserting into parts_by_model
	res := db.Exec(`INSERT OR IGNORE INTO parts_by_model (part_catalog_id, product_model_code, quantity) VALUES (?, ?, ?)`, id, "82SN003JVN", 1)
	if res.Error != nil {
		log.Fatalf("insert error: %v", res.Error)
	}
	fmt.Println("rows affected:", res.RowsAffected)

	var cnt int64
	db.Table("parts_by_model").Count(&cnt)
	fmt.Println("parts_by_model count after insert:", cnt)
}
