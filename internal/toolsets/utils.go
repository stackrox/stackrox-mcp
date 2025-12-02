package toolsets

import (
	"encoding/json"
	"fmt"

	"github.com/stackrox/stackrox-mcp/internal/logging"
)

// MustJSONMarshal marshals value into a raw encoded JSON value or crashes.
func MustJSONMarshal(value any) json.RawMessage {
	marshaledValue, err := json.Marshal(value)
	if err != nil {
		logging.Fatal(fmt.Sprintf("marshaling failed for value: %v", value), err)
	}

	return marshaledValue
}
