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

	var cnt int64
	db.Table("part_catalogs").Count(&cnt)
	fmt.Println("part_catalogs:", cnt)

	db.Table("product_models").Count(&cnt)
	fmt.Println("product_models:", cnt)

	db.Table("products").Count(&cnt)
	fmt.Println("products:", cnt)

	// Check specific known BOM item from parts.json
	var cCnt int64
	db.Table("part_catalogs").Where("part_number = ? AND mfg_number = ?", "5W10V25844", "SBB1G28019").Count(&cCnt)
	fmt.Println("catalog match for 5W10V25844|SBB1G28019:", cCnt)

	// print distinct part_numbers for quick inspection
	rows, err := db.Raw("SELECT part_number, mfg_number FROM part_catalogs LIMIT 10").Rows()
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("sample part_catalogs (part_number|mfg_number):")
	for rows.Next() {
		var pn, mn string
		rows.Scan(&pn, &mn)
		fmt.Printf(" - %s|%s\n", pn, mn)
	}

	// Check whether parts_by_model (singular) or parts_by_models (plural) table exists
	var tblName string
	row := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='parts_by_model'").Row()
	_ = row.Scan(&tblName)
	if tblName == "" {
		fmt.Println("parts_by_model table: NOT FOUND")
	} else {
		fmt.Println("parts_by_model table: FOUND")
	}

	fmt.Println("parts_by_model columns:")
	colRows, err := db.Raw("PRAGMA table_info('parts_by_model')").Rows()
	if err == nil {
		defer colRows.Close()
		for colRows.Next() {
			var cid int
			var name, ctype string
			var notnull, pk int
			var dflt interface{}
			colRows.Scan(&cid, &name, &ctype, &notnull, &dflt, &pk)
			fmt.Printf(" - %s %s (pk=%d)\n", name, ctype, pk)
		}
	}

	// Check plural table
	var tblPlural string
	row2 := db.Raw("SELECT name FROM sqlite_master WHERE type='table' AND name='parts_by_models'").Row()
	_ = row2.Scan(&tblPlural)
	if tblPlural == "" {
		fmt.Println("parts_by_models table: NOT FOUND")
	} else {
		fmt.Println("parts_by_models table: FOUND")
		var cnt int64
		db.Table("parts_by_models").Count(&cnt)
		fmt.Println("parts_by_models count:", cnt)
	}
}
