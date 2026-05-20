package pagination

import (
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
)

const (
	DefaultPage  = 1
	DefaultLimit = 15
	DefaultSort  = "created_at"
	DefaultOrder = "desc"
	MaxLimit     = 100
)

type Params struct {
	Page  int
	Limit int
	Sort  string
	Order string
}

type Meta struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	TotalRows  int64 `json:"total_rows"`
	TotalPages int   `json:"total_pages"`
}

type Response struct {
	Data       any  `json:"data"`
	Pagination Meta `json:"pagination"`
}

func Paginate(db *gorm.DB, params Params, dest any, safeSortCols ...string) (*Meta, error) {
	if params.Page < 1 {
		params.Page = DefaultPage
	}
	if params.Limit < 1 || params.Limit > MaxLimit {
		params.Limit = DefaultLimit
	}

	var totalRows int64
	if err := db.Count(&totalRows).Error; err != nil {
		return nil, err
	}

	sortCol := params.Sort
	if sortCol == "" {
		sortCol = DefaultSort
	} else if len(safeSortCols) > 0 {
		ok := false
		for _, col := range safeSortCols {
			if strings.EqualFold(sortCol, col) {
				sortCol = col
				ok = true
				break
			}
		}
		if !ok {
			sortCol = DefaultSort
		}
	}

	order := strings.ToLower(params.Order)
	if order != "asc" && order != "desc" {
		order = DefaultOrder
	}

	offset := (params.Page - 1) * params.Limit
	if err := db.Offset(offset).Limit(params.Limit).Order(fmt.Sprintf("%s %s", sortCol, order)).Find(dest).Error; err != nil {
		return nil, err
	}

	totalPages := int(math.Ceil(float64(totalRows) / float64(params.Limit)))

	return &Meta{
		Page:       params.Page,
		Limit:      params.Limit,
		TotalRows:  totalRows,
		TotalPages: totalPages,
	}, nil
}
