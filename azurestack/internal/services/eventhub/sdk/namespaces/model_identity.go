package namespaces

type Identity struct {
	PrincipalId *string       `json:"principalId,omitempty"`
	TenantId    *string       `json:"tenantId,omitempty"`
	Type        *IdentityType `json:"type,omitempty"`
}
