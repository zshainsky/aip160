# Module 6: Schema Validation

## 🎯 Learning Objectives

By completing this module, you will:

1. ✅ Understand how to use Go's reflection to inspect struct types
2. ✅ Implement an AST visitor pattern for validation
3. ✅ Validate field existence and nested traversal
4. ✅ Check type compatibility for operators
5. ✅ Handle array/slice fields with the has operator
6. ✅ Validate function calls against a supported function registry
7. ✅ Collect and return multiple validation errors

## 📖 What is Schema Validation?

**Schema validation** checks if a parsed filter is **valid for a specific struct type** before execution. This catches errors early with clear messages instead of runtime failures.

This module implements **schematic validation** as specified in [AIP-160](https://google.aip.dev/160#validation), ensuring filters comply with the Google API Improvement Proposal standards.

### Example

Given this struct:
```go
type User struct {
    Name   string
    Age    int32
    Active bool
}
```

**Valid filters:**
```
name = "John"              ✅ Field exists, type compatible
age > 25                   ✅ Field exists, numeric comparison
active = true              ✅ Field exists, boolean value
```

**Invalid filters:**
```
email = "test@example.com" ❌ Field 'email' doesn't exist
name > 100                 ❌ Type mismatch: string vs int
age = "hello"              ❌ Type mismatch: int vs string
```

### Why Validate?

Without validation:
```go
filter := parser.Parse("email = 'test@example.com'") // Field doesn't exist
sql := translator.ToSQL(filter)  // Generates: SELECT * FROM users WHERE email = 'test@example.com'
db.Query(sql)  // ❌ SQL error: column "email" does not exist
```

With validation:
```go
filter := parser.Parse("email = 'test@example.com'")
validator := NewValidator(reflect.TypeOf(User{}))
errs := validator.Validate(filter)  // ✅ Returns clear error immediately
// Error: "field 'email' does not exist in User"
```

## 🏗️ Architecture Overview

### Components

```
┌─────────────────┐
│  Filter String  │
└────────┬────────┘
         │ Parse (Modules 1-5)
         ▼
┌─────────────────┐     ┌──────────────┐
│   AST (Tree)    │────▶│  Validator   │
└─────────────────┘     └──────┬───────┘
                               │
                        ┌──────▼────────┐
                        │ Reflection on │
                        │ Go Struct     │
                        └──────┬────────┘
                               │
                        ┌──────▼────────┐
                        │ Validation    │
                        │ Errors []     │
                        └───────────────┘
```AIP-160 Compliance

This validator implements all required schematic validations from [AIP-160](https://google.aip.dev/160#validation):

✅ **Field existence** - Fields referenced must exist in the schema  
✅ **Type alignment** - Field values must match field types (e.g., `age = "hello"` is invalid for int32 field)  
✅ **Comparison operators** - Enforces operator restrictions by type:
   - **Strings**: All operators (`=`, `!=`, `<`, `>`, `<=`, `>=`)
   - **Numeric**: All operators  
   - **Timestamps**: All operators
   - **Durations**: All operators
   - **Booleans**: Only `=` and `!=` (NOT `<`, `>`, `<=`, `>=`)
   - **Enums**: Only `=` and `!=` (NOT comparison operators)

✅ **Array operators** - Enforces has operator (`:`) for array/slice fields  
✅ **Function validation** - Validates function calls and argument types  

### Validation Rules

| Check | Example | Error |
|-------|---------|-------|
| **Field Existence** | `email = "test"` on User (no email field) | Field 'email' not found |
| **Nested Traversal** | `profile.bio = "hi"` on User (no profile field) | Field 'profile' not found |
| **Type Compatibility** | `age > "hello"` (int field, string literal) | Type mismatch: int vs string |
| **Boolean Operators** | `active > true` | Booleans only support = and !=
| **Nested Traversal** | `profile.bio = "hi"` on User (no profile field) | Field 'profile' not found |
| **Type Compatibility** | `age > "hello"` (int field, string literal) | Type mismatch: int vs string |
| **Array Operator** | `tags = "urgent"` (should use `:`) | Use has operator `:` for array fields |
| **Function Validation** | `unknown(x)` | Function 'unknown' not supported |
| **Function Args** | `timestamp(name)` (name is string, not int64) | timestamp() requires int64 field |

## 📚 Key Concepts

### 1. Reflection in Go

Go's `reflect` package lets you inspect types at runtime:

```go
type User struct {
    Name string
    Age  int32
}

// Get type information
t := reflect.TypeOf(User{})

// Get field by name
field, ok := t.FieldByName("Name")
if ok {
    fmt.Println(field.Type) // string
    fmt.Println(field.Type.Kind()) // reflect.String
}

// Check if field exists
_, exists := t.FieldByName("Email") // false
```

### 2. AST Visitor Pattern

The validator walks the AST tree and validates each node:

```go
func (v *Validator) Validate(node ast.Node) []error {
    switch n := node.(type) {
    case *ast.Program:
        return v.validateProgram(n)
    case *ast.ComparisonExpression:
        return v.validateComparison(n)
    case *ast.TraversalExpression:
        return v.validateTraversal(n)
    // ... handle each node type
    }
}
```

### 3. Type Compatibility

Check if operator operands have compatible types:

```go
// age > 25
// Left: identifier "age" → int32 (from struct reflection)
// Right: int literal 25 → int
// Compatible? Yes (both numeric)

// name > 100
// Left: identifier "name" → string
// Right: int literal 100 → int
// Compatible? No! ❌
```

### 4. Multiple Error Collection

Instead of failing on first error, collect all validation issues:

```go
type ValidationError struct {
    Field   string
    Message string
    Pos     int
}

// Validate returns ALL errors
errs := validator.Validate(ast)
// [
//   "field 'email' not found at position 0",
//   "type mismatch at position 20"
// ]
```

### 5. AIP-160 Type-Specific Rules

Different types have different operator restrictions per AIP-160:

#### Strings
- **Allowed operators**: `=`, `!=`, `<`, `>`, `<=`, `>=`
- **Wildcard support**: `name = "*.example.com"` (ends with .example.com)
- **Lexical ordering**: `name > "foo"` uses alphabetical order

```go
name = "John"           ✅ Equality
name != "Jane"          ✅ Inequality  
name > "Alice"          ✅ Comparison (lexical)
name = "*.com"          ✅ Wildcard pattern
```

#### Numeric (int, int32, int64, float32, float64)
- **Allowed operators**: `=`, `!=`, `<`, `>`, `<=`, `>=`
- **Exponents supported**: `value = 2.997e9`
- **Cross-type compatible**: int32 can compare with int, int64, etc.

```go
age = 30                ✅ Equality
age > 18                ✅ Comparison
count <= 100            ✅ Comparison
rate = 2.5e3            ✅ Scientific notation
```

#### Booleans
- **Allowed operators**: `=`, `!=` **ONLY**
- **NOT allowed**: `<`, `>`, `<=`, `>=`
- **Values**: `true` and `false` literals only

```go
active = true           ✅ Equality
verified != false       ✅ Inequality
active > true           ❌ Comparison not allowed on booleans
```

#### Enums
- **Allowed operators**: `=`, `!=` **ONLY**
- **NOT allowed**: `<`, `>`, `<=`, `>=`
- **Values**: String representation (case-sensitive)

```go
status = "ACTIVE"       ✅ Equality (enum string value)
status != "DELETED"     ✅ Inequality
status > "ACTIVE"       ❌ Comparison not allowed on enums
```

**Note**: Go doesn't have first-class enums. Enum validation requires custom metadata or field tags.

#### Timestamps
- **Allowed operators**: `=`, `!=`, `<`, `>`, `<=`, `>=`
- **Format**: RFC-3339 (e.g., `"2012-04-21T11:30:00-04:00"`)
- **Go representation**: Often `int64` (Unix epoch) or `time.Time`

```go
created_at > "2024-01-01T00:00:00Z"   ✅ RFC-3339 format
created_at < "2024-12-31T23:59:59Z"   ✅ Comparison
timestamp(created_at) > 1704067200    ✅ Unix epoch (int64)
```

**Note**: This tutorial uses int64 for timestamps (Unix epoch seconds). RFC-3339 parsing should be done by your parser/lexer.

#### Durations
- **Allowed operators**: `=`, `!=`, `<`, `>`, `<=`, `>=`
- **Format**: Numeric value + 's' suffix (e.g., "20s", "1.2s")
- **Go representation**: `time.Duration` (int64 nanoseconds)

```go
timeout = "30s"         ✅ 30 seconds
latency < "0.5s"        ✅ 500 milliseconds  
duration >= "1.5s"      ✅ 1.5 seconds
```

**Note**: Duration parsing should be handled by your lexer. This validator checks the field type is duration-compatible.

## 🎓 Tutorial Structure

### Two Examples

**Example 1: Simple User (for learning basics)**
```go
type SimpleUser struct {
    Name   string
    Age    int32
    Active bool
}
```

**Example 2: Rich User (for complete validation)**
```go
type User struct {
    Name      string
    Email     struct {
        Address  string
        Verified bool
    }
    Age       int32
    Tags      []string
    CreatedAt int64
}
```

**Example 3: User with Struct Tags (protobuf/json support)**
```go
type UserWithTags struct {
    ID        int64    `json:"id,omitempty" protobuf:"varint,1,opt,name=id,proto3"`
    Name      string   `json:"name,omitempty" protobuf:"bytes,2,opt,name=name,proto3"`
    Email     string   `json:"email_address" protobuf:"bytes,3,opt,name=email_address,proto3"`
    IsActive  bool     `json:"is_active" protobuf:"varint,5,opt,name=is_active,proto3"`
    Tags      []string `json:"tags" protobuf:"bytes,6,rep,name=tags,proto3"`
}

// Validate against json tags (lowercase snake_case)
validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithJSONTags())
// Now filters use: id, name, email_address, is_active, tags

// Validate against protobuf tags (extracts name= value)
validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithProtobufTags())
// Now filters use: id, name, email_address, is_active, tags

// Default behavior: PascalCase field names
validator := NewValidator(reflect.TypeOf(UserWithTags{}))
// Filters use: ID, Name, Email, IsActive, Tags
```

### Tasks

You'll implement validation in 6 steps:

1. **Task 1**: Field existence validation (simple fields)
2. **Task 2**: Nested field traversal (`email.address`)
3. **Task 3**: Type compatibility checking (including AIP-160 boolean restrictions)
4. **Task 4**: Array field validation (has operator, handling bare identifiers)
5. **Task 5**: Function call validation
6. **Task 6**: Struct tag support (json/protobuf tags with functional options)

Each task has tests that guide your implementation.

## ⭐ Quick Start

### 1. Understand the Structure

First, look at the test file to see what you're building:
```bash
cat ../../pkg/filter/validator/validator_test.go
```

### 2. Run Tests (They'll Fail)

```bash
cd /Users/zshainky/Projects/aip160
go test ./pkg/filter/validator -v
```

You'll see failing tests - that's expected! Your job is to make them pass.

### 3. Start with Task 1

Read Task 1 below and implement field existence validation.

### 4. Run Task 1 Tests

```bash
go test ./pkg/filter/validator -v -run "TestTask1"
```

### 5. Repeat for Each Task

Continue with Tasks 2-5, running tests after each one.

## 📝 Implementation Guide

### ⭐ Task 1: Field Existence Validation

**Goal**: Validate that simple field names exist in the struct.

**Where to add code**: `pkg/filter/validator/validator.go`

**What to implement**:

1. **Create Validator struct with optional configuration**:
```go
type Validator struct {
    structType reflect.Type
    tagName    string // Tag name to use for field matching (e.g., "json", "protobuf"), empty = use field name
}

type ValidatorOption func(*Validator)

// NewValidator creates validator with optional tag configuration
func NewValidator(structType reflect.Type, opts ...ValidatorOption) *Validator {
    v := &Validator{structType: structType}
    // NOTE: Only apply the first option. Currently we only support one tagName option.
    if len(opts) > 0 {
        opts[0](v)
    }
    return v
}
```

2. **Implement Validate method**:
```go
func (v *Validator) Validate(node ast.Node) []error {
    // Walk the AST and collect errors
    // Start with the Program node and recurse
}
```

3. **Validate Identifier nodes**:
```go
func (v *Validator) validateIdentifier(name string) error {
    _, ok := v.structType.FieldByName(name)
    if !ok {
        return fmt.Errorf("field '%s' does not exist", name)
    }
    return nil
}
```

**Example**:
```go
// Input: "email = 'test@example.com'"
// Struct: SimpleUser (no email field)
// Output: []error{"field 'email' does not exist"}
```

**Success criteria**:
```bash
go test ./pkg/filter/validator -v -run "TestTask1"
# All Task 1 tests pass
```

---

### ⭐ Task 2: Nested Field Traversal

**Goal**: Validate traversal expressions like `email.address`.

**Where to add code**: `pkg/filter/validator/validator.go`

**What to implement**:

1. **Handle TraversalExpression nodes**:
```go
func (v *Validator) validateTraversal(expr *ast.TraversalExpression) (reflect.Type, error) {
    // Start with base type
    currentType := v.structType
    
    // Walk the traversal chain
    for _, segment := range expr.Segments {
        field, ok := currentType.FieldByName(segment)
        if !ok {
            return nil, fmt.Errorf("field '%s' does not exist", segment)
        }
        currentType = field.Type
    }
    
    return currentType, nil  // Return final type for type checking
}
```

**Example**:
```go
// Input: "email.address = 'test'"
// Struct: User with Email struct { Address string }
// Validation:
//   1. Check 'email' exists → Yes (struct field)
//   2. Check 'address' exists in Email struct → Yes (string field)
//   3. Return final type: string
```

**Success criteria**:
```bash
go test ./pkg/filter/validator -v -run "TestTask2"
```

---

### ⭐ Task 3: Type Compatibility Checking

**Goal**: Validate that comparison operands have compatible types AND enforce AIP-160 operator restrictions.

**Where to add code**: `pkg/filter/validator/validator.go`

**What to implement**:

1. **Get literal types**:
```go
func getLiteralType(lit ast.Node) reflect.Kind {
    switch lit.(type) {
    case *ast.StringLiteral:
        return reflect.String
    case *ast.IntLiteral:
        return reflect.Int
    case *ast.BoolLiteral:
        return reflect.Bool
    // Add more types...
    }
}
```

2. **Check type compatibility AND operator validity**:
```go
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) error {
    // Get left side type (from struct field)
    leftType := v.getExpressionType(expr.Left)
    
    // Get right side type (from literal or field)
    rightType := v.getExpressionType(expr.Right)
    
    // AIP-160: Booleans only support = and !=
    if leftType.Kind() == reflect.Bool {
        if expr.Operator != "=" && expr.Operator != "!=" {
            return fmt.Errorf("boolean fields only support = and != operators, not '%s'", expr.Operator)
        }
    }
    
    // Check type compatibility
    if !typesCompatible(leftType, rightType) {
        return fmt.Errorf("type mismatch: %s vs %s", leftType, rightType)
    }
    
    return nil
}

func typesCompatible(left, right reflect.Kind) bool {
    // String can only compare with string
    if left == reflect.String || right == reflect.String {
        return left == right
    }
    
    // Booleans only compatible with booleans
    if left == reflect.Bool || right == reflect.Bool {
        return left == right
    }
    
    // Numeric types are compatible with each other
    if isNumericKind(left) && isNumericKind(right) {
        return true
    }
    
    return false
}
```

**Example**:
```go
// Input: "name > 100"
// name is string, 100 is int
// Error: "type mismatch: cannot compare string with int"

// Input: "active > true"
// active is bool, operator is >
// Error: "boolean fields only support = and != operators, not '>'"
```

**Success criteria**:
```bash
go test ./pkg/filter/validator -v -run "TestTask3"
```

---

### ⭐ Task 4: Array Field Validation

**Goal**: Validate that array fields use the has operator (`:`), and handle bare identifiers correctly.

**Where to add code**: `pkg/filter/validator/validator.go`

**What to implement**:

1. **Validate HasExpression**:
```go
func (v *Validator) validateHasExpression(expr *ast.HasExpression) error {
    // Validate the collection (left side) - this should be a field
    errors = append(errors, v.validateNode(expr.Collection)...)
    
    // IMPORTANT: For the member (right side), DON'T validate it as a field reference!
    // In AIP-160, bare identifiers in has expressions are treated as string literals
    // Example: Tags:urgent means check if "urgent" is in Tags array (not a field named "urgent")
    
    // Get the field type
    fieldType := v.getExpressionType(expr.Collection)
    
    // Check if it's an array/slice
    if fieldType.Kind() != reflect.Slice && fieldType.Kind() != reflect.Array {
        return fmt.Errorf("has operator ':' requires array/slice field, got %s", fieldType.Kind())
    }
    
    // Get element type
    elemType := fieldType.Elem()
    
    // Determine member type based on node type
    var memberType reflect.Type
    switch expr.Member.(type) {
    case *ast.StringLiteral:
        memberType = reflect.TypeOf("")
    case *ast.NumberLiteral:
        memberType = reflect.TypeOf(float64(0))
    case *ast.BooleanLiteral:
        memberType = reflect.TypeOf(false)
    case *ast.Identifier:
        // In AIP-160, bare identifiers are treated as strings
        // Tags:urgent is equivalent to Tags:"urgent"
        memberType = reflect.TypeOf("")
    }
    
    // Validate the value type matches element type
    if !typesCompatible(elemType.Kind(), memberType.Kind()) {
        return fmt.Errorf("has operator: element type %s incompatible with %s", elemType, memberType)
    }
    
    return nil
}
```

2. **Reject equality operator on arrays**:
```go
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) error {
    leftType := v.getExpressionType(expr.Left)
    
    // Check if left is array/slice
    if leftType.Kind() == reflect.Slice || leftType.Kind() == reflect.Array {
        return fmt.Errorf("use has operator ':' for array/slice fields, not '%s'", expr.Operator)
    }
    
    // ... rest of comparison validation
}
```

**Example**:
```go
// Input: "tags:urgent"
// tags is []string, "urgent" is treated as string literal
// Valid: ✅

// Input: "tags = 'urgent'"
// Error: "use has operator ':' for array fields, not '='"
```

**Success criteria**:
```bash
go test ./pkg/filter/validator -v -run "TestTask4"
```

---

### ⭐ Task 5: Function Call Validation

**Goal**: Validate function calls against a supported function registry.

**Where to add code**: `pkg/filter/validator/functions.go` and `validator.go`

**What to implement**:

1. **Define function registry** (`functions.go`):
```go
type FunctionDef struct {
    Name            string
    ArgCount        int
    ValidFieldKinds []reflect.Kind
}

var SupportedFunctions = map[string]FunctionDef{
    "timestamp": {
        Name:            "timestamp",
        ArgCount:        1,
        ValidFieldKinds: []reflect.Kind{reflect.Int64, reflect.Int32},
    },
    "size": {
        Name:            "size",
        ArgCount:        1,
        ValidFieldKinds: []reflect.Kind{reflect.Slice, reflect.Array, reflect.String},
    },
}
```

2. **Validate function calls** (`validator.go`):
```go
func (v *Validator) validateFunctionCall(expr *ast.FunctionCall) error {
    // Check if function exists
    funcDef, ok := SupportedFunctions[expr.Name]
    if !ok {
        return fmt.Errorf("function '%s' is not supported", expr.Name)
    }
    
    // Check argument count
    if len(expr.Arguments) != funcDef.ArgCount {
        return fmt.Errorf("function '%s' expects %d argument(s), got %d", 
            expr.Name, funcDef.ArgCount, len(expr.Arguments))
    }
    
    // Validate argument types
    for i, arg := range expr.Arguments {
        argType := v.getExpressionType(arg)
        if !isValidKind(argType.Kind(), funcDef.ValidFieldKinds) {
            return fmt.Errorf("function '%s' argument %d: expected %v, got %s",
                expr.Name, i+1, funcDef.ValidFieldKinds, argType.Kind())
        }
    }
    
    return nil
}
```

**Example**:
```go
// Input: "timestamp(created_at) > 1000"
// created_at is int64
// Valid: ✅

// Input: "timestamp(name)"
// name is string
// Error: "function 'timestamp' requires int32/int64 field, got string"

// Input: "unknown_func(x)"
// Error: "function 'unknown_func' is not supported"
```

**Success criteria**:
```bash
go test ./pkg/filter/validator -v -run "TestTask5"
# All tests pass!
```

---

### ⭐ Task 6: Struct Tag Support

**Goal**: Support validating against json and protobuf struct tags instead of PascalCase field names.

**Why**: Proto-generated structs and JSON APIs often use snake_case field names in tags while Go uses PascalCase struct field names. Users should be able to validate filters against the tag names their API actually exposes.

**Where to add code**: `pkg/filter/validator/validator.go`

**What to implement**:

1. **Add functional option pattern**:
```go
type ValidatorOption func(*Validator)

// WithJSONTags configures validator to match against json struct tags
func WithJSONTags() ValidatorOption {
    return func(v *Validator) {
        v.tagName = "json"
    }
}

// WithProtobufTags configures validator to match against protobuf struct tags
func WithProtobufTags() ValidatorOption {
    return func(v *Validator) {
        v.tagName = "protobuf"
    }
}

// WithTag configures validator to match against any custom struct tag
func WithTag(tagName string) ValidatorOption {
    return func(v *Validator) {
        v.tagName = tagName
    }
}
```

2. **Implement field lookup with tag support**:
```go
// findFieldByName finds a struct field by name, respecting the configured tag name.
// If tagName is empty, it uses the field name directly (PascalCase).
// Otherwise, it searches through fields matching the tag value.
func (v *Validator) findFieldByName(structType reflect.Type, name string) (reflect.StructField, bool) {
    // Default behavior: use field name directly
    if v.tagName == "" {
        return structType.FieldByName(name)
    }
    
    // Search through all fields to find matching tag
    for i := 0; i < structType.NumField(); i++ {
        field := structType.Field(i)
        tagValue := field.Tag.Get(v.tagName)
        
        if tagValue == "" {
            continue
        }
        
        // Extract the field name from the tag
        var tagFieldName string
        if v.tagName == "protobuf" {
            // Protobuf format: "bytes,2,opt,name=field_name,proto3"
            tagFieldName = extractProtobufName(tagValue)
        } else {
            // JSON and most other tags: "field_name,omitempty"
            // Take the first part before comma
            tagFieldName = strings.Split(tagValue, ",")[0]
        }
        
        if tagFieldName == name {
            return field, true
        }
    }
    
    return reflect.StructField{}, false
}

// extractProtobufName extracts the "name=value" from a protobuf tag.
// Example: "bytes,2,opt,name=field_name,proto3" returns "field_name"
func extractProtobufName(tag string) string {
    parts := strings.Split(tag, ",")
    for _, part := range parts {
        part = strings.TrimSpace(part)
        if strings.HasPrefix(part, "name=") {
            return strings.TrimPrefix(part, "name=")
        }
    }
    return ""
}
```

3. **Update all field lookups to use findFieldByName**:
Replace all calls to `structType.FieldByName(name)` with `v.findFieldByName(structType, name)`.

**Example**:
```go
type UserWithTags struct {
    ID    int64  `json:"id" protobuf:"varint,1,opt,name=id,proto3"`
    Name  string `json:"name" protobuf:"bytes,2,opt,name=name,proto3"`
}

// Default: PascalCase field names
validator := NewValidator(reflect.TypeOf(UserWithTags{}))
// Filters use: ID, Name

// JSON tags: lowercase
validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithJSONTags())
// Filters use: id, name

// Protobuf tags: extracts name= value  
validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithProtobufTags())
// Filters use: id, name

// Custom tag
validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithTag("db"))
// Filters use whatever is in db:"..." tags
```

**Design Decision**: 
- Only one tag option is allowed per validator (first option wins if multiple provided)
- This keeps the implementation simple since we only need one tagName field
- Can be expanded in the future if multiple option types are needed

**Success criteria**:
```bash
go test ./pkg/filter/validator -v -run "TestJSONTags|TestProtobufTags|TestDefaultBehavior"
# All tag-related tests pass!
```

---

## 🧪 Testing Your Implementation

### Run All Tests
```bash
go test ./pkg/filter/validator -v
```

### Test Simple Examples
```go
// In validator_test.go - SimpleUser tests
type SimpleUser struct {
    Name   string
    Age    int32
    Active bool
}

// Valid filter
filter := parser.Parse("name = 'John' AND age > 25")
validator := NewValidator(reflect.TypeOf(SimpleUser{}))
errs := validator.Validate(filter)
assert.Empty(t, errs)

// Invalid filter
filter = parser.Parse("email = 'test@example.com'")  // No email field
errs = validator.Validate(filter)
assert.Contains(t, errs[0].Error(), "email")
```

### Test Complex Examples
```go
// Rich User with nested fields, arrays, functions
type User struct {
    Name      string
    Email     struct {
        Address  string
        Verified bool
    }
    Age       int32
    Tags      []string
    CreatedAt int64
}

// Test nested traversal
filter := parser.Parse("email.address = 'test@example.com'")
validator := NewValidator(reflect.TypeOf(User{}))
errs := validator.Validate(filter)
assert.Empty(t, errs)

// Test array has operator
filter = parser.Parse("tags:urgent")
errs = validator.Validate(filter)
assert.Empty(t, errs)

// Test function validation
filter = parser.Parse("timestamp(created_at) > 1000")
errs = validator.Validate(filter)
assert.Empty(t, errs)
```

## 💡 Tips and Gotchas

### Tip 1: Named vs Anonymous Structs

Proto-generated structs often use anonymous fields:
```go
type User struct {
    Email struct {  // Anonymous struct!
        Address string
    }
}

// For anonymous structs, you need field.Type
field, _ := t.FieldByName("Email")
emailType := field.Type  // This is the struct type
addressField, _ := emailType.FieldByName("Address")
```

### Tip 2: Handling Slice Element Types

```go
field, _ := t.FieldByName("Tags")  // Tags []string
field.Type.Kind()  // reflect.Slice
field.Type.Elem()  // reflect.Type for string
field.Type.Elem().Kind()  // reflect.String
```

### Tip 3: Numeric Type Compatibility

Be flexible with numeric types:
```go
// These should all be compatible:
int32 vs int     ✅
int64 vs int32   ✅
float64 vs int   ✅ (with warning in production code)
```

### Tip 4: Error Messages Should Be Clear

Bad: "invalid"
Good: "field 'email' does not exist in User"
Better: "field 'email' does not exist in User at position 0"

### Tip 5: Collect ALL Errors

Don't stop at first error! Collect all validation issues so users can fix multiple problems at once.

### Tip 6: AIP-160 Boolean Operator Restriction

Remember that booleans only support `=` and `!=`:
```go
active = true     ✅ Valid
active != false   ✅ Valid
active > true     ❌ Invalid (boolean comparison not allowed)
verified <= false ❌ Invalid (boolean comparison not allowed)
```

This is enforced by checking the field type and operator in `validateComparison`.

### Tip 7: Struct Tags for Proto/JSON APIs

When working with protobuf-generated structs or JSON APIs, use tag-based validation:

```go
// Proto-generated struct
type User struct {
    ID   int64  `protobuf:"varint,1,opt,name=id,proto3" json:"id,omitempty"`
    Name string `protobuf:"bytes,2,opt,name=name,proto3" json:"name,omitempty"`
}

// Match against json tags (lowercase)
validator := NewValidator(reflect.TypeOf(User{}), WithJSONTags())
// Filter: id = 123 (not ID = 123)

// Match against protobuf name= values
validator := NewValidator(reflect.TypeOf(User{}), WithProtobufTags())
// Filter: name = "John" (extracted from protobuf tag)

// Default: PascalCase struct field names
validator := NewValidator(reflect.TypeOf(User{}))
// Filter: Name = "John" (not name)
```

**Why this matters**: API filters typically use the external field names (json/protobuf tags) not internal Go field names.

### Tip 8: Bare Identifiers in Has Operator

In AIP-160, bare identifiers in has expressions are string literals:
```go
tags:urgent      // Equivalent to tags:"urgent"
tags:bug         // Equivalent to tags:"bug" 
```

Don't validate the right side of `:` as a field reference - it's a literal value!

## ⚠️ AIP-160 Limitations and Extensions

### Implemented Features

This validator implements all core AIP-160 schematic validation requirements:

✅ **Field existence validation** - Fields referenced must exist in the schema  
✅ **Type compatibility checking** - Field values must match field types  
✅ **Comparison operators** - Enforces operator restrictions by type (booleans only support = and !=)  
✅ **Array operators** - Enforces has operator (`:`) for array/slice fields  
✅ **Function validation** - Validates function calls and argument types  
✅ **Struct tag support** - Match against json/protobuf tags instead of PascalCase field names
✅ **Bare identifiers in has** - Correctly treats `Tags:urgent` as `Tags:"urgent"` per AIP-160

### Not Implemented (Advanced Topics)

The following AIP-160 features require additional work:

1. **Enum Validation**
   - **Issue**: Go doesn't have first-class enums
   - **AIP-160**: Enums should only support `=` and `!=`, not `<`, `>`, etc.
   - **Solution**: Implement custom enum detection via struct tags or metadata
   ```go
   type User struct {
       Status int32 `validate:"enum"`  // Custom tag
   }
   ```

2. **Wildcard Pattern Validation**
   - **AIP-160**: Support wildcards like `name = "*.example.com"`
   - **Current**: Parser accepts wildcards, validator doesn't validate pattern syntax
   - **Solution**: Add wildcard pattern validation in string literal processing

3. **Timestamp Format Validation**
   - **AIP-160**: Timestamps should use RFC-3339 format
   - **Current**: We validate int64 fields (Unix epoch)
   - **Note**: Many APIs use int64 timestamps. RFC-3339 parsing should be in lexer/parser

4. **Duration Format Validation**
   - **AIP-160**: Durations use numeric + 's' suffix (e.g., "30s", "1.5s")
   - **Current**: Not specifically handled
   - **Note**: If using `time.Duration`, it's represented as int64 nanoseconds

### Extending the Validator

You can extend the validator for production use:

```go
// Custom validator with enum support
type ProductionValidator struct {
    *Validator
    enumFields map[string][]string  // field name -> valid values
}

func (pv *ProductionValidator) validateComparison(expr *ast.ComparisonExpression) []error {
    // Call base validation
    errs := pv.Validator.validateComparison(expr)
    
    // Add custom enum validation
    if ident, ok := expr.Left.(*ast.Identifier); ok {
        if validValues, isEnum := pv.enumFields[ident.Value]; isEnum {
            // Check operator (only = and !=)
            if expr.Operator != "=" && expr.Operator != "!=" {
                errs = append(errs, fmt.Errorf("enum field '%s' only supports = and !=", ident.Value))
            }
            // Check value is in valid set
            if strLit, ok := expr.Right.(*ast.StringLiteral); ok {
                if !contains(validValues, strLit.Value) {
                    errs = append(errs, fmt.Errorf("invalid enum value '%s' for field '%s'", strLit.Value, ident.Value))
                }
            }
        }
    }
    
    return errs
}
```

## 🎯 Success Criteria

You've completed Module 6 when:

✅ All 38+ tests pass  
✅ Validator catches non-existent fields  
✅ Validator handles nested traversal  
✅ Type mismatches are detected  
✅ Boolean comparison operators are rejected (AIP-160 compliance)  
✅ Array fields require has operator  
✅ Bare identifiers in has expressions treated as strings (AIP-160 compliance)  
✅ Function calls are validated  
✅ JSON struct tags are supported (WithJSONTags option)  
✅ Protobuf struct tags are supported (WithProtobufTags option)  
✅ Multiple errors are collected and returned  

## 🚀 Next Steps

After completing Module 6, you have a complete AIP-160 filter parser with validation!

**What you've built:**
- ✅ Lexer (Module 1)
- ✅ AST data structures (Module 2)
- ✅ Recursive descent parser (Modules 3-5)
- ✅ Schema validator with struct tag support (Module 6)

**Key Features Implemented:**
- Field existence and nested traversal validation
- Type compatibility checking with AIP-160 compliance
- Boolean operator restrictions (= and != only)
- Array/slice has operator validation
- Function call validation with registry
- JSON struct tag support for API field names
- Protobuf struct tag support for proto-generated structs
- Functional options pattern for extensibility

**Future modules** (optional, not in this tutorial):
- Module 7: AST-to-SQL translation
- Module 8: Query optimization
- Module 9: Custom function extensions

**Production considerations:**
- Add position tracking for better error messages
- ✅ Handle proto field name mappings (implemented via WithProtobufTags)
- ✅ Handle JSON field name mappings (implemented via WithJSONTags)
- Add custom function registry per resource type
- Implement validation caching for performance
- Consider supporting multiple tag options if needed (currently first-wins)

---

**Ready to begin?** Start with Task 1! 🎓

Need help? Check:
- [HINTS.md](HINTS.md) - Nudges in the right direction
- [SOLUTION.md](SOLUTION.md) - Complete implementation with explanations
