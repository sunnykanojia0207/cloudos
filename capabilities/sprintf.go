// Package capabilities provides capability interfaces.
//
// This file exists solely to avoid importing "fmt" in capability.go for a single
// Sprintf call, keeping the interface package as lightweight as possible.
package capabilities

import "fmt"

// sprintf is an alias for fmt.Sprintf used internally to avoid direct fmt import
// in the main capability file.
func sprintf(format string, args ...interface{}) string {
	return fmt.Sprintf(format, args...)
}
