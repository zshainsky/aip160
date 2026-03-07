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

## More Examples Coming

As you complete each module, more examples will be added here:

- Module 3: Understanding EBNF grammar
- Module 4: Parsing filter expressions
- Module 5: Advanced parsing features
- Module 6: Full filter evaluation

Stay tuned!
