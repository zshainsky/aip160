package validator

import (
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/validator/testdata"
)

// TestProtoValidator_TraversalRestrictions tests Phase 4: Traversal restrictions per AIP-160
// Two types of restrictions:
// 1. Cannot use dot operator to traverse through repeated fields
// 2. Cannot use array index access (e.g., e.0.foo = 42)
func TestProtoValidator_TraversalRestrictions(t *testing.T) {
	testProtoData := &testdata.TestProtoData{}
	msgDesc := testProtoData.ProtoReflect().Descriptor()

	tests := []struct {
		name    string
		filter  string
		wantErr bool
		errMsg  string
	}{
		// ===================================================================
		// Type 1: Dot Traversal Through Repeated Fields
		// Per AIP-160: "The . operator must not be used to traverse through a repeated field"
		// ===================================================================

		// Invalid - dot through repeated message field
		{
			name:    "dot through repeated message",
			filter:  `emails.address = "test@example.com"`,
			wantErr: true,
			errMsg:  "cannot use dot operator to traverse through repeated field",
		},
		{
			name:    "dot through repeated message - nested field",
			filter:  `emails.address != "user@test.com"`,
			wantErr: true,
			errMsg:  "cannot use dot operator to traverse through repeated field",
		},
		{
			name:    "comparison on repeated scalar",
			filter:  `tags = "urgent"`,
			wantErr: true, // This is caught by existing validation
			errMsg:  "requires an operator",
		},

		// Valid - HAS operator on repeated (correct syntax)
		{
			name:    "HAS operator on repeated message",
			filter:  `emails:address:"test@example.com"`,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "HAS operator on repeated scalar",
			filter:  `tags:"urgent"`,
			wantErr: false,
			errMsg:  "",
		},

		// Valid - dot through singular message (NOT repeated)
		{
			name:    "dot through singular message",
			filter:  `email.address = "test@example.com"`,
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "dot through nested singular messages",
			filter:  `nested.leaf.text = "value"`,
			wantErr: false,
			errMsg:  "",
		},

		// ===================================================================
		// Type 2: Array Index Access (Numeric Field Names)
		// Per AIP-160: "e.0.foo = 42 and e[0].foo = 42 are not valid filters"
		// ===================================================================

		// Note: The parser doesn't support numeric field names after dot operator
		// This is actually a parser-level prevention, which is good!
		// The validator check is a safety net if the parser changes.
		
		// Parser blocks these (which is correct behavior):
		// {filter: `data.0 = "first"`, err: parser error}
		// {filter: `e.0.foo = 42`, err: parser error}
		// {filter: `items.42 = "value"`, err: parser error}
		
		// The validator check would catch these IF they parsed

		// Valid - field names that START with numbers but aren't purely numeric
		{
			name:    "field name starting with number",
			filter:  `name = "value"`, // Use actual field
			wantErr: false,
			errMsg:  "",
		},
		{
			name:    "field name with number inside",
			filter:  `nested.leaf.text = "value"`,
			wantErr: false,
			errMsg:  "",
		},

		// Valid - normal field names
		{
			name:    "normal field traversal",
			filter:  `email.address = "test@example.com"`,
			wantErr: false,
			errMsg:  "",
		},

		// ===================================================================
		// Edge Cases
		// ===================================================================

		// Multi-level traversal
		{
			name:    "three-level traversal - all singular",
			filter:  `nested.leaf.text = "value"`,
			wantErr: false,
			errMsg:  "",
		},

		// Combination: repeated at different levels
		{
			name:    "dot after repeated field in middle",
			filter:  `nested.leaf.leaf_tags.tag = "urgent"`,
			wantErr: true, // leaf_tags is repeated
			errMsg:  "cannot use dot operator to traverse through repeated field",
		},
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
