// Package util provides utility function.
package util //nolint:revive

import (
	"encoding/json"
	"fmt"

	"github.com/stackrox/stackrox-mcp/internal/logging"
)

// MustMarshal returns marshaled value or exits on fail.
func MustMarshal(x any) json.RawMessage {
	data, err := json.Marshal(x)
	if err != nil {
		logging.Fatal(fmt.Sprintf("Error marshalling json: %v", x), err)
	}

	return json.RawMessage(data)
}
