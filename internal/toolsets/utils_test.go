package toolsets

import (
	"errors"
	"os"
	"os/exec"
	"testing"

	"github.com/google/jsonschema-go/jsonschema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMustMarshal_SimpleTypes tests marshaling of primitive types.
func TestMustMarshal_SimpleTypes(t *testing.T) {
	tests := map[string]struct {
		input    any
		expected string
	}{
		"integer":        {42, "42"},
		"negative int":   {-100, "-100"},
		"zero":           {0, "0"},
		"string":         {"hello", `"hello"`},
		"empty string":   {"", `""`},
		"boolean true":   {true, "true"},
		"boolean false":  {false, "false"},
		"float":          {3.14, "3.14"},
		"negative float": {-2.5, "-2.5"},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			result := MustJSONMarshal(testCase.input)

			require.NotNil(t, result)
			assert.JSONEq(t, testCase.expected, string(result))
		})
	}
}

// TestMustMarshal_Structs tests marshaling of structs with JSON tags.
func TestMustMarshal_Structs(t *testing.T) {
	type SimpleStruct struct {
		Name  string `json:"name"`
		Value int    `json:"value"`
	}

	type NestedStruct struct {
		ID     string       `json:"id"`
		Simple SimpleStruct `json:"simple"`
	}

	tests := map[string]struct {
		input    any
		expected string
	}{
		"simple struct": {
			SimpleStruct{Name: "test", Value: 123},
			`{"name":"test","value":123}`,
		},
		"struct with empty values": {
			SimpleStruct{},
			`{"name":"","value":0}`,
		},
		"nested struct": {
			NestedStruct{
				ID:     "nested-1",
				Simple: SimpleStruct{Name: "inner", Value: 456},
			},
			`{"id":"nested-1","simple":{"name":"inner","value":456}}`,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			result := MustJSONMarshal(testCase.input)

			require.NotNil(t, result)
			assert.JSONEq(t, testCase.expected, string(result))
		})
	}
}

// TestMustMarshal_Collections tests marshaling of slices, arrays, and maps.
func TestMustMarshal_Collections(t *testing.T) {
	tests := map[string]struct {
		input    any
		expected string
	}{
		"int slice": {
			[]int{1, 2, 3},
			`[1,2,3]`,
		},
		"empty slice": {
			[]string{},
			`[]`,
		},
		"string array": {
			[3]string{"a", "b", "c"},
			`["a","b","c"]`,
		},
		"map string to int": {
			map[string]int{"one": 1, "two": 2},
			`{"one":1,"two":2}`,
		},
		"empty map": {
			map[string]string{},
			`{}`,
		},
		"nested slice": {
			[][]int{{1, 2}, {3, 4}},
			`[[1,2],[3,4]]`,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			result := MustJSONMarshal(testCase.input)

			require.NotNil(t, result)
			assert.JSONEq(t, testCase.expected, string(result))
		})
	}
}

// TestMustMarshal_SpecialValues tests marshaling of nil, zero values, and edge cases.
func TestMustMarshal_SpecialValues(t *testing.T) {
	tests := map[string]struct {
		input    any
		expected string
	}{
		"nil slice": {
			([]int)(nil),
			`null`,
		},
		"nil map": {
			(map[string]int)(nil),
			`null`,
		},
		"nil pointer": {
			(*string)(nil),
			`null`,
		},
		"pointer to string": {
			jsonschema.Ptr("test"),
			`"test"`,
		},
		"pointer to int": {
			jsonschema.Ptr(42),
			`42`,
		},
	}

	for testName, testCase := range tests {
		t.Run(testName, func(t *testing.T) {
			result := MustJSONMarshal(testCase.input)

			require.NotNil(t, result)
			assert.JSONEq(t, testCase.expected, string(result))
		})
	}
}

func TestMustMarshal_Failure(t *testing.T) {
	if os.Getenv("CRASH_MustMarshal") == "true" {
		MustJSONMarshal(func() {})

		return
	}

	// Run the test in a subprocess.
	cmd := exec.CommandContext(t.Context(), "go", "test", "-test.run=TestMustMarshal_Failure")

	cmd.Env = append(os.Environ(), "CRASH_MustMarshal=true")
	err := cmd.Run()
	require.Error(t, err)

	exitState := &exec.ExitError{}
	correctType := errors.As(err, &exitState)
	require.True(t, correctType)
	assert.Equal(t, 1, exitState.ExitCode())
}
