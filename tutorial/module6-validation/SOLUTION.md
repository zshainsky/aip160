# Module 6: Schema Validation - Solution

This document contains the complete solution for Module 6 with detailed explanations.

## Complete Implementation

### validator.go (Complete Solution)

```go
package validator

import (
	"fmt"
	"reflect"

	"github.com/zshainky/aip160/pkg/filter/ast"
)

// Validator validates filter ASTs against a Go struct schema using reflection.
type Validator struct {
	structType reflect.Type
}

// NewValidator creates a new validator for the given struct type.
func NewValidator(structType reflect.Type) *Validator {
	return &Validator{
		structType: structType,
	}
}

// Validate validates the entire AST and returns a list of all validation errors.
func (v *Validator) Validate(node ast.Node) []error {
	program, ok := node.(*ast.Program)
	if !ok {
		return []error{fmt.Errorf("expected Program node, got %T", node)}
	}

	return v.validateNode(program.Expression)
}

// validateNode validates a single AST node and its children recursively.
func (v *Validator) validateNode(node ast.Node) []error {
	var errors []error

	switch n := node.(type) {
	case *ast.LogicalExpression:
		errors = append(errors, v.validateLogical(n)...)

	case *ast.ComparisonExpression:
		errors = append(errors, v.validateComparison(n)...)

	case *ast.UnaryExpression:
		errors = append(errors, v.validateUnary(n)...)

	case *ast.Identifier:
		if err := v.validateIdentifier(n.Value); err != nil {
			errors = append(errors, err)
		}

	case *ast.TraversalExpression:
		_, err := v.validateTraversal(n)
		if err != nil {
			errors = append(errors, err)
		}

	case *ast.HasExpression:
		errors = append(errors, v.validateHasExpression(n)...)

	case *ast.FunctionCall:
		errors = append(errors, v.validateFunctionCall(n)...)

	case *ast.StringLiteral, *ast.IntLiteral, *ast.BoolLiteral, *ast.FloatLiteral:
		// Literals don't need validation
		return nil

	default:
		// Unknown node type - shouldn't happen with current AST
		return []error{fmt.Errorf("unknown node type: %T", node)}
	}

	return errors
}

// Task 1: Validate simple identifier field existence
func (v *Validator) validateIdentifier(name string) error {
	_, ok := v.structType.FieldByName(name)
	if !ok {
		return fmt.Errorf("field '%s' does not exist", name)
	}
	return nil
}

// Task 2: Validate nested field traversal
func (v *Validator) validateTraversal(expr *ast.TraversalExpression) (reflect.Type, error) {
	// Start with the struct type
	currentType := v.structType

	// First, validate and resolve the base identifier
	baseIdent, ok := expr.Base.(*ast.Identifier)
	if !ok {
		return nil, fmt.Errorf("expected identifier as traversal base, got %T", expr.Base)
	}

	field, ok := currentType.FieldByName(baseIdent.Value)
	if !ok {
		return nil, fmt.Errorf("field '%s' does not exist", baseIdent.Value)
	}
	currentType = field.Type

	// Now traverse each segment
	for _, segment := range expr.Segments {
		// Check if current type is a struct
		if currentType.Kind() != reflect.Struct {
			return nil, fmt.Errorf("cannot traverse into non-struct type %s", currentType.Kind())
		}

		field, ok := currentType.FieldByName(segment)
		if !ok {
			return nil, fmt.Errorf("field '%s' does not exist", segment)
		}
		currentType = field.Type
	}

	return currentType, nil
}

// Task 3: Validate comparison expression
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) []error {
	var errors []error

	// Get types for both sides
	leftType := v.getExpressionType(expr.Left)
	if leftType == nil {
		// Field doesn't exist - will be caught by identifier/traversal validation
		// Still validate right side for completeness
		errors = append(errors, v.validateNode(expr.Right)...)
		return errors
	}

	// Check if left side is array/slice - should use has operator instead
	if leftType.Kind() == reflect.Slice || leftType.Kind() == reflect.Array {
		errors = append(errors, fmt.Errorf("use has operator ':' for array/slice fields, not '%s'", expr.Operator))
		return errors
	}

	// AIP-160: Booleans only support = and != operators
	if leftType.Kind() == reflect.Bool {
		if expr.Operator != "=" && expr.Operator != "!=" {
			errors = append(errors, fmt.Errorf("boolean fields only support = and != operators, not '%s'", expr.Operator))
			return errors
		}
	}

	rightType := v.getExpressionType(expr.Right)
	if rightType == nil {
		errors = append(errors, v.validateNode(expr.Right)...)
		return errors
	}

	// Check type compatibility
	if !typesCompatible(leftType.Kind(), rightType.Kind()) {
		errors = append(errors, fmt.Errorf("type mismatch: cannot compare %s with %s", leftType.Kind(), rightType.Kind()))
	}

	// Recursively validate both sides
	errors = append(errors, v.validateNode(expr.Left)...)
	errors = append(errors, v.validateNode(expr.Right)...)

	return errors
}

// Task 4: Validate has expression
func (v *Validator) validateHasExpression(expr *ast.HasExpression) []error {
	var errors []error

	// Get the collection type
	collectionType := v.getExpressionType(expr.Collection)
	if collectionType == nil {
		// Field doesn't exist - will be caught by other validation
		errors = append(errors, v.validateNode(expr.Collection)...)
		return errors
	}

	// Check that collection is actually an array/slice
	if collectionType.Kind() != reflect.Slice && collectionType.Kind() != reflect.Array {
		errors = append(errors, fmt.Errorf("has operator ':' requires array/slice field, got %s", collectionType.Kind()))
		return errors
	}

	// Get element type
	elemType := collectionType.Elem()

	// Get value type
	valueType := v.getExpressionType(expr.Value)
	if valueType == nil {
		errors = append(errors, v.validateNode(expr.Value)...)
		return errors
	}

	// Check type compatibility
	if !typesCompatible(elemType.Kind(), valueType.Kind()) {
		errors = append(errors, fmt.Errorf("has operator: element type %s incompatible with value type %s", elemType.Kind(), valueType.Kind()))
	}

	// Recursively validate both sides
	errors = append(errors, v.validateNode(expr.Collection)...)
	errors = append(errors, v.validateNode(expr.Value)...)

	return errors
}

// Task 5: Validate function call
func (v *Validator) validateFunctionCall(expr *ast.FunctionCall) []error {
	var errors []error

	// Look up function in registry
	funcDef, ok := SupportedFunctions[expr.Name]
	if !ok {
		return []error{fmt.Errorf("function '%s' is not supported", expr.Name)}
	}

	// Check argument count
	if len(expr.Arguments) != funcDef.ArgCount {
		errors = append(errors, fmt.Errorf("function '%s' expects %d argument(s), got %d",
			expr.Name, funcDef.ArgCount, len(expr.Arguments)))
		return errors
	}

	// Validate each argument
	for i, arg := range expr.Arguments {
		// First validate that the argument node itself is valid
		argErrors := v.validateNode(arg)
		errors = append(errors, argErrors...)

		// Then check type compatibility
		argType := v.getExpressionType(arg)
		if argType != nil && !isValidKind(argType.Kind(), funcDef.ValidFieldKinds) {
			errors = append(errors, fmt.Errorf("function '%s' argument %d: expected field types %v, got %s",
				expr.Name, i+1, funcDef.ValidFieldKinds, argType.Kind()))
		}
	}

	return errors
}

// validateLogical validates logical AND/OR expressions
func (v *Validator) validateLogical(expr *ast.LogicalExpression) []error {
	var errors []error

	// Validate left side
	leftErrs := v.validateNode(expr.Left)
	errors = append(errors, leftErrs...)

	// Validate right side
	rightErrs := v.validateNode(expr.Right)
	errors = append(errors, rightErrs...)

	return errors
}

// validateUnary validates unary NOT expressions
func (v *Validator) validateUnary(expr *ast.UnaryExpression) []error {
	return v.validateNode(expr.Operand)
}

// getExpressionType returns the Go type for an expression node
func (v *Validator) getExpressionType(node ast.Node) reflect.Type {
	switch n := node.(type) {
	case *ast.Identifier:
		field, ok := v.structType.FieldByName(n.Value)
		if !ok {
			return nil
		}
		return field.Type

	case *ast.TraversalExpression:
		typ, err := v.validateTraversal(n)
		if err != nil {
			return nil
		}
		return typ

	case *ast.StringLiteral:
		var s string
		return reflect.TypeOf(s)

	case *ast.IntLiteral:
		var i int
		return reflect.TypeOf(i)

	case *ast.BoolLiteral:
		var b bool
		return reflect.TypeOf(b)

	case *ast.FloatLiteral:
		var f float64
		return reflect.TypeOf(f)

	case *ast.FunctionCall:
		// For now, assume functions return int64 (like timestamp)
		// In a more sophisticated system, you'd look up the return type
		var i int64
		return reflect.TypeOf(i)

	default:
		return nil
	}
}

// typesCompatible checks if two types are compatible for comparison
func typesCompatible(left, right reflect.Kind) bool {
	// Same type is always compatible
	if left == right {
		return true
	}

	// Strings only compatible with strings
	if left == reflect.String || right == reflect.String {
		return left == right
	}

	// Bools only compatible with bools
	if left == reflect.Bool || right == reflect.Bool {
		return left == right
	}

	// All numeric types are compatible with each other
	if isNumericKind(left) && isNumericKind(right) {
		return true
	}

	return false
}

// isNumericKind checks if a kind represents a numeric type
func isNumericKind(k reflect.Kind) bool {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return true
	}
	return false
}

// isValidKind checks if a kind is in a list of valid kinds
func isValidKind(kind reflect.Kind, validKinds []reflect.Kind) bool {
	for _, validKind := range validKinds {
		if kind == validKind {
			return true
		}
	}
	return false
}
```

## Explanation by Task

### Task 1: Field Existence Validation

**Key Concept**: Use `reflect.Type.FieldByName()` to check if a field exists.

```go
func (v *Validator) validateIdentifier(name string) error {
	_, ok := v.structType.FieldByName(name)
	if !ok {
		return fmt.Errorf("field '%s' does not exist", name)
	}
	return nil
}
```

**How it works**:
1. `FieldByName` looks up a field in the struct
2. Returns `(StructField, bool)` where bool indicates if field was found
3. If not found, return error with the field name

**Test example**:
```go
type User struct { Name string }
validator := NewValidator(reflect.TypeOf(User{}))

// Valid
validator.Validate("name = 'John'")  // ✅

// Invalid
validator.Validate("email = 'test'")  // ❌ "field 'email' does not exist"
```

---

### Task 2: Nested Field Traversal

**Key Concept**: Walk the field path step by step, updating the current type.

```go
func (v *Validator) validateTraversal(expr *ast.TraversalExpression) (reflect.Type, error) {
	currentType := v.structType
	
	// Start with base identifier
	baseIdent := expr.Base.(*ast.Identifier)
	field, ok := currentType.FieldByName(baseIdent.Value)
	if !ok {
		return nil, fmt.Errorf("field '%s' does not exist", baseIdent.Value)
	}
	currentType = field.Type
	
	// Walk each segment
	for _, segment := range expr.Segments {
		field, ok := currentType.FieldByName(segment)
		if !ok {
			return nil, fmt.Errorf("field '%s' does not exist", segment)
		}
		currentType = field.Type
	}
	
	return currentType, nil
}
```

**How it works**:
1. Start with the struct type
2. Look up base field (e.g., "email") → get Email struct type
3. For each segment (e.g., "address"), look it up in current type
4. Update current type to the found field's type
5. Return final type (needed for type checking)

**Example**:
```go
type User struct {
	Email struct {
		Address string
	}
}

// "email.address"
// Step 1: Look up "email" in User → get Email struct type
// Step 2: Look up "address" in Email → get string type
// Result: string
```

---

### Task 3: Type Compatibility Checking

**Key Concept**: Get types for both sides, check compatibility rules, and enforce AIP-160 operator restrictions.

```go
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) []error {
	leftType := v.getExpressionType(expr.Left)
	rightType := v.getExpressionType(expr.Right)
	
	// Check for arrays (should use has operator)
	if leftType.Kind() == reflect.Slice || leftType.Kind() == reflect.Array {
		return []error{fmt.Errorf("use has operator ':' for array fields")}
	}

	// AIP-160: Booleans only support = and !=
	if leftType.Kind() == reflect.Bool {
		if expr.Operator != "=" && expr.Operator != "!=" {
			return []error{fmt.Errorf("boolean fields only support = and !=, not '%s'", expr.Operator)}
		}
	}
	
	// Check type compatibility
	if !typesCompatible(leftType.Kind(), rightType.Kind()) {
		return []error{fmt.Errorf("type mismatch: %s vs %s", leftType, rightType)}
	}
	
	return nil
}
```

**Type compatibility rules**:
- Strings only with strings
- Booleans only with booleans (AND only `=` and `!=` operators allowed per AIP-160)
- All numeric types are compatible (int, int32, int64, float, etc.)

**AIP-160 Boolean Restriction**:
Per [AIP-160 comparison operators](https://google.aip.dev/160#comparison-operators), boolean fields should only support `=` and `!=`, NOT `<`, `>`, `<=`, `>=`.

```go
active = true    ✅ Valid
active != false  ✅ Valid
active > true    ❌ Invalid (boolean comparison not allowed)
verified <= true ❌ Invalid (boolean comparison not allowed)
```

**Why allow numeric flexibility?**
```go
type User struct {
	Age int32
}

// Filter: age > 25
// Age is int32, literal 25 is int
// These should be compatible! ✅
```

---

### Task 4: Array Field Validation

**Key Concept**: Arrays must use has operator (`:`), not equality/comparison operators.

```go
func (v *Validator) validateHasExpression(expr *ast.HasExpression) []error {
	collectionType := v.getExpressionType(expr.Collection)
	
	// Check it's an array/slice
	if collectionType.Kind() != reflect.Slice && collectionType.Kind() != reflect.Array {
		return []error{fmt.Errorf("has operator requires array/slice, got %s", collectionType)}
	}
	
	// Check element type matches value type
	elemType := collectionType.Elem()  // Get []string → string
	valueType := v.getExpressionType(expr.Value)
	
	if !typesCompatible(elemType.Kind(), valueType.Kind()) {
		return []error{fmt.Errorf("element type %s incompatible with %s", elemType, valueType)}
	}
	
	return nil
}
```

**Also reject equality on arrays**:
```go
// In validateComparison:
if leftType.Kind() == reflect.Slice || leftType.Kind() == reflect.Array {
	return []error{fmt.Errorf("use has operator ':' for arrays")}
}
```

**Example**:
```go
type User struct {
	Tags []string
}

tags:urgent      ✅ Correct (has operator)
tags = "urgent"  ❌ Wrong (should use has operator)
```

---

### Task 5: Function Call Validation

**Key Concept**: Look up function in registry, validate arg count and types.

```go
func (v *Validator) validateFunctionCall(expr *ast.FunctionCall) []error {
	// Look up function
	funcDef, ok := SupportedFunctions[expr.Name]
	if !ok {
		return []error{fmt.Errorf("function '%s' not supported", expr.Name)}
	}
	
	// Check arg count
	if len(expr.Arguments) != funcDef.ArgCount {
		return []error{fmt.Errorf("wrong number of arguments")}
	}
	
	// Validate each argument type
	for i, arg := range expr.Arguments {
		argType := v.getExpressionType(arg)
		if !isValidKind(argType.Kind(), funcDef.ValidFieldKinds) {
			return []error{fmt.Errorf("argument %d has wrong type", i+1)}
		}
	}
	
	return nil
}
```

**Function registry** (from `functions.go`):
```go
var SupportedFunctions = map[string]FunctionDef{
	"timestamp": {
		ArgCount: 1,
		ValidFieldKinds: []reflect.Kind{reflect.Int64, reflect.Int32},
	},
	"size": {
		ArgCount: 1,
		ValidFieldKinds: []reflect.Kind{reflect.Slice, reflect.Array, reflect.String},
	},
}
```

**Example**:
```go
type User struct {
	Name      string
	CreatedAt int64
}

timestamp(created_at) > 1000  ✅ Valid (int64 field)
timestamp(name) > 1000        ❌ Invalid (string field)
unknown_func(x)               ❌ Invalid (function doesn't exist)
```

---

## Helper Functions Explained

### getExpressionType

Returns the Go type for any expression node:

```go
func (v *Validator) getExpressionType(node ast.Node) reflect.Type {
	switch n := node.(type) {
	case *ast.Identifier:
		field, _ := v.structType.FieldByName(n.Value)
		return field.Type
		
	case *ast.TraversalExpression:
		typ, _ := v.validateTraversal(n)
		return typ
		
	case *ast.StringLiteral:
		var s string
		return reflect.TypeOf(s)  // Returns type for string
		
	case *ast.IntLiteral:
		var i int
		return reflect.TypeOf(i)
		
	// ... more cases
	}
}
```

### typesCompatible

Checks if two types can be compared:

```go
func typesCompatible(left, right reflect.Kind) bool {
	if left == right {
		return true
	}
	
	// Strings only with strings
	if left == reflect.String || right == reflect.String {
		return left == right
	}
	
	// Bools only with bools
	if left == reflect.Bool || right == reflect.Bool {
		return left == right
	}
	
	// All numerics are compatible
	if isNumericKind(left) && isNumericKind(right) {
		return true
	}
	
	return false
}
```

---

## Design Patterns Used

### 1. Visitor Pattern

The validator "visits" each node in the AST tree:

```go
func (v *Validator) validateNode(node ast.Node) []error {
	switch n := node.(type) {
	case *ast.LogicalExpression:
		return v.validateLogical(n)
	case *ast.ComparisonExpression:
		return v.validateComparison(n)
	// ... handle each node type
	}
}
```

### 2. Error Collection

Instead of failing fast, collect all errors:

```go
var errors []error

// Validate left
leftErrs := v.validateNode(expr.Left)
errors = append(errors, leftErrs...)

// Validate right (even if left had errors!)
rightErrs := v.validateNode(expr.Right)
errors = append(errors, rightErrs...)

return errors
```

### 3. Recursive Validation

Each validation method recursively validates child nodes:

```go
func (v *Validator) validateLogical(expr *ast.LogicalExpression) []error {
	var errors []error
	errors = append(errors, v.validateNode(expr.Left)...)   // Recursive!
	errors = append(errors, v.validateNode(expr.Right)...)  // Recursive!
	return errors
}
```

---

## Testing Strategy

### Unit Tests (Task-Specific)

Each task has dedicated tests:
- Task 1: Simple field existence
- Task 2: Nested traversal
- Task 3: Type compatibility
- Task 4: Array operators
- Task 5: Function calls

### Integration Tests

Test combinations of features:
```go
filter := `name = "John" AND age > 25 AND email.verified = true AND tags:urgent AND timestamp(created_at) > 1000`
```

### Error Collection Tests

Ensure multiple errors are returned:
```go
filter := `unknown_field = "test" AND age = "invalid" AND tags = "wrong_op"`
// Should return 3 errors!
```

---

## Common Pitfalls and Solutions

### Pitfall 1: Not handling anonymous structs

**Problem**:
```go
type User struct {
	Email struct {  // Anonymous struct!
		Address string
	}
}
```

**Solution**: Use `field.Type` to get the anonymous struct type:
```go
field, _ := userType.FieldByName("Email")
emailType := field.Type  // This is the struct type
```

### Pitfall 2: Stopping on first error

**Problem**: Only returning first error found.

**Solution**: Use `append` to collect all errors:
```go
var errors []error
errors = append(errors, v.validateLeft()...)
errors = append(errors, v.validateRight()...)  // Continue even if left had errors
return errors
```

### Pitfall 3: Not validating child nodes

**Problem**: Only validating current node, not its children.

**Solution**: Always recursively validate:
```go
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) []error {
	var errors []error
	
	// ... validate comparison itself
	
	// Don't forget to validate children!
	errors = append(errors, v.validateNode(expr.Left)...)
	errors = append(errors, v.validateNode(expr.Right)...)
	
	return errors
}
```

---

## Performance Considerations

### Reflection Performance

Reflection is slower than direct field access, but:
- Validation typically happens once per request
- The flexibility outweighs the small performance cost
- For high-performance needs, consider caching field lookups

### Optimization Ideas

1. **Cache field lookups**:
```go
type Validator struct {
	structType reflect.Type
	fieldCache map[string]reflect.StructField  // Cache lookups
}
```

2. **Validate once, use many times**:
```go
// Validate filter string once
validator.Validate(filterAST)

// Use same AST for SQL generation, documentation, etc.
```

---

## Extensions and Future Work

### 1. Custom Function Registry Per Resource

```go
userValidator := NewValidator(reflect.TypeOf(User{}))
userValidator.AddFunction("user_active", ...)

bookValidator := NewValidator(reflect.TypeOf(Book{}))
bookValidator.AddFunction("in_stock", ...)
```

### 2. Proto Field Name Mapping

Handle proto snake_case naming:
```go
// Proto: email_address
// Go struct: EmailAddress
// Filter: email_address = "test"  ← Should work!
```

Solution: Check struct tags:
```go
field, _ := structType.FieldByName("EmailAddress")
jsonName := field.Tag.Get("json")  // "email_address"
```

### 3. Position Tracking for Better Errors

Include position in error messages:
```go
"field 'email' does not exist at position 12"
```

### 4. Warning vs Error

Some issues might be warnings, not errors:
```go
// Warning: comparing int64 with float might lose precision
age > 25.5
```

---

## AIP-160 Compliance Summary

### ✅ Fully Implemented Requirements

This validator implements all core AIP-160 schematic validation requirements:

1. **Field Existence** ([AIP-160 §Validation](https://google.aip.dev/160#validation))
   - ✅ "Fields referenced in the filter must exist on the filtered schema"
   - Implementation: Tasks 1-2 validate all field paths using reflection

2. **Type Alignment** ([AIP-160 §Validation](https://google.aip.dev/160#validation))
   - ✅ "Field values provided in the filter must align to the type of the field"
   - ✅ "For example, for a field int32 age a filter like 'age=hello' is invalid"
   - Implementation: Task 3 validates type compatibility for all comparisons

3. **Boolean Operator Restrictions** ([AIP-160 §Comparison Operators](https://google.aip.dev/160#comparison-operators))
   - ✅ "should not provide [<, >, <=, >=] for booleans"
   - Implementation: Task 3 rejects `<`, `>`, `<=`, `>=` on boolean fields

4. **Array/Slice Operators** (AIP-160 has operator)
   - ✅ Enforces has operator (`:`) for array/slice fields
   - Implementation: Task 4 rejects `=` on arrays

5. **Function Call Validation**
   - ✅ Validates function names against supported registry
   - ✅ Validates argument count and types
   - Implementation: Task 5

### ⚠️ Partial Implementation

These AIP-160 features are partially implemented or depend on other components:

1. **Timestamp Validation**
   - **AIP-160**: "Timestamps expect an RFC-3339 formatted string (e.g. 2012-04-21T11:30:00-04:00)"
   - **Our Implementation**: Validates int64 fields (Unix epoch seconds)
   - **Rationale**: Many production APIs use int64 timestamps. RFC-3339 parsing should be in lexer/parser
   - **Full Compliance**: Parse RFC-3339 strings in lexer, validate format

2. **Duration Validation**
   - **AIP-160**: "Durations expect a numeric representation followed by an s suffix (e.g. 20s, 1.2s)"
   - **Our Implementation**: Validates numeric fields that could be durations
   - **Rationale**: Go's `time.Duration` is int64 nanoseconds. Duration literal parsing should be in lexer
   - **Full Compliance**: Parse duration strings (e.g., "30s") in lexer, validate format

3. **Wildcard Support**
   - **AIP-160**: "services should support wildcards using the * character; for example, a = '*.foo'"
   - **Our Implementation**: Parser accepts string literals with '*', validator doesn't validate pattern syntax
   - **Rationale**: Wildcard pattern validation is orthogonal to schema validation
   - **Full Compliance**: Add wildcard pattern syntax validation (e.g., ensure valid glob pattern)

### ❌ Not Implemented (Advanced/Optional)

These AIP-160 features require additional metadata or are language-specific:

1. **Enum Value Validation**
   - **AIP-160**: "Field values for bounded data types e.g. enum provided in the filter must be a valid value in the set"
   - **AIP-160**: Enums should only support `=` and `!=` operators
   - **Issue**: Go doesn't have first-class enum support
   - **Workaround**: Use struct tags or custom metadata to identify enum fields

Example extension for enum support:
```go
type User struct {
    Status int32 `validate:"enum=ACTIVE,INACTIVE,DELETED"`
}

// Custom validator checks tag and validates operator + value
func (v *ExtendedValidator) validateComparison(expr *ast.ComparisonExpression) []error {
    if field, ok := v.structType.FieldByName(identName); ok {
        if enumTag := field.Tag.Get("validate"); strings.HasPrefix(enumTag, "enum=") {
            // Extract valid values
            validValues := strings.Split(strings.TrimPrefix(enumTag, "enum="), ",")
            
            // Check operator (only = and !=)
            if expr.Operator != "=" && expr.Operator != "!=" {
                return []error{fmt.Errorf("enum field only supports = and !=, not '%s'", expr.Operator)}
            }
            
            // Check value is in allowed set
            if !contains(validValues, literalValue) {
                return []error{fmt.Errorf("invalid enum value '%s'", literalValue)}
            }
        }
    }
}
```

2. **Standardized Type Format Validation**
   - **AIP-160**: "Field values for standardized types e.g. Timestamp must conform to the documented standard"
   - **Current**: We validate field types, not value formats
   - **Rationale**: Format validation (e.g., valid RFC-3339, valid duration string) should happen at parse time or execution time

### Implementation Checklist

**For AIP-160 Compliance:**

- [x] Validate field existence (all paths)
- [x] Validate type alignment (field type matches value type)
- [x] Restrict boolean operators to = and !=
- [x] Enforce has operator for arrays/slices
- [x] Validate function calls
- [ ] Enum operator restrictions (requires custom metadata)
- [ ] Enum value validation (requires custom metadata)
- [x] Timestamp field type validation (int64)
- [ ] Timestamp format validation (RFC-3339 - do in lexer)
- [x] Duration field type validation (int64)
- [ ] Duration format validation ("30s" - do in lexer)
- [ ] Wildcard pattern syntax validation (optional)

**Production Recommendations:**

1. **Enum Support**: Add struct tags to identify enum fields and their valid values
2. **Timestamp/Duration**: Parse RFC-3339 and duration strings in the lexer, not validator
3. **Wildcards**: Add pattern validation if wildcard support is critical for your API
4. **Error Codes**: Return structured errors that map to gRPC `INVALID_ARGUMENT` status

---

## Congratulations! 🎉

You've built a complete schema validator that:
- ✅ Validates field existence using reflection
- ✅ Handles nested field traversal
- ✅ Checks type compatibility
- ✅ Enforces correct operators for arrays
- ✅ Validates function calls
- ✅ Collects and returns all errors

This validator is production-ready and can be used to validate user filters before execution!
