# Module 6: Schema Validation - Hints

Need a nudge in the right direction? Here are some hints for each task.

## 💭 Task 1: Field Existence Validation

**Stuck on: How to walk the AST?**
- Use type assertions to check what kind of node you're looking at
- Use a recursive approach: validate the current node, then validate its children
- Pattern: `switch node := n.(type) { case *ast.Program: ... }`

**Stuck on: How to use reflection?**
```go
// Get the struct type
structType := reflect.TypeOf(YourStruct{})

// Check if a field exists
field, ok := structType.FieldByName("FieldName")
if !ok {
    // Field doesn't exist!
}

// Get the field's type
fieldType := field.Type
fieldKind := field.Type.Kind() // string, int, struct, etc.
```

**Stuck on: Where to validate identifiers?**
- Identifiers appear in `ComparisonExpression`, `TraversalExpression`, and `HasExpression`
- Start by handling simple identifiers in comparisons
- Example: `name = "John"` → validate "name" exists

**Stuck on: How to collect multiple errors?**
```go
var errors []error

// As you validate, append errors
if err := v.validateSomething(); err != nil {
    errors = append(errors, err)
}

// Keep going! Don't return early

return errors
```

---

## 💭 Task 2: Nested Field Traversal

**Stuck on: How to handle TraversalExpression?**
- A `TraversalExpression` has a base identifier and segments
- Example: `email.address` → base: "email", segments: ["address"]
- Walk each segment, updating the current type as you go

**Stuck on: How to get the type of a nested field?**
```go
// Start with the struct type
currentType := v.structType

// First segment: "email"
emailField, _ := currentType.FieldByName("email")
currentType = emailField.Type  // Now currentType is the Email struct type

// Second segment: "address"
addressField, _ := currentType.FieldByName("address")
currentType = addressField.Type  // Now currentType is string
```

**Stuck on: What if the traversal base is already a TraversalExpression?**
- This shouldn't happen with the current parser
- TraversalExpression.Base is always an Identifier
- The segments list contains all the nested parts

---

## 💭 Task 3: Type Compatibility Checking

**Stuck on: How to get the type of a literal?**
```go
func getLiteralType(node ast.Node) reflect.Kind {
    switch node.(type) {
    case *ast.StringLiteral:
        return reflect.String
    case *ast.IntLiteral:
        return reflect.Int
    case *ast.BoolLiteral:
        return reflect.Bool
    case *ast.FloatLiteral:
        return reflect.Float64
    }
    return reflect.Invalid
}
```

**Stuck on: What types are compatible?**
- Strings only compatible with strings
- Booleans only compatible with booleans
- All numeric types are compatible with each other:
  - int, int32, int64
  - float32, float64
  - uint, uint32, uint64

**Stuck on: How to handle comparison expressions?**
```go
func (v *Validator) validateComparison(expr *ast.ComparisonExpression) error {
    // Get left side type
    leftType := v.getExpressionType(expr.Left)
    
    // Get right side type
    rightType := v.getExpressionType(expr.Right)
    
    // Compare them
    if !typesCompatible(leftType.Kind(), rightType) {
        return fmt.Errorf("type mismatch")
    }
    
    return nil
}
```

**Stuck on: getExpressionType for different node types?**
- `Identifier` → look up in struct, return field type
- `TraversalExpression` → walk the path, return final type
- `StringLiteral` → return reflect.String
- `IntLiteral` → return reflect.Int
- etc.

---

## 💭 Task 4: Array Field Validation

**Stuck on: How to detect array/slice fields?**
```go
field, _ := structType.FieldByName("Tags")
if field.Type.Kind() == reflect.Slice {
    // It's a slice!
    elementType := field.Type.Elem()  // Type of individual elements
}
```

**Stuck on: Where to validate has expressions?**
- Handle `*ast.HasExpression` in your node visitor
- Check that the collection is an array/slice type
- Validate the value type matches the element type

**Stuck on: How to reject equality on arrays?**
- In your `validateComparison` method
- After getting the left type, check if it's an array/slice
- If so, return an error suggesting the has operator

```go
if leftType.Kind() == reflect.Slice || leftType.Kind() == reflect.Array {
    return fmt.Errorf("use has operator ':' for array fields")
}
```

---

## 💭 Task 5: Function Call Validation

**Stuck on: How to define the function registry?**
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
}
```

**Stuck on: How to validate function calls?**
```go
func (v *Validator) validateFunctionCall(expr *ast.FunctionCall) error {
    // 1. Look up function in registry
    funcDef, ok := SupportedFunctions[expr.Name]
    if !ok {
        return fmt.Errorf("function '%s' not supported", expr.Name)
    }
    
    // 2. Check argument count
    if len(expr.Arguments) != funcDef.ArgCount {
        return fmt.Errorf("wrong number of arguments")
    }
    
    // 3. Validate each argument type
    for i, arg := range expr.Arguments {
        argType := v.getExpressionType(arg)
        // Check if argType.Kind() is in funcDef.ValidFieldKinds
    }
    
    return nil
}
```

**Stuck on: How to check if a kind is in a list?**
```go
func isValidKind(kind reflect.Kind, validKinds []reflect.Kind) bool {
    for _, validKind := range validKinds {
        if kind == validKind {
            return true
        }
    }
    return false
}
```

---

## 🎯 General Hints

### Hint: Error Collection Pattern

```go
func (v *Validator) Validate(node ast.Node) []error {
    var errors []error
    
    // Collect errors as you go
    if err := v.validateNodeType1(node); err != nil {
        errors = append(errors, err)
    }
    
    if err := v.validateNodeType2(node); err != nil {
        errors = append(errors, err)
    }
    
    // Continue validation even if errors found
    // Return all errors at the end
    return errors
}
```

### Hint: Type Switch Pattern

```go
func (v *Validator) validateNode(node ast.Node) []error {
    switch n := node.(type) {
    case *ast.Program:
        return v.validateProgram(n)
    case *ast.LogicalExpression:
        return v.validateLogical(n)
    case *ast.ComparisonExpression:
        return v.validateComparison(n)
    // ... handle each type
    default:
        return nil
    }
}
```

### Hint: Recursive Validation

```go
func (v *Validator) validateLogical(expr *ast.LogicalExpression) []error {
    var errors []error
    
    // Validate left side (recursive!)
    leftErrs := v.validateNode(expr.Left)
    errors = append(errors, leftErrs...)
    
    // Validate right side (recursive!)
    rightErrs := v.validateNode(expr.Right)
    errors = append(errors, rightErrs...)
    
    return errors
}
```

### Hint: Starting Point

Your main validation entry point should probably look like:

```go
func (v *Validator) Validate(node ast.Node) []error {
    // Start with the Program node
    program, ok := node.(*ast.Program)
    if !ok {
        return []error{fmt.Errorf("expected Program node")}
    }
    
    // Validate the expression inside
    return v.validateNode(program.Expression)
}
```

---

Still stuck? Check [SOLUTION.md](SOLUTION.md) for the complete implementation!
