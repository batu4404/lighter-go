// Package apimodels contains Go types generated from the Lighter OpenAPI
// spec (../../../lighter-python/openapi.json), giving field-faithful parity
// with the Python SDK's models. Endpoint methods are hand-wired in
// client/http (so the signer-coupled SendTx path is preserved); only the
// models are generated here.
//
// To regenerate after an openapi.json bump:
//
//	go generate ./types/apimodels
//
//go:generate -command oapi go run github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.7.1
//go:generate python3 ../../tools/patch_openapi_for_go.py ../../tools/_openapi_go.json
//go:generate oapi -config ../../tools/oapi-models.yaml ../../tools/_openapi_go.json
package apimodels
