package model

type RolePermission struct {
	RoleID       string `db:"role_id" json:"role_id"`
	PermissionID string `db:"permission_id" json:"permission_id"`
}

type CreateRolePermissionRequest struct {
	RoleID       string `json:"role_id"`
	PermissionID string `json:"permission_id"`
}

type UpdateRolePermissionRequest struct {
	NewRoleID       string `json:"new_role_id"`
	NewPermissionID string `json:"new_permission_id"`
}
