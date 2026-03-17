package validator

import (
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/validator/testdata"
)

// TestProtoValidator_Map_KeyPresence tests the `m:key` syntax for checking map key presence.
// Per AIP-160: "m:foo" checks if map m contains the key "foo"
func TestProtoValidator_Map_KeyPresence(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name    string
		filter  string
		wantErr bool
		errMsg  string
	}{
		// Valid key presence checks
		{"string map key presence", `labels:env`, false, ""},
		{"int32 value map key presence", `settings:timeout`, false, ""},
		{"bool value map key presence", `features:beta`, false, ""},
		{"int64 value map", `counters:requests`, false, ""},
		{"double value map", `metrics:cpu`, false, ""},

		// Numeric keys (int32/int64 maps)
		{"int32 key map - numeric key", `id_names:100`, false, ""},
		{"int64 key map - numeric key", `id_counts:999`, false, ""},

		// Invalid - not a map
		{"not a map field", `name:key`, true, "repeated field or map"},
		{"not a map - int field", `age:key`, true, "repeated field or map"},

		// Invalid - field doesn't exist
		{"nonexistent map", `nonexistent:key`, true, "does not exist"},

		// Invalid - missing HAS operator
		{"map without operator", `labels`, true, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error containing %q, got no errors", tt.errMsg)
				} else if tt.errMsg != "" && !contains(errs[0].Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", errs[0], tt.errMsg)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
			}
		})
	}
}

// TestProtoValidator_Map_KeyPresenceStar tests the `m.key:*` syntax.
// Per AIP-160: "m.foo:*" checks if map m contains key "foo" with any value
func TestProtoValidator_Map_KeyPresenceStar(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name    string
		filter  string
		wantErr bool
		errMsg  string
	}{
		// Valid key presence with star
		{"string map key:star", `labels.env:*`, false, ""},
		{"int32 value map key:star", `settings.timeout:*`, false, ""},
		{"bool value map key:star", `features.beta:*`, false, ""},
		{"double value map key:star", `metrics.cpu:*`, false, ""},

		// Invalid - star not in HAS operator context
		{"star in comparison not HAS", `labels.env = "*"`, false, ""}, // This is literal string "*", should be valid

		// Invalid - not a map
		{"not a map with star", `name.env:*`, true, "cannot traverse"},

		// Invalid - field doesn't exist
		{"nonexistent map with star", `nonexistent.key:*`, true, "does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error containing %q, got no errors", tt.errMsg)
				} else if tt.errMsg != "" && !contains(errs[0].Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", errs[0], tt.errMsg)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
			}
		})
	}
}

// TestProtoValidator_Map_KeyValueMatch tests the `m.key = value` syntax.
// Per AIP-160: "m.foo:42" or "m.foo = 42" checks if m["foo"] == 42
func TestProtoValidator_Map_KeyValueMatch(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name    string
		filter  string
		wantErr bool
		errMsg  string
	}{
		// Valid key-value matches
		{"string map string value", `labels.env = "production"`, false, ""},
		{"string map different key", `labels.region = "us-west"`, false, ""},
		{"int32 value map", `settings.timeout = 30`, false, ""},
		{"int64 value map", `counters.requests = 1000000`, false, ""},
		{"double value map", `metrics.cpu = 0.75`, false, ""},
		{"bool value map true", `features.beta = true`, false, ""},
		{"bool value map false", `features.enabled = false`, false, ""},

		// Numeric key maps with traversal syntax - PARSER LIMITATION
		// The parser doesn't support numeric identifiers after '.' (id_names.100)
		// Numeric keys work fine with HAS syntax: id_names:100 (see KeyPresence tests)
		// TODO: Update parser to support numeric field names in traversal
		// {"int32 key map with value", `id_names.100 = "user"`, false, ""},
		// {"int64 key map with value", `id_counts.999 = 42`, false, ""},

		// Type mismatches
		{"string map with int value", `labels.env = 123`, true, "type mismatch"},
		{"int32 map with string value", `settings.timeout = "30"`, true, "type mismatch"},
		{"bool map with string value", `features.beta = "true"`, true, "type mismatch"},
		{"double map with bool", `metrics.cpu = true`, true, "type mismatch"},

		// Invalid - not a map
		{"not a map traversal", `name.key = "value"`, true, ""},

		// Invalid - field doesn't exist
		{"nonexistent map key-value", `nonexistent.key = "val"`, true, "does not exist"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error containing %q, got no errors", tt.errMsg)
				} else if tt.errMsg != "" && !contains(errs[0].Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", errs[0], tt.errMsg)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
			}
		})
	}
}

// TestProtoValidator_Map_ComparisonOperators tests comparison operators on map values.
// Per AIP-160: Numeric map values support all comparison operators
func TestProtoValidator_Map_ComparisonOperators(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name    string
		filter  string
		wantErr bool
		errMsg  string
	}{
		// Valid numeric comparisons
		{"int32 map greater than", `settings.timeout > 20`, false, ""},
		{"int32 map less than", `settings.timeout < 100`, false, ""},
		{"int64 map greater equal", `counters.requests >= 1000`, false, ""},
		{"double map less than", `metrics.cpu < 1.0`, false, ""},
		{"double map not equal", `metrics.cpu != 0.5`, false, ""},

		// String comparisons (AIP-160 supports these)
		{"string map greater than", `labels.env > "prod"`, false, ""},
		{"string map less equal", `labels.region <= "z"`, false, ""},

		// Invalid - comparison on bool map
		{"bool map comparison", `features.beta > true`, true, "operator"},

		// Invalid - not a map
		{"comparison on non-map", `age > 25`, false, ""}, // This is valid for int field
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error containing %q, got no errors", tt.errMsg)
				} else if tt.errMsg != "" && !contains(errs[0].Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", errs[0], tt.errMsg)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
			}
		})
	}
}

// TestProtoValidator_Map_EdgeCases tests edge cases and invalid patterns.
func TestProtoValidator_Map_EdgeCases(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name    string
		filter  string
		wantErr bool
		errMsg  string
	}{
		// Valid numeric keys
		{"int32 key in HAS", `id_names:100`, false, ""},
		{"int64 key in HAS", `id_counts:999`, false, ""},

		// Invalid - can't traverse into map value (it's a scalar)
		{"traverse into string map value", `labels.env.nested = "val"`, true, "cannot traverse"},

		// Invalid - can't traverse into map value (it's a scalar)
		{"traverse into int32 map value", `settings.timeout.nested = 1`, true, "cannot traverse"},

		// HAS with star on whole map (ambiguous - might be invalid)
		{"star on map without key", `labels:*`, false, ""}, // Valid: map presence check per AIP-160

		// NOT operator with maps
		{"NOT with map key presence", `NOT labels:env`, false, ""},
		// Minus operator is parsed differently than NOT - appears to be parser issue
		// TODO: Investigate why `-labels:env` doesn't parse like `NOT labels:env`
		{"minus operator with map", `-labels:env`, true, "invalid collection"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got no errors")
				} else if tt.errMsg != "" && !contains(errs[0].Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", errs[0], tt.errMsg)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
			}
		})
	}
}

// TestProtoValidator_Map_MessageValues tests traversal into map values that are messages.
// Per Proto3: map values can be any type including messages (but not another map).
// Example: map<string, Address> locations allows: locations.home.city = "NYC"
func TestProtoValidator_Map_MessageValues(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name    string
		filter  string
		wantErr bool
		errMsg  string
	}{
		// Valid - traverse into message map value
		// e.g. locations["home"] = { "city": "NYC"}
		{"traverse into message field", `locations.home.city = "NYC"`, false, ""},
		{"traverse into message field zip", `locations.office.zip = "10001"`, false, ""},

		// Valid - comparison operators on message map values
		{"greater than on message field", `locations.home.zip > "10000"`, false, ""},

		// Valid - HAS operator on message map value field
		{"has on nested field", `locations.home.city:*`, false, ""},
		
		// Valid - bool field with = operator
		{"bool field equals", `locations.home.is_primary = true`, false, ""},
		{"bool field not equals", `locations.home.is_primary != false`, false, ""},
		
		// Invalid - bool field with comparison operators
		{"bool field greater than", `locations.home.is_primary > true`, true, "does not support operator"},
		{"bool field less than", `locations.home.is_primary < false`, true, "does not support operator"},

		// Invalid - field doesn't exist in message value
		{"nonexistent field in message", `locations.home.country = "US"`, true, "does not exist"},

		// Invalid - can't traverse beyond scalar in message
		{"traverse into scalar in message", `locations.home.city.nested = "x"`, true, "cannot traverse"},

		// Invalid - still can't traverse into scalar map values
		{"traverse into scalar map", `labels.env.nested = "x"`, true, "cannot traverse"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := validateProtoFilter(t, tt.filter, msgDesc)
			if tt.wantErr {
				if len(errs) == 0 {
					t.Errorf("expected error, got no errors")
				} else if tt.errMsg != "" && !contains(errs[0].Error(), tt.errMsg) {
					t.Errorf("error = %v, want substring %q", errs[0], tt.errMsg)
				}
			} else {
				if len(errs) > 0 {
					t.Errorf("unexpected errors: %v", errs)
				}
			}
		})
	}
}
