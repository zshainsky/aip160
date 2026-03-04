# AIP-160 Filter Implementation Tutorial

A comprehensive, test-driven tutorial for learning and implementing Google's [AIP-160 Filtering specification](https://google.aip.dev/160) in Go.

## What You'll Learn

- **Lexical Analysis**: Building a tokenizer to convert filter strings into tokens
- **Abstract Syntax Trees (AST)**: Designing data structures to represent parsed filters
- **EBNF Grammars**: Understanding and implementing formal language specifications
- **Recursive Descent Parsing**: Hand-writing a parser from grammar rules
- **Filter Evaluation**: Executing parsed filters against Go data structures

## What You'll Build

A production-ready, scalable Go package that implements AIP-160 filtering, suitable for use in real-world APIs.

## Project Structure

```
aip160tutorial/
├── tutorial/           # Learning materials with guided modules
├── pkg/filter/         # The actual implementation package
│   ├── lexer/         # Tokenization
│   ├── ast/           # Abstract Syntax Tree definitions
│   ├── parser/        # Parser implementation
│   └── eval/          # Filter evaluation engine
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

## Running Tests

```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./pkg/filter/lexer

# Run tests with coverage
go test -cover ./...

# Run tests with verbose output
go test -v ./...
```

## License

MIT License - feel free to use this tutorial and code for learning and production use.

## Contributing

This tutorial is designed to be reusable. If you find improvements or issues, contributions are welcome!
