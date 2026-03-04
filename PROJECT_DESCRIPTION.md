# Project Description
I want to work on a project together. The idea is to create a coding tutorial to help me learn the AIP, EBNF Grammer, and walk away with an efficient and scaleable Golang package solution for implementing the AIP-160.

# Project Goals
1. Finish tutorial produces an efficient and scaleable Golang package implementing AIP-160
2. Turorial modules should be TDD because this is a great way to learn and the package must have comprehensive test condtions
3. Tutorial shouldn't be reasonable lenght something one can complete in several 2 hr sittings
4. The student of the tutorial should walk away from the tutorial understanding the datastructures required for impelemtning this solution and the algorithm behind parsing the input strings
5. The tutorial should ask the student questions or propose coding challenges to help learn. There should be answers available if needed and hints along the way. 
6. The student can submit code to determine if their solution works.

# Technical Documentation
AIP-160 Full Definition: https://google.aip.dev/160

Summary of AIP-160 (genrated by copilot)
```
AIP-160: Filtering defines a common specification for filtering collections in List methods using a structured string-based syntax accessible to non-technical audiences.

Key Features:

Literals: Bare values matched against fields. Whitespace-separated literals imply fuzzy AND (e.g., Victor Hugo ≈ Victor AND Hugo)

Logical Operators:

AND, OR (OR has higher precedence than AND, unlike most programming languages)
Negation: NOT and - (interchangeable)
Comparison Operators: =, !=, <, >, <=, >=

Field names must be on the left side
Supports wildcards (*) for string equality
Type conversion for enums, booleans, numbers, durations (e.g., 20s), timestamps (RFC-3339)
Traversal Operator (.): Navigate through messages, maps, structs

Example: a.b.c = "foo"
If any non-primitive field in chain is unset, the entry is skipped
Cannot traverse repeated fields except with : operator
Has Operator (:): Check presence/values in collections

Repeated fields: r:42 (contains 42), r.foo:42 (contains element where foo=42)
Maps/messages: m:foo (has key), m.foo:* (has key), m.foo:42 (specific value)
Functions: API-specific extensions using call(arg...) syntax

Validation: Invalid filters return INVALID_ARGUMENT

The formal syntax is defined in an EBNF grammar. Services may impose additional limitations but must document them clearly.
```

# Working together on this project
- I want to make sure this is collaborative as we build this tutorial together. Please ask questions as we go. The more questions the more reinforced and better the tutorial will be.
- I am building this project so i can learn
- I am building this project so i can have a solution to bring to my company that functions as intended and is a scaleable solution