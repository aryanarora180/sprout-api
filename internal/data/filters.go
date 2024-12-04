package data

import (
	"fmt"
	"math"
	"sprout.example.com/internal/validator"
	"strings"
	"time"
)

type Filters struct {
	Page         int
	PageSize     int
	Sort         string
	SortSafeList []string
}

func ValidateFilters(v *validator.Validator, f Filters) {
	v.Check(f.Page > 0, "page", "must be greater than zero")
	v.Check(f.Page <= 10_000_000, "page", "must be a maximum of 10 million")

	v.Check(f.PageSize > 0, "page_size", "must be greater than zero")
	v.Check(f.PageSize <= 100, "page_size", "must be a maximum of 100")

	v.Check(validator.In(f.Sort, f.SortSafeList...), "sort", "invalid sort value")
}

func (f Filters) sortColumn() string {
	for _, safeValue := range f.SortSafeList {
		if f.Sort == safeValue {
			return strings.TrimPrefix(f.Sort, "-")
		}
	}

	panic("unsafe sort parameter: " + f.Sort)
}

func (f Filters) sortDirection() string {
	if strings.HasPrefix(f.Sort, "-") {
		return "DESC"
	}

	return "ASC"
}

func (f Filters) limit() int {
	return f.PageSize
}

func (f Filters) offset() int {
	return (f.Page - 1) * f.PageSize
}

type MonthYearFilter struct {
	Month int
	Year  int
}

func ValidateMonthYearFilter(v *validator.Validator, f MonthYearFilter) {
	v.Check(f.Month > 0, "month", "must be between 1-12 (inclusive)")
	v.Check(f.Month < 13, "month", "must be between 1-12 (inclusive)")

	currentYear := time.Now().Year()
	v.Check(f.Year >= currentYear-100, "year", "must be within 100 years from today")
	v.Check(f.Year <= currentYear+100, "year", "must be within 100 years from today")
}

func (myf MonthYearFilter) monthRange() (string, string) {
	startDate := fmt.Sprintf("%d-%02d-01", myf.Year, myf.Month)
	endDate := fmt.Sprintf("%d-%02d-01", myf.Year, myf.Month+1)
	if myf.Month == 12 {
		endDate = fmt.Sprintf("%d-01-01", myf.Year+1)
	}

	return startDate, endDate
}

type Metadata struct {
	CurrentPage  int `json:"current_page,omitempty"`
	PageSize     int `json:"page_size,omitempty"`
	FirstPage    int `json:"first_page,omitempty"`
	LastPage     int `json:"last_page,omitempty"`
	TotalRecords int `json:"total_records,omitempty"`
}

func calculateMetadata(totalRecords, page, pageSize int) Metadata {
	if totalRecords == 0 {
		return Metadata{}
	}

	return Metadata{
		CurrentPage:  page,
		PageSize:     pageSize,
		FirstPage:    1,
		LastPage:     int(math.Ceil(float64(totalRecords) / float64(pageSize))),
		TotalRecords: totalRecords,
	}
}
