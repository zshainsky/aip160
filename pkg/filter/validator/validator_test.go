package validator

import (
	"reflect"
	"strings"
	"testing"

	"github.com/zshainsky/aip160/pkg/filter/lexer"
	"github.com/zshainsky/aip160/pkg/filter/parser"
)

// SimpleUser is a basic struct for testing simple field validation
type SimpleUser struct {
	Name   string
	Age    int32
	Active bool
}

// User is a richer struct for testing nested fields, arrays, and complex scenarios
type User struct {
	Name  string
	Email struct {
		Address  string
		Verified bool
	}
	Age       int32
	Tags      []string
	CreatedAt int64
}

// Helper function to parse and validate a filter
func validateFilter(t *testing.T, filterStr string, structType reflect.Type) []error {
	t.Helper()
	l := lexer.New(filterStr)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Failed to parse filter '%s': %v", filterStr, p.Errors())
	}

	validator := NewValidator(structType)
	return validator.Validate(ast)
}

// Task 1: Field Existence Validation

func TestTask1_SimpleFieldExists(t *testing.T) {
	errs := validateFilter(t, `Name = "John"`, reflect.TypeOf(SimpleUser{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
}

func TestTask1_SimpleFieldDoesNotExist(t *testing.T) {
	errs := validateFilter(t, `Email = "test@example.com"`, reflect.TypeOf(SimpleUser{}))
	if len(errs) == 0 {
		t.Error("Expected error for non-existent field 'Email'")
		return
	}
	if !strings.Contains(errs[0].Error(), "Email") {
		t.Errorf("Error should mention 'Email', got: %v", errs[0])
	}
}

func TestTask1_MultipleFields(t *testing.T) {
	errs := validateFilter(t, `Name = "John" AND Age > 25`, reflect.TypeOf(SimpleUser{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
}

func TestTask1_MultipleInvalidFields(t *testing.T) {
	errs := validateFilter(t, `Email = "test" AND Phone = "123"`, reflect.TypeOf(SimpleUser{}))
	if len(errs) < 2 {
		t.Errorf("Expected at least 2 errors (Email and Phone), got %d: %v", len(errs), errs)
	}
}

// Task 2: Nested Field Traversal

func TestTask2_NestedFieldExists(t *testing.T) {
	errs := validateFilter(t, `Email.Address = "test@example.com"`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
}

func TestTask2_NestedFieldDoesNotExist(t *testing.T) {
	errs := validateFilter(t, `Email.Domain = "example.com"`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for non-existent field 'Domain'")
		return
	}
	if !strings.Contains(errs[0].Error(), "Domain") {
		t.Errorf("Error should mention 'Domain', got: %v", errs[0])
	}
}

func TestTask2_DeepNesting(t *testing.T) {
	errs := validateFilter(t, `Email.Verified = true`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors, got: %v", errs)
	}
}

// Task 3: Type Compatibility Checking

func TestTask3_StringComparison(t *testing.T) {
	errs := validateFilter(t, `Name = "John"`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for string comparison, got: %v", errs)
	}
}

func TestTask3_IntComparison(t *testing.T) {
	errs := validateFilter(t, `Age > 25`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for int comparison, got: %v", errs)
	}
}

func TestTask3_TypeMismatch_StringVsInt(t *testing.T) {
	errs := validateFilter(t, `Name > 100`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for comparing string with int")
		return
	}
	errMsg := errs[0].Error()
	if !strings.Contains(errMsg, "type") && !strings.Contains(errMsg, "mismatch") {
		t.Errorf("Error should mention type mismatch, got: %v", errs[0])
	}
}

func TestTask3_TypeMismatch_IntVsString(t *testing.T) {
	errs := validateFilter(t, `Age = "hello"`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for comparing int with string")
	}
}

func TestTask3_NumericCompatibility_Int32VsInt(t *testing.T) {
	errs := validateFilter(t, `Age = 30`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for int32 vs int comparison, got: %v", errs)
	}
}

// AIP-160 Boolean Operator Restrictions

func TestTask3_BooleanComparisonNotAllowed_GreaterThan(t *testing.T) {
	errs := validateFilter(t, `Email.Verified > true`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for boolean comparison with > operator")
		return
	}
	errMsg := errs[0].Error()
	if !strings.Contains(errMsg, "boolean") {
		t.Errorf("Error should mention boolean operator restriction, got: %v", errs[0])
	}
}

func TestTask3_BooleanComparisonNotAllowed_LessThan(t *testing.T) {
	errs := validateFilter(t, `Active < false`, reflect.TypeOf(SimpleUser{}))
	if len(errs) == 0 {
		t.Error("Expected error for boolean comparison with < operator")
	}
}

func TestTask3_BooleanEqualityAllowed(t *testing.T) {
	errs := validateFilter(t, `Active = true`, reflect.TypeOf(SimpleUser{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for boolean = comparison, got: %v", errs)
	}
}

func TestTask3_BooleanInequalityAllowed(t *testing.T) {
	errs := validateFilter(t, `Active != false`, reflect.TypeOf(SimpleUser{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for boolean != comparison, got: %v", errs)
	}
}

// Task 4: Array Field Validation

func TestTask4_HasOperatorOnArray(t *testing.T) {
	errs := validateFilter(t, `Tags:urgent`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for has operator on array, got: %v", errs)
	}
}

func TestTask4_HasOperatorTypeIncompatible(t *testing.T) {
	errs := validateFilter(t, `Tags:123`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for has operator with incompatible type")
	}
}

func TestTask4_EqualityOnArrayShouldFail(t *testing.T) {
	errs := validateFilter(t, `Tags = "urgent"`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for using = operator on array field")
		return
	}
	errMsg := errs[0].Error()
	if !strings.Contains(errMsg, "has") && !strings.Contains(errMsg, ":") && !strings.Contains(errMsg, "array") {
		t.Errorf("Error should suggest using has operator, got: %v", errs[0])
	}
}

func TestTask4_HasOperatorOnNonArray(t *testing.T) {
	errs := validateFilter(t, `Name:John`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for using has operator on non-array field")
	}
}

// Task 5: Function Call Validation

func TestTask5_TimestampFunctionValid(t *testing.T) {
	errs := validateFilter(t, `timestamp(CreatedAt) > 1000000`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for valid timestamp function, got: %v", errs)
	}
}

func TestTask5_TimestampFunctionInvalidFieldType(t *testing.T) {
	errs := validateFilter(t, `timestamp(Name) > 1000`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for timestamp function on string field")
		return
	}
	errMsg := errs[0].Error()
	if !strings.Contains(errMsg, "timestamp") {
		t.Errorf("Error should mention function name, got: %v", errs[0])
	}
}

func TestTask5_SizeFunctionOnArray(t *testing.T) {
	errs := validateFilter(t, `size(Tags) > 2`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for size function on array, got: %v", errs)
	}
}

func TestTask5_SizeFunctionOnString(t *testing.T) {
	errs := validateFilter(t, `size(Name) > 5`, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for size function on string, got: %v", errs)
	}
}

func TestTask5_SizeFunctionInvalidFieldType(t *testing.T) {
	errs := validateFilter(t, `size(Age) > 5`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for size function on int field")
	}
}

func TestTask5_UnsupportedFunction(t *testing.T) {
	errs := validateFilter(t, `unknown_func(Name) = "test"`, reflect.TypeOf(User{}))
	if len(errs) == 0 {
		t.Error("Expected error for unsupported function")
		return
	}
	errMsg := errs[0].Error()
	if !strings.Contains(errMsg, "unknown_func") && !strings.Contains(errMsg, "not supported") {
		t.Errorf("Error should mention unsupported function, got: %v", errs[0])
	}
}

// Integration Tests

func TestIntegration_ComplexFilter(t *testing.T) {
	filter := `Name = "John" AND Age > 25 AND Email.Verified = true AND Tags:urgent`
	errs := validateFilter(t, filter, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for complex valid filter, got: %v", errs)
	}
}

func TestIntegration_AllFeatures(t *testing.T) {
	filter := `(Name = "John" OR Name = "Jane") AND Age > 25 AND Email.Address = "test@example.com" AND Email.Verified = true AND Tags:urgent AND timestamp(CreatedAt) > 1000000`
	errs := validateFilter(t, filter, reflect.TypeOf(User{}))
	if len(errs) != 0 {
		t.Errorf("Expected no errors for comprehensive valid filter, got: %v", errs)
	}
}

func TestIntegration_BooleanOperatorRestrictionInComplex(t *testing.T) {
	filter := `Name = "John" AND Active > true`
	errs := validateFilter(t, filter, reflect.TypeOf(SimpleUser{}))
	if len(errs) == 0 {
		t.Error("Expected error for boolean comparison in complex filter")
	}
}

// Struct Tag Support Tests

// UserWithTags demonstrates a struct with both json and protobuf tags
type UserWithTags struct {
	ID        int64    `json:"id,omitempty" protobuf:"varint,1,opt,name=id,proto3"`
	Name      string   `json:"name,omitempty" protobuf:"bytes,2,opt,name=name,proto3"`
	Email     string   `json:"email_address" protobuf:"bytes,3,opt,name=email_address,proto3"`
	Age       int32    `json:"age" protobuf:"varint,4,opt,name=age,proto3"`
	IsActive  bool     `json:"is_active" protobuf:"varint,5,opt,name=is_active,proto3"`
	Tags      []string `json:"tags" protobuf:"bytes,6,rep,name=tags,proto3"`
	CreatedAt int64    `json:"created_at" protobuf:"varint,7,opt,name=created_at,proto3"`
}

// TestJSONTags_FieldMatching tests that JSON tags are used for field matching
func TestJSONTags_FieldMatching(t *testing.T) {
	l := lexer.New(`id = 123`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	// Validate with JSON tags - should match "id" tag
	validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithJSONTags())
	errs := validator.Validate(ast)
	if len(errs) != 0 {
		t.Errorf("Expected no errors with JSON tags, got: %v", errs)
	}
}

// TestJSONTags_FieldNotFound tests that field names that don't match JSON tags fail
func TestJSONTags_FieldNotFound(t *testing.T) {
	l := lexer.New(`ID = 123`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	// Validate with JSON tags - "ID" (PascalCase) should NOT match
	validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithJSONTags())
	errs := validator.Validate(ast)
	if len(errs) == 0 {
		t.Error("Expected error for PascalCase field name when using JSON tags")
	}
}

// TestJSONTags_ComplexFieldName tests JSON tags with underscores
func TestJSONTags_ComplexFieldName(t *testing.T) {
	l := lexer.New(`email_address = "test@example.com" AND is_active = true`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithJSONTags())
	errs := validator.Validate(ast)
	if len(errs) != 0 {
		t.Errorf("Expected no errors with JSON tag field names, got: %v", errs)
	}
}

// TestProtobufTags_FieldMatching tests that protobuf tags are used for field matching
func TestProtobufTags_FieldMatching(t *testing.T) {
	l := lexer.New(`name = "John" AND age > 25`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	// Validate with protobuf tags - should extract name= values
	validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithProtobufTags())
	errs := validator.Validate(ast)
	if len(errs) != 0 {
		t.Errorf("Expected no errors with protobuf tags, got: %v", errs)
	}
}

// TestProtobufTags_ComplexFilter tests protobuf tags with has operator
func TestProtobufTags_ComplexFilter(t *testing.T) {
	l := lexer.New(`tags:urgent AND created_at > 1000000`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithProtobufTags())
	errs := validator.Validate(ast)
	if len(errs) != 0 {
		t.Errorf("Expected no errors with protobuf tag field names, got: %v", errs)
	}
}

// TestDefaultBehavior_PascalCase tests that without options, PascalCase field names are used
func TestDefaultBehavior_PascalCase(t *testing.T) {
	l := lexer.New(`Name = "John"`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	// Without options, should use PascalCase field names
	validator := NewValidator(reflect.TypeOf(UserWithTags{}))
	errs := validator.Validate(ast)
	if len(errs) != 0 {
		t.Errorf("Expected no errors with PascalCase field name, got: %v", errs)
	}
}

// TestDefaultBehavior_JSONTagsNotUsed tests that json tags are ignored without WithJSONTags
func TestDefaultBehavior_JSONTagsNotUsed(t *testing.T) {
	l := lexer.New(`id = 123`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	// Without WithJSONTags, "id" should NOT match (only "ID" would match)
	validator := NewValidator(reflect.TypeOf(UserWithTags{}))
	errs := validator.Validate(ast)
	if len(errs) == 0 {
		t.Error("Expected error for lowercase field name without JSON tags option")
	}
}

// TestMultipleTagOptions_FirstWins tests that when multiple tag options are provided, only the first is used
func TestMultipleTagOptions_FirstWins(t *testing.T) {
	l := lexer.New(`id = 123`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	// Provide both JSON and Protobuf options - first one (JSON) should win
	validator := NewValidator(reflect.TypeOf(UserWithTags{}), WithJSONTags(), WithProtobufTags())
	errs := validator.Validate(ast)

	// "id" matches json tag, so should succeed
	if len(errs) != 0 {
		t.Errorf("Expected no errors (first option should be used), got: %v", errs)
	}
}

// TestJSONTags_NestedFields tests that nested fields work with JSON tags
func TestJSONTags_NestedFields(t *testing.T) {
	type Address struct {
		Street string `json:"street_name"`
		City   string `json:"city"`
	}
	type Person struct {
		Name    string  `json:"full_name"`
		Address Address `json:"address"`
	}

	l := lexer.New(`address.city = "NYC"`)
	p := parser.New(l)
	ast := p.ParseProgram()
	if len(p.Errors()) > 0 {
		t.Fatalf("Parse error: %v", p.Errors())
	}

	validator := NewValidator(reflect.TypeOf(Person{}), WithJSONTags())
	errs := validator.Validate(ast)
	if len(errs) != 0 {
		t.Errorf("Expected no errors for nested field with JSON tags, got: %v", errs)
	}
}
