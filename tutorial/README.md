# AIP-160 Filter Implementation Tutorial

Welcome to the AIP-160 Filter Implementation Tutorial! This hands-on guide will teach you how to build a production-grade filter parser and evaluator from scratch.

## 🎯 Learning Objectives

By completing this tutorial, you will:

1. ✅ Understand how parsers work (lexing → parsing → evaluation)
2. ✅ Read and implement EBNF grammar specifications
3. ✅ Design efficient data structures (ASTs) for representing parsed expressions
4. ✅ Build a recursive descent parser by hand
5. ✅ Create a scalable, production-ready Go package
6. ✅ Master test-driven development practices

## 📚 Tutorial Modules

### Module 1: Lexer (Tokenization) - ~2 hours
**Goal**: Convert filter strings into tokens

Learn how to break down a filter string like `name = "John" AND age > 25` into individual tokens. You'll implement a lexer (tokenizer) that recognizes keywords, operators, literals, and identifiers.

- 📖 [Start Module 1](module1-lexer/README.md)

### Module 2: Abstract Syntax Trees (AST) - ~1.5 hours
**Goal**: Design data structures for parsed filters

Learn how to represent filter expressions as tree structures in memory. You'll design Go types for different expression nodes (comparisons, logical operators, etc.).

- 📖 [Start Module 2](module2-ast/README.md)

### Module 3: EBNF & Parser Design - ~2 hours
**Goal**: Understand the AIP-160 grammar and plan the parser

Learn to read EBNF grammar notation and translate grammar rules into parsing functions. You'll map the AIP-160 spec to code structure.

- 📖 [Start Module 3](module3-parser-design/README.md)

### Module 4: Building the Parser (Core) - ~2 hours
**Goal**: Implement parsing for core operators

Write the parser functions for literals, comparisons, and logical operators. Handle operator precedence correctly (remember: OR before AND!).

- 📖 Coming after Module 3

### Module 5: Advanced Features - ~1.5 hours
**Goal**: Add traversal, has operator, and wildcards

Implement the more complex parts of AIP-160: field traversal (`a.b.c`), has operator (`list:value`), and wildcard matching.

- 📖 Coming after Module 4

### Module 6: Evaluation Engine - ~2 hours
**Goal**: Execute filters against real data

Build the evaluation engine that takes your AST and runs it against Go structs/maps to determine if they match the filter.

- 📖 Coming after Module 5

## 🎓 How to Use This Tutorial

### 1. Read Each Module's README
Each module has detailed explanations with:
- 📊 Visual diagrams
- 💡 Key concepts
- 🎯 Clear objectives
- ❓ Questions to deepen understanding

### 2. Implement to Pass Tests
Each module provides test files. Your job:
```bash
# Run tests (they'll fail initially)
go test ./pkg/filter/lexer

# Implement the code
# ... write your implementation ...

# Run tests again (until they pass!)
go test ./pkg/filter/lexer -v
```

### 3. Use Hints and Solutions
Stuck? Each module includes:
- 💭 **Hints**: Nudge you in the right direction
- ✅ **Solutions**: Complete working code with explanations
- 🤔 **Challenges**: Optional exercises to deepen learning

### 4. Commit Your Progress
```bash
# After completing each module
git add .
git commit -m "Completed Module 1: Lexer"
```

## 📋 Prerequisites

- **Go Experience**: Advanced level (comfortable with structs, interfaces, methods)
- **No Parser Experience**: We'll teach you everything about parsers!
- **Time**: Plan for 10-12 hours total across several sessions

## 🚀 Getting Started

1. **Set up your environment**:
   ```bash
   # Verify Go installation
   go version  # Should be 1.22+
   
   # Initialize git (if not done)
   git init
   git add .
   git commit -m "Initial commit: Tutorial setup"
   ```

2. **Start with Module 1**:
   ```bash
   cd tutorial/module1-lexer
   cat README.md
   ```

3. **Run the first tests**:
   ```bash
   go test ./pkg/filter/lexer
   ```
   
   You'll see failing tests - that's perfect! Your job is to make them pass.

## 💡 Learning Tips

- **Take Your Time**: Understanding is more important than speed
- **Draw Diagrams**: Visual learners - sketch out token flows and tree structures
- **Experiment**: Try breaking things to understand how they work
- **Ask Questions**: The tutorial includes reflection questions - think through them
- **Test Frequently**: Run tests after small changes to get quick feedback

## 🎯 Success Criteria

You'll know you've mastered this when you can:
- ✅ Explain how a filter string becomes executable code
- ✅ Design AST nodes for new filter features
- ✅ Read EBNF grammar and implement it
- ✅ Debug parser issues confidently
- ✅ Use your package in production at your company

## 📖 Additional Resources

- [AIP-160 Specification](https://google.aip.dev/160)
- [EBNF Grammar for AIP-160](https://google.aip.dev/assets/misc/ebnf-filtering.txt)
- [Recursive Descent Parsing](https://en.wikipedia.org/wiki/Recursive_descent_parser)

---

**Ready to begin?** → [Start Module 1: Lexer](module1-lexer/README.md)
