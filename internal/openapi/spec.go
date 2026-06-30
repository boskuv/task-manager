package openapi

import _ "embed"

// Spec is the OpenAPI 3.0 description of the Task Manager API.
//
//go:embed openapi.yaml
var Spec []byte
