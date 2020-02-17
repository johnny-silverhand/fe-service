package model

type UserGetOptions struct {
	// Filters the users in the team
	InTeamId string
	// Filters the users not in the team
	NotInTeamId string
	// Filters the users in the channel
	InChannelId string
	// Filters the users not in the channel
	NotInChannelId string
	// Filters the users without a team
	WithoutTeam bool
	// Filters the inactive users
	Inactive bool
	// Filters for the given role
	Role string
	// Sorting option
	Sort string
	// Page
	Page int
	// Page size
	PerPage int
	// application id
	AppId string
	// email
	Email string
}
