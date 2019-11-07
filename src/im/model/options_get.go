package model

type GetOptions struct {
	// Filters the inactive users
	Inactive bool
	// Sorting option
	Sort Sort
	// Page
	Page int
	// Page size
	PerPage int
}

type Sort struct {
	Field string `json:"field"`
	Type  string `json:"type"`
}
