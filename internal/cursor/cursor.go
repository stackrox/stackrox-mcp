// Package cursor implements logic for paging.
package cursor

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
)

// Cursor represents pagination state with offset for next offset.
//
// We want to follow a pattern defined in MCP specification:
// https://modelcontextprotocol.io/specification/2025-11-25/server/utilities/pagination
type Cursor struct {
	Offset int32 `json:"offset"`
}

// New creates and validates a new Cursor.
func New(offset int32) (*Cursor, error) {
	cursor := &Cursor{
		Offset: offset,
	}

	if err := cursor.validate(); err != nil {
		return nil, err
	}

	return cursor, nil
}

// Encode serializes the cursor to a Base64-encoded string.
func (c *Cursor) Encode() (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(c)
	if err != nil {
		return "", fmt.Errorf("failed to marshal cursor: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonBytes)

	return encoded, nil
}

// Decode deserializes a Base64-encoded string to a Cursor.
func Decode(encoded string) (*Cursor, error) {
	if encoded == "" {
		return nil, errors.New("encoded cursor cannot be empty")
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("invalid base64 encoding: %w", err)
	}

	var cursor Cursor
	if err := json.Unmarshal(decoded, &cursor); err != nil {
		return nil, fmt.Errorf("invalid cursor format: %w", err)
	}

	if err := cursor.validate(); err != nil {
		return nil, err
	}

	return &cursor, nil
}

// GetOffset returns offset that can be used for API call.
func (c *Cursor) GetOffset() int32 {
	return c.Offset
}

// GetNextCursor returns cursor for the next offset.
func (c *Cursor) GetNextCursor(limit int32) *Cursor {
	if limit < 0 || c.Offset+limit < 0 {
		limit = 0
	}

	return &Cursor{
		Offset: c.Offset + limit,
	}
}

// validate checks if the cursor has valid values.
func (c *Cursor) validate() error {
	if c.Offset < 0 {
		return errors.New("offset must be non-negative")
	}

	return nil
}
