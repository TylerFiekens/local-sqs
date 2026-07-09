// Package api embeds the OpenAPI specification for local-sqs.
package api

import _ "embed"

// Spec is the OpenAPI 3.0 specification in YAML format.
//
//go:embed openapi.yaml
var Spec []byte
