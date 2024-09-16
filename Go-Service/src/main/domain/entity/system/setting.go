package system
type Setting struct {
	EditorRoleId string `json:"editor_role_id"`
	StreamAccessRoleIds []string `json:"stream_access_role_ids"`
}