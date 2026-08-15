package contracts

import _ "embed"

//go:embed openapi/openapi.json
var canonicalOpenAPI []byte

// CanonicalOpenAPI returns a copy of the canonical OpenAPI document.
func CanonicalOpenAPI() []byte {
	return append([]byte(nil), canonicalOpenAPI...)
}
