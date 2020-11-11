package store

import "im/model"
import "context"

type LayeredStoreSupplierResult struct {
	StoreResult
}

func NewSupplierResult() *LayeredStoreSupplierResult {
	return &LayeredStoreSupplierResult{}
}

type LayeredStoreSupplier interface {
	//
	// Control
	//
	SetChainNext(LayeredStoreSupplier)
	Next() LayeredStoreSupplier

	// Roles
	RoleSave(ctx context.Context, role *model.Role, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	RoleGet(ctx context.Context, roleId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	RoleGetAll(ctx context.Context, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	RoleGetByName(ctx context.Context, name string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	RoleGetByNames(ctx context.Context, names []string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	RoleDelete(ctx context.Context, roldId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	RolePermanentDeleteAll(ctx context.Context, hints ...LayeredStoreHint) *LayeredStoreSupplierResult

	// Schemes
	SchemeSave(ctx context.Context, scheme *model.Scheme, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeGet(ctx context.Context, schemeId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeGetByName(ctx context.Context, schemeName string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeDelete(ctx context.Context, schemeId string, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemeGetAllPage(ctx context.Context, scope string, offset int, limit int, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	SchemePermanentDeleteAll(ctx context.Context, hints ...LayeredStoreHint) *LayeredStoreSupplierResult

	/*TeamMembersToAdd(ctx context.Context, since int64, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	ChannelMembersToAdd(ctx context.Context, since int64, hints ...LayeredStoreHint) *LayeredStoreSupplierResult

	TeamMembersToRemove(ctx context.Context, hints ...LayeredStoreHint) *LayeredStoreSupplierResult
	ChannelMembersToRemove(ctx context.Context, hints ...LayeredStoreHint) *LayeredStoreSupplierResult*/

}
