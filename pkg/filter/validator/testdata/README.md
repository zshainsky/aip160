# Validator Test Data

This directory contains protobuf definitions used for testing the ProtoValidator implementation.

## Files

- **`testdata.proto`** - Protocol buffer definitions with comprehensive test cases
- **`testdata.pb.go`** - Generated Go code (do not edit manually)

## What's Tested

The `TestProtoData` message includes fields for testing:

### Basic Types
- All scalar types (string, int32, int64, uint32, uint64, bool, float, double, bytes, etc.)
- Enums with different prefix patterns
- Nested messages (3 levels deep)

### Advanced Features
- **Repeated fields** - Testing HAS operator (`:`)
- **Duration fields** - Testing AIP-160 duration literals (`30s`, `1.5h`)
- **Timestamp fields** - Testing RFC-3339 timestamps
- **Map fields** - Testing all AIP-160 map syntaxes:
  - Key presence: `labels:env`
  - Key-value match: `labels.env = "prod"`
  - Map traversal: `locations.home.city = "NYC"`

### Map Types Tested
- `map<string, string>` - String keys and values
- `map<string, int32/int64/double/bool>` - String keys with numeric/bool values
- `map<int32/int64, string>` - Numeric keys
- `map<string, Message>` - Map with message values (supports traversal)

## Regenerating Go Code

After modifying `testdata.proto`, regenerate the Go code:

### Prerequisites

```bash
# Install protoc compiler
# macOS: brew install protobuf
# Linux: apt-get install protobuf-compiler

# Install Go protobuf plugin
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
```

### Generate Code

From the testdata directory:

```bash
cd pkg/filter/validator/testdata

protoc --go_out=. --go_opt=paths=source_relative \
    --go_opt=Mtestdata.proto=github.com/zshainsky/aip160/pkg/filter/validator/testdata \
    testdata.proto
```

Or from the repository root:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
    --go_opt=Mpkg/filter/validator/testdata/testdata.proto=github.com/zshainsky/aip160/pkg/filter/validator/testdata \
    pkg/filter/validator/testdata/testdata.proto
```

### Verify Generation

After regenerating, run tests to ensure everything still works:

```bash
cd ../..  # Back to validator directory
go test ./...
```

## Modifying Test Data

When adding new test cases:

1. **Add fields to `testdata.proto`** following Proto3 syntax
2. **Add descriptive comments** explaining what's being tested
3. **Regenerate** `testdata.pb.go` using the commands above
4. **Update tests** in `proto_validator_test.go` to cover new fields
5. **Run tests** to verify everything works

### Example: Adding a New Field

```protobuf
message TestProtoData {
  // ... existing fields ...
  
  // New field for testing custom validation
  string username = 99;  // Use unused field number
}
```

Then regenerate and test:

```bash
protoc --go_out=. --go_opt=paths=source_relative testdata.proto
cd .. && go test -v
```

## Common Issues

### Import Errors

If you get import errors for `google/protobuf/duration.proto` or `google/protobuf/timestamp.proto`:

```bash
# These are included with protoc installation
# Verify protoc can find them:
protoc --version
protoc --proto_path=/usr/local/include --go_out=. testdata.proto
```

### Module Path Issues

The `go_package` option must match your module path:

```protobuf
option go_package = "github.com/zshainsky/aip160/pkg/filter/validator/testdata";
```

Update this if you fork the repository.

## Test Coverage

The testdata supports comprehensive ProtoValidator testing:

- ✅ Field existence validation
- ✅ Type compatibility checking
- ✅ Enum value validation (with prefix stripping)
- ✅ Nested message traversal (unlimited depth)
- ✅ HAS operator on repeated fields
- ✅ Duration literal parsing (`30s`, `1.5h`)
- ✅ Timestamp RFC-3339 parsing
- ✅ Map key presence checks
- ✅ Map value comparisons
- ✅ Map message value traversal
- ✅ Operator restrictions (bool, enum)

All test cases follow AIP-160 specification requirements.
