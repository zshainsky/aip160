# AIP-160 Filter Implementation Tutorial

A comprehensive, test-driven tutorial for learning and implementing Google's [AIP-160 Filtering specification](https://google.aip.dev/160) in Go.

## What You'll Learn

- **Lexical Analysis**: Building a tokenizer to convert filter strings into tokens
- **Abstract Syntax Trees (AST)**: Designing data structures to represent parsed filters
- **EBNF Grammars**: Understanding and implementing formal language specifications
- **Recursive Descent Parsing**: Hand-writing a parser from grammar rules
- **Schema Validation**: Validating filters against data schemas
- **Protobuf Integration**: Using protoreflect for high-performance validation

## What You'll Build

A production-ready, scalable Go package that implements AIP-160 filtering, suitable for use in real-world APIs.

## Project Structure

```
aip160/
├── tutorial/           # Learning materials with guided modules
├── pkg/filter/         # The actual implementation package
│   ├── lexer/         # Tokenization
│   ├── ast/           # Abstract Syntax Tree definitions
│   ├── parser/        # Parser implementation
│   └── validator/     # Schema validation
│       ├── validator.go           # Reflection-based validator
│       ├── proto_validator.go     # Protobuf-based validator
│       └── PROTO_VALIDATOR.md     # ProtoValidator guide
└── examples/          # Usage examples
```

## Getting Started

1. **Prerequisites**: 
   - Go 1.22 or higher
   - Familiarity with Go (advanced level)
   - No prior parser/grammar experience needed!

2. **Start the Tutorial**:
   ```bash
   cd tutorial
   cat README.md
   ```

3. **Follow the Modules** in order:
   - Module 1: Lexer (Tokenization)
   - Module 2: AST Design
   - Module 3: EBNF & Parser Design
   - Module 4: Parser Core
   - Module 5: Advanced Features
   - Module 6: Evaluator

## Development Approach

This tutorial uses **Test-Driven Development (TDD)**:
- Each module provides test cases
- You implement the code to make tests pass
- Solutions and hints are available when needed

## Usage Examples

Once you've completed the tutorial, you can use the AIP-160 filter package in your applications.

### Quick Start: ProtoValidator (Recommended)

**For protobuf-based APIs**, use `ProtoValidator` for 2-5x better performance:

```go
package main

import (
    "fmt"
    "github.com/zshainsky/aip160/pkg/filter/parser"
    "github.com/zshainsky/aip160/pkg/filter/validator"
)

func main() {
    // Get message descriptor from your proto-generated type
    msgDesc := (&pb.User{}).ProtoReflect().Descriptor()
    
    // Create validator
    v := validator.NewProtoValidator(msgDesc)
    
    // Parse and validate filter
    expr, _ := parser.ParseFilter(`age > 25 AND status = "ACTIVE"`)
    errs := v.Validate(expr)
    
    if len(errs) > 0 {
        fmt.Printf("Validation errors: %v\n", errs)
        return
    }
    
    fmt.Println("Filter is valid!")
}
```

**Features:**
- ✅ 2-5x faster than reflection-based validation
- ✅ Native enum support with prefix stripping
- ✅ HAS operator for repeated fields (`tags:"urgent"`)
- ✅ Nested message traversal
- ✅ Type-safe validation for all proto types

📖 **[Read the ProtoValidator Guide](pkg/filter/validator/PROTO_VALIDATOR.md)** for detailed documentation.

### Using Reflection-Based Validator

**For regular Go structs**, use the reflection-based validator:

```go
package main

import (
	"fmt"
	"log"
	"reflect"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator"
)

// Define your data model
type User struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

func main() {
	// Parse filter string
	filterString := `name = "John" AND age > 25`
	expr, err := parser.ParseFilter(filterString)
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}
	
	// Validate against struct
	v := validator.NewValidator(reflect.TypeOf(User{}), validator.WithJSONTags())
	errors := v.Validate(expr)
	
	if len(errors) > 0 {
		log.Fatalf("Validation errors: %v", errors)
	}
	
	fmt.Println("Filter is valid!")
}
```

### Validator Comparison

| Feature | ProtoValidator | Reflection Validator |
|---------|---------------|-------------------|
| **Performance** | 2-5x faster (O(1) lookups) | Baseline (O(n) iteration) |
| **Data Model** | Protobuf (`.proto` files) | Go structs |
| **Enum Support** | ✅ Native with prefix stripping | ⚠️ Limited (string comparison) |
| **Type Safety** | ✅ Strong proto types | ⚠️ Go reflection |
| **Repeated Fields** | ✅ HAS operator (`:`) | ❌ Not supported |
| **Nested Messages** | ✅ Unlimited depth | ✅ Unlimited depth |
| **Struct Tags** | ❌ Not applicable | ✅ `json`, `filter` tags |
| **Setup** | Requires `.proto` + codegen | Just Go structs |

**Choose ProtoValidator if:**
- You have `.proto` files and generated `*.pb.go` files
- Performance matters (high-throughput APIs)
- You need enum validation or HAS operator support

**Choose Reflection Validator if:**
- You use plain Go structs (no protobuf)
- You need struct tag support
- Simplicity is more important than performance

## Running Tests

```bash
# Run all tests
go test ./...

# Run validator tests
go test ./pkg/filter/validator

# Run tests with coverage
go test -cover ./pkg/filter/validator
# Output: coverage: 83.1% of statements

# Run tests with verbose output
go test -v ./pkg/filter/validator
```

## License

MIT License - feel free to use this tutorial and code for learning and production use.

## Contributing

This tutorial is designed to be reusable. If you find improvements or issues, contributions are welcome!
