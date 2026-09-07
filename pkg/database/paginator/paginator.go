package paginator

import (
	"fmt"
	"math"
	"strings"

	"gorm.io/gorm"
)

type Pagination struct {
	TotalRows  int64 `json:"totalRows"`
	TotalPages int   `json:"totalPages"`
	Data       any   `json:"data"`
	SearchOption
}

type SearchOptionFunc func(*SearchOption)

type SearchOption struct {
	Limit      int                    `json:"limit"`
	Page       int                    `json:"page"`
	OrderBy    string                 `json:"orderBy"`
	Sort       string                 `json:"sort"`
	Attributes map[string]interface{} `json:"attributes"`
}

func defaultSearchOptions() SearchOption {
	return SearchOption{
		Limit:      20,
		Page:       1,
		OrderBy:    "id",
		Sort:       "asc",
		Attributes: make(map[string]interface{}, 0),
	}
}

func NewPagination(opts ...SearchOptionFunc) *Pagination {
	searchOptions := defaultSearchOptions()

	for _, fn := range opts {
		fn(&searchOptions)
	}

	return &Pagination{
		SearchOption: searchOptions,
	}
}

func (p *Pagination) GetOffset() int {
	return (p.GetPage() - 1) * p.GetLimit()
}

func (p *Pagination) GetLimit() int {
	if p.Limit == 0 {
		p.Limit = 10
	}

	return p.Limit
}

func (p *Pagination) GetPage() int {
	if p.Page == 0 {
		p.Page = 1
	}

	return p.Page
}

func (p *Pagination) GetOrderBy() string {
	if p.OrderBy == "" {
		p.OrderBy = "id"
	}

	return p.OrderBy
}

func (p *Pagination) SetOrderBy(sort string) string {
	if sort != "" {
		p.OrderBy = sort
	} else {
		p.OrderBy = "id"
	}

	return p.OrderBy
}

func (p *Pagination) GetSort() string {
	if p.Sort == "" {
		p.Sort = "asc"
	}

	return p.Sort
}

func (p *Pagination) SetSort(sort string) string {
	if strings.ToLower(sort) == "desc" {
		p.Sort = "desc"
	} else {
		p.Sort = "asc"
	}

	return p.Sort
}

func (p *Pagination) GetStringAttribute(name string) (string, error) {
	var val string
	var ok bool

	if p.Attributes[name] != nil {
		val, ok = p.Attributes[name].(string)
		if !ok {
			return "", fmt.Errorf("attribute %s is not a string", name)
		}
	}

	return val, nil
}

func (p *Pagination) GetIntAttribute(name string) (int, error) {
	var val int
	var ok bool

	if p.Attributes[name] != nil {
		val, ok = p.Attributes[name].(int)
		if !ok {
			return 0, fmt.Errorf("attribute %s is not an integer", name)
		}
	}

	return val, nil
}

func (p *Pagination) GetBoolAttribute(name string) (bool, error) {
	var val bool
	var ok bool

	if p.Attributes[name] != nil {
		val, ok = p.Attributes[name].(bool)
		if !ok {
			return false, fmt.Errorf("attribute %s is not a boolean", name)
		}
	}

	return val, nil
}

func (p *Pagination) GetFloatAttribute(name string) (float64, error) {
	var val float64
	var ok bool

	if p.Attributes[name] != nil {
		val, ok = p.Attributes[name].(float64)
		if !ok {
			return 0, fmt.Errorf("attribute %s is not a float", name)
		}
	}

	return val, nil
}

func (p *Pagination) GetAttributes() map[string]interface{} {
	return p.Attributes
}

func WithLimit(limit int) SearchOptionFunc {
	return func(so *SearchOption) {
		so.Limit = limit
	}
}

func WithPage(pageNumber int) SearchOptionFunc {
	return func(so *SearchOption) {
		so.Page = pageNumber
	}
}

func WithOrderBy(orderBy string) SearchOptionFunc {
	return func(so *SearchOption) {
		so.OrderBy = orderBy
	}
}

func WithSort(sort string) SearchOptionFunc {
	return func(so *SearchOption) {
		so.Sort = sort
	}
}

func WithAttributes(attributeName string, value any) SearchOptionFunc {
	return func(so *SearchOption) {
		so.Attributes[attributeName] = value
	}
}

func Paginate(value any, pagination *Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	totalPages := int(math.Ceil(float64(totalRows) / float64(pagination.Limit)))
	pagination.TotalPages = totalPages

	return func(db *gorm.DB) *gorm.DB {
		return db.
			Offset(pagination.GetOffset()).
			Limit(pagination.GetLimit()).
			Order(fmt.Sprintf(
				"%s %s",
				pagination.GetOrderBy(),
				pagination.GetSort(),
			))
	}
}
