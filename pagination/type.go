package pagination

// Pagination type schema
type Pagination struct {
	Limit     int `form:"limit" cache:"optional"`
	Page      int `form:"page" cache:"optional"`
	Offset    int
	TotalPage int // Total number of pages
	TotalData int // Total of data in database
}
