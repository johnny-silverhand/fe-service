package model

type ProductGetOptions struct {
	// Sorting option
	Sort string
	// Page
	Page int
	// Page size
	PerPage int
	// Filter the products by application id
	AppId string
	// Filter the products by category id
	CategoryId string
	// Filter the products by office id
	OfficeId string
	// Filter the products by status
	Status string
	// Filter the products by active
	Active *bool
}
