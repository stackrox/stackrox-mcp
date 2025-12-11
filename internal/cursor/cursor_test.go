package cursor

import (
	"encoding/base64"
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew_ValidInput(t *testing.T) {
	tests := map[string]struct {
		offset int32
	}{
		"offset 0": {
			offset: 0,
		},
		"offset 5": {
			offset: 5,
		},
		"large values": {
			offset: 1000000,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			cursor, err := New(testCase.offset)
			require.NoError(t, err)
			require.NotNil(t, cursor)
			assert.Equal(t, testCase.offset, cursor.Offset)
		})
	}
}

func TestNew_InvalidInput(t *testing.T) {
	cursor, err := New(-1)
	require.Error(t, err)
	assert.Nil(t, cursor)
	assert.Contains(t, err.Error(), "offset must be non-negative")
}

func TestDecode_Success(t *testing.T) {
	original := &Cursor{Offset: 1}
	encoded, err := original.Encode()
	require.NoError(t, err)

	decoded, err := Decode(encoded)
	require.NoError(t, err)
	require.NotNil(t, decoded)
	assert.Equal(t, original.Offset, decoded.Offset)
}

func TestDecode_InvalidInput(t *testing.T) {
	tests := map[string]struct {
		encoded       string
		expectedError string
	}{
		"empty string": {
			encoded:       "",
			expectedError: "encoded cursor cannot be empty",
		},
		"invalid base64": {
			encoded:       "not-base64!@#$%",
			expectedError: "invalid base64 encoding",
		},
		"invalid json": {
			encoded:       base64.StdEncoding.EncodeToString([]byte("not json")),
			expectedError: "invalid cursor format",
		},
		"valid json but invalid cursor - negative offset": {
			encoded:       base64.StdEncoding.EncodeToString([]byte(`{"offset":-1}`)),
			expectedError: "offset must be non-negative",
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			decoded, err := Decode(testCase.encoded)
			require.Error(t, err)
			assert.Nil(t, decoded)
			assert.Contains(t, err.Error(), testCase.expectedError)
		})
	}
}

func TestEncodeDecode_RoundTrip(t *testing.T) {
	tests := map[string]struct {
		offset int32
	}{
		"zero offset": {
			offset: 0,
		},
		"non-zero offset": {
			offset: 5,
		},
		"large offset": {
			offset: 10000,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			original, err := New(testCase.offset)
			require.NoError(t, err)

			encoded, err := original.Encode()
			require.NoError(t, err)
			assert.NotEmpty(t, encoded)

			decoded, err := Decode(encoded)
			require.NoError(t, err)
			require.NotNil(t, decoded)

			assert.Equal(t, original.Offset, decoded.Offset)
		})
	}
}

func TestEncode_InvalidInput(t *testing.T) {
	cursor := &Cursor{Offset: -1}
	encoded, err := cursor.Encode()

	require.Error(t, err)
	assert.Empty(t, encoded)
	assert.Contains(t, err.Error(), "offset must be non-negative")
}

func TestGetOffset(t *testing.T) {
	cursor := &Cursor{Offset: 1}
	assert.Equal(t, cursor.Offset, cursor.GetOffset())
}

func TestGetNextCursor(t *testing.T) {
	cursor := &Cursor{Offset: 0}

	cursorStep1 := cursor.GetNextCursor(10)
	assert.Equal(t, int32(10), cursorStep1.GetOffset())

	cursorStep2 := cursorStep1.GetNextCursor(5)
	assert.Equal(t, int32(15), cursorStep2.GetOffset())

	cursorNegativeLimit := cursorStep2.GetNextCursor(-1)
	assert.Equal(t, int32(15), cursorNegativeLimit.GetOffset(), "negative limit should not change offset")

	cursorOverflow := cursorStep2.GetNextCursor(math.MaxInt32)
	assert.Equal(t, int32(15), cursorOverflow.GetOffset(), "overflow paging should not change offset")
}
