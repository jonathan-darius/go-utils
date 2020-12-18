package pagination

// Pagination type schema
type Pagination struct {
	Limit     int `form:"limit" cache:"optional"`
	Page      int `form:"page" cache:"optional"`
	Offset    int
	TotalPage int // Total number of pages
	Total     int // Total of data in database
}
