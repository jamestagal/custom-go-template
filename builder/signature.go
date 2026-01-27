// Package builder provides utilities for building Plenti-compatible output.
// This file re-exports signature utilities from the types package for backward compatibility.
package builder

import (
	"github.com/jimafisk/custom_go_template/types"
)

// SignatureInfo is re-exported from types for backward compatibility.
// Use types.SignatureInfo directly for new code.
type SignatureInfo = types.SignatureInfo

// GenerateSignature converts a layout file path to a Plenti-compatible signature.
// Re-exported from types package for backward compatibility.
//
// See types.GenerateSignature for full documentation.
func GenerateSignature(filePath string) string {
	return types.GenerateSignature(filePath)
}

// ParseSignature extracts components from a Plenti-compatible signature.
// Re-exported from types package for backward compatibility.
//
// See types.ParseSignature for full documentation.
func ParseSignature(signature string) SignatureInfo {
	return types.ParseSignature(signature)
}

// ShortNameToSignature converts a short component name to a full Plenti signature.
// Re-exported from types package for backward compatibility.
//
// See types.ShortNameToSignature for full documentation.
func ShortNameToSignature(name string, extension string) string {
	return types.ShortNameToSignature(name, extension)
}

// SignatureToShortName extracts just the component name from a full signature.
// Re-exported from types package for backward compatibility.
//
// See types.SignatureToShortName for full documentation.
func SignatureToShortName(signature string) string {
	return types.SignatureToShortName(signature)
}
