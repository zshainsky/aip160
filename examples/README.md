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

## More Examples Coming

As you complete each module, more examples will be added here:

- Module 2: Working with AST nodes
- Module 3: Parsing filter expressions
- Module 4: Complex filter examples
- Module 5: Traversal and has operators
- Module 6: Full filter evaluation

Stay tuned!
