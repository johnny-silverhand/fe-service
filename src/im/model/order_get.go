package model

type OrderGetOptions struct {
	IncludeStatuses string
	ExcludeStatuses string

	Status string
	// Sorting option
	Sort string
	// Page
	Page int
	// Page size
	PerPage int
	// application id
	AppId string
	// user id
	UserId string
}
