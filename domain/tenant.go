package domain

const DefaultTenantID = "default"

// NormalizeTenantID preserves legacy records by assigning them to the stable
// default tenant. Store implementations must call it before persistence and
// must require an explicit tenant for tenant-aware lookup operations.
func NormalizeTenantID(tenantID string) string {
	if tenantID == "" {
		return DefaultTenantID
	}
	return tenantID
}
