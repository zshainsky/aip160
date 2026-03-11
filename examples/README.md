# Examples

This directory contains examples of using the AIP-160 filter package.

## Basic Lexer Usage

Once you've completed Module 1, you can use the lexer like this:

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	input := `name = "John" AND age > 25`
	
	l := lexer.New(input)
	
	for {
		tok := l.NextToken()
		fmt.Printf("%s\n", tok)
		
		if tok.Type == lexer.EOF {
			break
		}
	}
}
```

**Output:**
```
IDENTIFIER(name)
=(=)
STRING(John)
AND(AND)
IDENTIFIER(age)
>(>)
NUMBER(25)
EOF()
```

## Working with AST Nodes (Module 2)

After completing Module 2, you can create and manipulate AST nodes:

### Example 1: Simple Literal Nodes

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	// Create a number literal
	num := &ast.NumberLiteral{
		Token: lexer.Token{Type: lexer.NUMBER, Literal: "42"},
		Value: 42,
	}
	fmt.Println("Number:", num.String()) // Output: 42
	
	// Create a string literal
	str := &ast.StringLiteral{
		Token: lexer.Token{Type: lexer.STRING, Literal: "John"},
		Value: "John",
	}
	fmt.Println("String:", str.String()) // Output: "John"
	
	// Create an identifier
	ident := &ast.Identifier{
		Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
		Value: "age",
	}
	fmt.Println("Identifier:", ident.String()) // Output: age
	
	// Create boolean and null literals
	trueLit := &ast.BooleanLiteral{
		Token: lexer.Token{Type: lexer.TRUE, Literal: "true"},
		Value: true,
	}
	fmt.Println("Boolean:", trueLit.String()) // Output: true
	
	nullLit := &ast.NullLiteral{
		Token: lexer.Token{Type: lexer.NULL, Literal: "null"},
	}
	fmt.Println("Null:", nullLit.String()) // Output: null
}
```

### Example 2: Building a Comparison Expression

Create an expression tree for `age > 18`:

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	// Build: age > 18
	comparison := &ast.ComparisonExpression{
		Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
		Left: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
			Value: "age",
		},
		Operator: ">",
		Right: &ast.NumberLiteral{
			Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
			Value: 18,
		},
	}
	
	fmt.Println(comparison.String())
	// Output: (age > 18)
}
```

### Example 3: Building a Logical Expression

Create a tree for `age > 18 AND status = "active"`:

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	// Build: age > 18 AND status = "active"
	expr := &ast.LogicalExpression{
		Token: lexer.Token{Type: lexer.AND, Literal: "AND"},
		Left: &ast.ComparisonExpression{
			Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
			Left: &ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
				Value: "age",
			},
			Operator: ">",
			Right: &ast.NumberLiteral{
				Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
				Value: 18,
			},
		},
		Operator: "AND",
		Right: &ast.ComparisonExpression{
			Token: lexer.Token{Type: lexer.EQUALS, Literal: "="},
			Left: &ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "status"},
				Value: "status",
			},
			Operator: "=",
			Right: &ast.StringLiteral{
				Token: lexer.Token{Type: lexer.STRING, Literal: "active"},
				Value: "active",
			},
		},
	}
	
	fmt.Println(expr.String())
	// Output: ((age > 18) AND (status = "active"))
}
```

### Example 4: Traversal and Has Expressions

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	// Build: user.email
	traversal := &ast.TraversalExpression{
		Token: lexer.Token{Type: lexer.DOT, Literal: "."},
		Left: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "user"},
			Value: "user",
		},
		Right: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "email"},
			Value: "email",
		},
	}
	fmt.Println("Traversal:", traversal.String())
	// Output: user.email
	
	// Build: tags:urgent
	has := &ast.HasExpression{
		Token: lexer.Token{Type: lexer.HAS, Literal: ":"},
		Collection: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "tags"},
			Value: "tags",
		},
		Member: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "urgent"},
			Value: "urgent",
		},
	}
	fmt.Println("Has:", has.String())
	// Output: tags:urgent
	
	// Build: NOT active
	unary := &ast.UnaryExpression{
		Token: lexer.Token{Type: lexer.NOT, Literal: "NOT"},
		Operator: "NOT",
		Right: &ast.Identifier{
			Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "active"},
			Value: "active",
		},
	}
	fmt.Println("Unary:", unary.String())
	// Output: (NOT active)
}
```

### Example 5: Complex Nested Expression

Build a tree for `(age > 18 AND status = "active") OR role = "admin"`:

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	// Build complex nested expression
	expr := &ast.LogicalExpression{
		Token: lexer.Token{Type: lexer.OR, Literal: "OR"},
		// Left side: (age > 18 AND status = "active")
		Left: &ast.LogicalExpression{
			Token: lexer.Token{Type: lexer.AND, Literal: "AND"},
			Left: &ast.ComparisonExpression{
				Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
				Left: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
					Value: "age",
				},
				Operator: ">",
				Right: &ast.NumberLiteral{
					Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
					Value: 18,
				},
			},
			Operator: "AND",
			Right: &ast.ComparisonExpression{
				Token: lexer.Token{Type: lexer.EQUALS, Literal: "="},
				Left: &ast.Identifier{
					Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "status"},
					Value: "status",
				},
				Operator: "=",
				Right: &ast.StringLiteral{
					Token: lexer.Token{Type: lexer.STRING, Literal: "active"},
					Value: "active",
				},
			},
		},
		Operator: "OR",
		// Right side: role = "admin"
		Right: &ast.ComparisonExpression{
			Token: lexer.Token{Type: lexer.EQUALS, Literal: "="},
			Left: &ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "role"},
				Value: "role",
			},
			Operator: "=",
			Right: &ast.StringLiteral{
				Token: lexer.Token{Type: lexer.STRING, Literal: "admin"},
				Value: "admin",
			},
		},
	}
	
	fmt.Println(expr.String())
	// Output: (((age > 18) AND (status = "active")) OR (role = "admin"))
	
	// Notice how the tree structure naturally represents precedence!
	// The AND operation is nested deeper, so it gets evaluated first.
}
```

### Example 6: Function Calls

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	// Build: timestamp(created_at)
	singleArg := &ast.FunctionCall{
		Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "timestamp"},
		Function: "timestamp",
		Arguments: []ast.Expression{
			&ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "created_at"},
				Value: "created_at",
			},
		},
	}
	fmt.Println("Single arg:", singleArg.String())
	// Output: timestamp(created_at)
	
	// Build: duration(start, end)
	multipleArgs := &ast.FunctionCall{
		Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "duration"},
		Function: "duration",
		Arguments: []ast.Expression{
			&ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "start"},
				Value: "start",
			},
			&ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "end"},
				Value: "end",
			},
		},
	}
	fmt.Println("Multiple args:", multipleArgs.String())
	// Output: duration(start, end)
	
	// Build: now()
	noArgs := &ast.FunctionCall{
		Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "now"},
		Function: "now",
		Arguments: []ast.Expression{},
	}
	fmt.Println("No args:", noArgs.String())
	// Output: now()
}
```

### Example 7: Using Program Root Node

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/ast"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
)

func main() {
	// The Program node is the root of the AST
	program := &ast.Program{
		Expression: &ast.ComparisonExpression{
			Token: lexer.Token{Type: lexer.GREATER_THAN, Literal: ">"},
			Left: &ast.Identifier{
				Token: lexer.Token{Type: lexer.IDENTIFIER, Literal: "age"},
				Value: "age",
			},
			Operator: ">",
			Right: &ast.NumberLiteral{
				Token: lexer.Token{Type: lexer.NUMBER, Literal: "18"},
				Value: 18,
			},
		},
	}
	
	fmt.Println("Program:", program.String())
	// Output: (age > 18)
	
	fmt.Println("Token:", program.TokenLiteral())
	// Output: >
}
```

## Parser Usage (Modules 3-5)

After completing the parser modules, you can parse complete filter expressions.

### Example 8: Basic Parsing

```go
package main

import (
	"fmt"
	"log"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
)

func main() {
	// Parse a simple comparison
	input := `name = "Alice"`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	// Check for parsing errors
	if len(p.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p.Errors())
	}
	
	fmt.Println("Parsed:", program.String())
	// Output: Parsed: (name = "Alice")
}
```

### Example 9: Parsing Complex Expressions

```go
package main

import (
	"fmt"
	"log"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
)

func main() {
	// Parse complex expression with multiple operators
	input := `(age >= 18 AND age <= 65) OR status = "exempt"`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p.Errors())
	}
	
	fmt.Println("Parsed:", program.String())
	// Output: Parsed: (((age >= 18) AND (age <= 65)) OR (status = "exempt"))
	
	// The AST respects precedence
	// AND has higher precedence than OR (note how it's nested deeper)
}
```

### Example 10: Parsing with Traversal and Has

```go
package main

import (
	"fmt"
	"log"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
)

func main() {
	// Traversal operator for nested fields
	input := `user.profile.email = "alice@example.com"`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p.Errors())
	}
	
	fmt.Println("Traversal:", program.String())
	// Output: Traversal: (user.profile.email = "alice@example.com")
	
	// Has operator for collection membership
	input2 := `tags:"urgent" AND categories:"tech"`
	l2 := lexer.New(input2)
	p2 := parser.New(l2)
	program2 := p2.ParseProgram()
	
	if len(p2.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p2.Errors())
	}
	
	fmt.Println("Has:", program2.String())
	// Output: Has: (tags:urgent AND categories:tech)
}
```

### Example 11: Parsing with Negation

```go
package main

import (
	"fmt"
	"log"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
)

func main() {
	// NOT operator
	input := `NOT deleted AND active = true`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p.Errors())
	}
	
	fmt.Println("NOT:", program.String())
	// Output: NOT: ((NOT deleted) AND (active = true))
	
	// Alternative negation with minus
	input2 := `-archived AND status != "expired"`
	l2 := lexer.New(input2)
	p2 := parser.New(l2)
	program2 := p2.ParseProgram()
	
	if len(p2.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p2.Errors())
	}
	
	fmt.Println("Minus:", program2.String())
	// Output: Minus: ((- archived) AND (status != "expired"))
}
```

### Example 12: Parsing Function Calls

```go
package main

import (
	"fmt"
	"log"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
)

func main() {
	// Function with single argument
	input := `timestamp(created_at) > timestamp("2024-01-01T00:00:00Z")`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p.Errors())
	}
	
	fmt.Println("Function:", program.String())
	// Output: Function: (timestamp(created_at) > timestamp("2024-01-01T00:00:00Z"))
	
	// Function with multiple arguments
	input2 := `duration(start_time, end_time) > duration("1h30m")`
	l2 := lexer.New(input2)
	p2 := parser.New(l2)
	program2 := p2.ParseProgram()
	
	if len(p2.Errors()) > 0 {
		log.Fatalf("Parser errors: %v", p2.Errors())
	}
	
	fmt.Println("Multi-arg:", program2.String())
	// Output: Multi-arg: (duration(start_time, end_time) > duration("1h30m"))
}
```

### Example 13: Error Handling

```go
package main

import (
	"fmt"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
)

func main() {
	// Invalid syntax - missing value
	input := `name = `
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	if len(p.Errors()) > 0 {
		fmt.Println("Parser detected errors:")
		for _, err := range p.Errors() {
			fmt.Printf("  - %s\n", err)
		}
	}
	
	// Invalid syntax - unmatched parentheses
	input2 := `(age > 18 AND status = "active"`
	l2 := lexer.New(input2)
	p2 := parser.New(l2)
	program2 := p2.ParseProgram()
	
	if len(p2.Errors()) > 0 {
		fmt.Println("\nMore errors:")
		for _, err := range p2.Errors() {
			fmt.Printf("  - %s\n", err)
		}
	}
}
```

## Validator Usage (Module 6)

The validator checks that parsed filters reference valid fields in your data model.

### Example 14: Basic Validation

```go
package main

import (
	"fmt"
	"reflect"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator"
)

type User struct {
	Name   string `json:"name"`
	Age    int    `json:"age"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

func main() {
	// Valid filter
	input := `age > 18 AND active = true`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	// Create validator using JSON tags
	v := validator.NewValidator(reflect.TypeOf(User{}), validator.WithJSONTags())
	errors := v.Validate(program)
	
	if len(errors) == 0 {
		fmt.Println("✓ Filter is valid")
	} else {
		fmt.Println("✗ Validation errors:", errors)
	}
}
```

### Example 15: Validation with Different Tag Types

```go
package main

import (
	"fmt"
	"reflect"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator"
)

type Product struct {
	Name     string  `json:"product_name" protobuf:"bytes,1,opt,name=name,proto3"`
	Price    float64 `json:"price" protobuf:"fixed64,2,opt,name=price,proto3"`
	InStock  bool    `json:"in_stock" protobuf:"varint,3,opt,name=in_stock,proto3"`
	Category string  `json:"category" protobuf:"bytes,4,opt,name=category,proto3"`
}

func main() {
	input := `product_name = "Widget"`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	// Validate using JSON tags
	jsonValidator := validator.NewValidator(reflect.TypeOf(Product{}), validator.WithJSONTags())
	errors := jsonValidator.Validate(program)
	fmt.Println("JSON tags validation:", len(errors) == 0)
	
	// Validate using protobuf tags
	input2 := `name = "Widget"`
	l2 := lexer.New(input2)
	p2 := parser.New(l2)
	program2 := p2.ParseProgram()
	
	protoValidator := validator.NewValidator(reflect.TypeOf(Product{}), validator.WithProtobufTags())
	errors2 := protoValidator.Validate(program2)
	fmt.Println("Protobuf tags validation:", len(errors2) == 0)
	
	// Default: validate using struct field names (PascalCase)
	input3 := `Name = "Widget"`
	l3 := lexer.New(input3)
	p3 := parser.New(l3)
	program3 := p3.ParseProgram()
	
	defaultValidator := validator.NewValidator(reflect.TypeOf(Product{}))
	errors3 := defaultValidator.Validate(program3)
	fmt.Println("Default validation:", len(errors3) == 0)
}
```

### Example 16: Catching Validation Errors

```go
package main

import (
	"fmt"
	"reflect"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator"
)

type Article struct {
	Title     string   `json:"title"`
	Author    string   `json:"author"`
	Published bool     `json:"published"`
	Tags      []string `json:"tags"`
}

func main() {
	// This filter references a non-existent field
	input := `rating > 4.5 AND published = true`
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()
	
	v := validator.NewValidator(reflect.TypeOf(Article{}), validator.WithJSONTags())
	errors := v.Validate(program)
	
	if len(errors) > 0 {
		fmt.Println("Validation failed:")
		for i, err := range errors {
			fmt.Printf("  %d. %s\n", i+1, err)
		}
		// Output:
		// Validation failed:
		//   1. field 'rating' not found in struct
	}
}
```

### Example 17: Complete End-to-End Example

```go
package main

import (
	"fmt"
	"log"
	"reflect"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator"
)

type Employee struct {
	Name       string  `json:"name"`
	Department string  `json:"department"`
	Salary     float64 `json:"salary"`
	Active     bool    `json:"active"`
	YearsExp   int     `json:"years_experience"`
}

func validateFilter(filterString string) error {
	// Step 1: Tokenize
	l := lexer.New(filterString)
	
	// Step 2: Parse
	p := parser.New(l)
	program := p.ParseProgram()
	
	// Check parsing errors
	if len(p.Errors()) > 0 {
		return fmt.Errorf("parsing failed: %v", p.Errors())
	}
	
	// Step 3: Validate
	v := validator.NewValidator(reflect.TypeOf(Employee{}), validator.WithJSONTags())
	validationErrors := v.Validate(program)
	
	if len(validationErrors) > 0 {
		return fmt.Errorf("validation failed: %v", validationErrors)
	}
	
	fmt.Printf("✓ Valid filter: %s\n", program.String())
	return nil
}

func main() {
	// Test various filters
	filters := []string{
		`department = "Engineering" AND active = true`,
		`salary > 100000 AND years_experience >= 5`,
		`name = "John" OR department = "Sales"`,
		`invalid_field = "test"`, // This will fail validation
	}
	
	for i, filter := range filters {
		fmt.Printf("\nFilter %d: %s\n", i+1, filter)
		if err := validateFilter(filter); err != nil {
			log.Printf("Error: %v\n", err)
		}
	}
}
```

## Real-World API Integration Example

### Example 18: Using in an HTTP API

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
	"github.com/zshainsky/aip160/pkg/filter/validator"
)

type Book struct {
	Title     string  `json:"title"`
	Author    string  `json:"author"`
	ISBN      string  `json:"isbn"`
	Price     float64 `json:"price"`
	Available bool    `json:"available"`
}

func listBooksHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter from query parameter
	filterString := r.URL.Query().Get("filter")
	
	if filterString != "" {
		// Validate the filter
		l := lexer.New(filterString)
		p := parser.New(l)
		program := p.ParseProgram()
		
		// Check for parsing errors
		if len(p.Errors()) > 0 {
			http.Error(w, fmt.Sprintf("Invalid filter syntax: %v", p.Errors()), http.StatusBadRequest)
			return
		}
		
		// Validate against Book schema
		v := validator.NewValidator(reflect.TypeOf(Book{}), validator.WithJSONTags())
		validationErrors := v.Validate(program)
		
		if len(validationErrors) > 0 {
			http.Error(w, fmt.Sprintf("Invalid filter fields: %v", validationErrors), http.StatusBadRequest)
			return
		}
		
		// Filter is valid - use it to query your database
		// (evaluation implementation would go here)
		fmt.Printf("Applying filter: %s\n", program.String())
	}
	
	// Return filtered results
	books := []Book{
		{Title: "Go Programming", Author: "John Doe", ISBN: "123", Price: 29.99, Available: true},
		// ... more books
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(books)
}

func main() {
	http.HandleFunc("/api/books", listBooksHandler)
	fmt.Println("Server starting on :8080")
	fmt.Println("Try: http://localhost:8080/api/books?filter=price<30%20AND%20available=true")
	http.ListenAndServe(":8080", nil)
}
```

## More Examples Coming

As you complete each module, more examples will be added here:

- Module 6: Full filter evaluation with real data
- Advanced patterns and best practices
- Performance optimization tips

Stay tuned!
