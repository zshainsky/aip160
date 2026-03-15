# ProtoValidator

A high-performance protobuf-aware filter validator that validates AIP-160 filter expressions against protobuf message descriptors using `google.golang.org/protobuf/reflect/protoreflect`.

## Overview

ProtoValidator validates filter expressions at the schema level, ensuring that:
- Fields exist in the proto message definition
- Values match field types (string, int32, enums, etc.)
- Operators are valid for field types (e.g., booleans only support `=` and `!=`)
- Nested field access is valid
- HAS operator usage is correct for repeated fields

**Performance:** 2-5x faster than reflection-based validation due to O(1) protobuf descriptor lookups.

## When to Use ProtoValidator

### ✅ Use ProtoValidator when:
- Your data model is defined in `.proto` files
- You have generated `*.pb.go` files
- You need proto-specific validation (enums, nested messages, repeated fields)
- Performance is important (high-throughput filtering)
- You want native enum value validation with prefix stripping

### ⚠️ Use reflection-based Validator when:
- Your data model uses regular Go structs (not proto)
- You don't want to generate proto files
- You need struct tag support (`json:"fieldName"`, `filter:"customName"`)

## How ProtoReflect Works

ProtoValidator leverages protobuf's native reflection system:

```go
// Regular reflection (Validator)
reflect.TypeOf(myStruct).FieldByName("UserAge")  // O(n) iteration

// Protobuf reflection (ProtoValidator)
descriptor.Fields().ByName("user_age")            // O(1) map lookup
```

**Key Benefits:**
1. **O(1) Field Lookups:** Descriptor provides hash-based field access
2. **Type Safety:** Proto types are strongly defined (int32 ≠ int64)
3. **Enum Support:** Native enum descriptors with value validation
4. **Nested Messages:** Message descriptors form a traversable tree
5. **Repeated Fields:** `IsList()` check for collection detection

## Usage

### Basic Usage

```go
package main

import (
    "fmt"
    "github.com/zshainsky/aip160/pkg/filter/parser"
    "github.com/zshainsky/aip160/pkg/filter/validator"
)

// Assuming you have a proto-generated message
func main() {
    // Get message descriptor from proto-generated type
    msgDesc := (&pb.User{}).ProtoReflect().Descriptor()
    
    // Create validator
    v := validator.NewProtoValidator(msgDesc)
    
    // Parse filter expression
    expr, err := parser.ParseFilter(`age > 25 AND status = "ACTIVE"`)
    if err != nil {
        panic(err)
    }
    
    // Validate
    errs := v.Validate(expr)
    if len(errs) > 0 {
        for _, e := range errs {
            fmt.Printf("Validation error: %v\n", e)
        }
        return
    }
    
    fmt.Println("Filter is valid!")
}
```

### With Options

```go
// Disable enum prefix stripping (require exact enum values)
v := validator.NewProtoValidator(msgDesc,
    validator.WithEnumPrefixStripping(false),
)
```

### Example Proto

```protobuf
syntax = "proto3";

message User {
  string name = 1;
  int32 age = 2;
  bool active = 3;
  UserStatus status = 4;
  Address address = 5;           // Nested message
  repeated string tags = 6;      // Repeated field
}

enum UserStatus {
  USER_STATUS_UNKNOWN = 0;
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_INACTIVE = 2;
}

message Address {
  string city = 1;
  string country = 2;
}
```

## Supported Features

### ✅ Field Existence Validation

```go
// Valid
age > 25
name = "John"

// Invalid
invalid_field = 1  // Error: field 'invalid_field' does not exist
```

### ✅ Type Compatibility

Supports all proto scalar types:
- **String:** string, bytes
- **Boolean:** bool
- **Integers:** int32, int64, uint32, uint64, sint32, sint64, fixed32, fixed64, sfixed32, sfixed64
- **Floats:** float, double
- **Enums:** Validated as string literals

```go
// Valid
age = 25               // int32 field
score = 3.14           // double field
active = true          // bool field

// Invalid
age = "twenty-five"    // Error: type mismatch (expects int, got string)
active = 1             // Error: type mismatch (expects bool, got int)
```

### ✅ Enum Validation with Prefix Stripping

ProtoValidator supports flexible enum matching:

```protobuf
enum UserStatus {
  USER_STATUS_UNKNOWN = 0;
  USER_STATUS_ACTIVE = 1;
  USER_STATUS_INACTIVE = 2;
}
```

```go
// All valid (with default prefix stripping)
status = "ACTIVE"              // ✓ Prefix stripped
status = "USER_STATUS_ACTIVE"  // ✓ Exact match

// Invalid
status = "INVALID"             // ✗ Not a valid enum value
```

**Disable prefix stripping:**
```go
v := validator.NewProtoValidator(msgDesc, 
    validator.WithEnumPrefixStripping(false))
// Now only "USER_STATUS_ACTIVE" is valid
```

### ✅ Operator Restrictions

Different field types support different operators:

| Field Type | Supported Operators |
|-----------|-------------------|
| Numeric (int32, int64, float, double) | `=`, `!=`, `<`, `>`, `<=`, `>=` |
| String, bytes | `=`, `!=`, `<`, `>`, `<=`, `>=` |
| Boolean | `=`, `!=` only |
| Enum | `=`, `!=` only |
| Repeated | `:` (HAS) only |

```go
// Valid
age > 25           // Numeric comparison
active = true      // Boolean equality
status != "INACTIVE"  // Enum comparison

// Invalid
active > true      // Error: boolean fields only support = and !=
status < "ACTIVE"  // Error: enum fields only support = and !=
tags = "urgent"    // Error: repeated fields only support HAS operator
```

### ✅ Nested Field Traversal

```go
// Valid (unlimited depth)
address.city = "Seattle"
address.country = "USA"
user.profile.settings.theme = "dark"

// Invalid
name.invalid = 1   // Error: cannot traverse into scalar field 'name'
```

### ✅ HAS Operator (`:`) for Repeated Fields

AIP-160 requires the HAS operator for collection membership:

```protobuf
message User {
  repeated string tags = 1;
  repeated int32 scores = 2;
  repeated Address addresses = 3;
}
```

```go
// Valid HAS usage
tags:"urgent"                    // String in repeated field
scores:100                       // Int in repeated field
addresses.city:"Seattle"         // Nested field in repeated message

// Invalid
tags = "urgent"                  // Error: use HAS (:) for repeated fields
addresses = something            // Error: comparison operators not allowed on repeated
```

### ✅ Logical Operators

```go
// Valid
age > 25 AND active = true
status = "ACTIVE" OR status = "PENDING"
NOT (age < 18)
tags:"urgent" AND scores:100     // HAS with logical operators
```

## Validation Errors

ProtoValidator provides clear error messages:

```go
field 'invalid_field' does not exist in message User
type mismatch: field 'age' is int32, cannot compare with string value
operator '<' not allowed for boolean field 'active'
enum field 'status' has invalid value "INVALID_STATUS". Valid values: ACTIVE, INACTIVE, PENDING
cannot traverse into non-message field 'name' (type: string)
comparison operators not allowed on repeated field 'tags', use HAS operator (:) instead
```

## Limitations

### ❌ Map Fields (Planned for v2)

Map fields are not yet supported. Planned syntax per AIP-160:
```go
labels:key           // Check if key exists
labels.key:*         // Check if key is present
labels.key:value     // Check key-value pair
```

### ❌ Star Operator (Parser Limitation)

The star operator for presence checks requires parser updates:
```go
tags:*        // Should work per AIP-160, but parser doesn't support it yet
message:*     // Check if message field is present
```

Workaround: Use nested field checks instead.

### ℹ️ Function Calls

Function call validation is not implemented. The parser supports function syntax, but validation is deferred until a specific use case emerges (YAGNI principle).

## Performance Characteristics

| Operation | ProtoValidator | Reflection Validator |
|-----------|---------------|-------------------|
| Field lookup | O(1) hash map | O(n) struct iteration |
| Type check | O(1) descriptor | O(1) reflect.Kind |
| Enum validation | O(1) descriptor | O(n) tag parsing |
| Nested traversal | O(depth) | O(depth × n) |

**Benchmark Results:**
- 2-5x faster for simple filters
- 5-10x faster for filters with enums
- Negligible overhead for nested traversal

## Architecture

```
ProtoValidator
├── Field Resolution
│   └── O(1) descriptor.Fields().ByName()
├── Type Checking
│   └── protoreflect.Kind comparison
├── Enum Validation
│   ├── Descriptor provides valid values
│   └── Optional prefix stripping
├── Operator Validation
│   └── Kind-based restrictions
├── HAS Operator
│   ├── IsList() detection
│   ├── Element type validation
│   └── Nested traversal support
└── Traversal
    └── Recursive descriptor navigation
```

## Testing

ProtoValidator is thoroughly tested with 217+ test cases covering:
- All proto scalar types (17 types)
- Enum validation (prefixed/non-prefixed)
- Nested messages (3+ levels deep)
- Operator restrictions
- HAS operator (scalars, enums, messages)
- Logical operator combinations
- Error message accuracy

Test coverage: **83.1%** of statements.

## Examples

See `proto_validator_test.go` for comprehensive examples of:
- Basic validation
- Enum handling
- Nested traversal
- HAS operator usage
- Error scenarios

## Contributing

When adding features:
1. Follow strict TDD (RED → GREEN → REFACTOR)
2. Maintain >80% test coverage
3. Update this documentation
4. Add godoc comments
5. Include examples in tests

## References

- [AIP-160: Filtering](https://google.aip.dev/160) - Google API Improvement Proposals
- [protoreflect package](https://pkg.go.dev/google.golang.org/protobuf/reflect/protoreflect) - Go protobuf reflection
- [Protocol Buffers](https://protobuf.dev/) - Official protobuf documentation
