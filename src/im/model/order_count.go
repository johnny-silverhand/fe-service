package model

// Options for counting users
type OrderCountOptions struct {
	AppId string
	// Should include deleted users (of any type)
	IncludeDeleted bool
}
