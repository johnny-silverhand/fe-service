package model

type PromoGetOptions struct {
	AllowFromCache bool
	// Sorting option
	Sort string
	// Page
	Page int
	// Page size
	PerPage int
	// Filter the promos by application id
	AppId string
	// Filter the promos by category id
	CategoryId string
	// Filter the promos by office id
	OfficeId string
	// Filter the promos by status
	Status string
	// Filter the promo by active
	Active *bool
	// Use filter to mobile users
	Mobile bool
}
