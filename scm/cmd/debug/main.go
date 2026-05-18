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

	var count int64
	db.Table("part_types").Count(&count)
	fmt.Println("part_types count:", count)

	db.Table("part_catalogs").Count(&count)
	fmt.Println("part_catalogs count:", count)

	db.Table("parts_by_model").Count(&count)
	fmt.Println("parts_by_model count:", count)

	db.Table("parts").Count(&count)
	fmt.Println("parts count:", count)
}
