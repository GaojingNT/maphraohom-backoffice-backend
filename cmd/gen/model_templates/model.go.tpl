package models

type {{.PascalModuleName}} struct {
	BaseModel
}

// {{.PascalModuleName}} searchable attributes
func {{.PascalModuleName}}Searchable() []string {
	return []string{}
}
