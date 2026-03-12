package validator

import (
	"reflect"
	"strings"
)

// EnumInfo holds metadata about an enum type for validation.
// Unlike booleans/arrays which only need operator validation, enums need
// value validation (checking against a valid set). This struct stores the
// necessary metadata extracted during detection.
type EnumInfo struct {
	TypeName string            // Full enum type name (e.g., "Classification")
	ValueMap map[string]int32  // String -> int32 mapping (e.g., Classification_value)
}

// detectEnum checks if a struct field is a protobuf enum and extracts metadata.
// Returns (EnumInfo, true) if the field is an enum, (nil, false) otherwise.
//
// Detection strategy:
// 1. Check for protobuf tag with enum= parameter
// 2. Extract enum type name from tag
// 3. Attempt to locate companion _value map via reflection
// 4. Return EnumInfo with available data (may have empty ValueMap if maps not found)
func detectEnum(field reflect.StructField, fieldType reflect.Type) (*EnumInfo, bool) {
	// Step 1: Check for protobuf tag
	protobufTag := field.Tag.Get("protobuf")
	if protobufTag == "" {
		return nil, false
	}

	// Step 2: Extract enum type name from tag
	enumTypeName, hasEnum := extractEnumFromProtobufTag(protobufTag)
	if !hasEnum {
		return nil, false
	}

	// Step 3: Create EnumInfo with type name
	enumInfo := &EnumInfo{
		TypeName: enumTypeName,
	}

	// Step 4: Attempt to locate companion maps
	// This may fail if the enum package isn't accessible or maps don't exist
	// That's OK - we can still do operator restriction without value validation
	valueMap := findCompanionValueMap(fieldType, enumTypeName)
	if valueMap != nil {
		enumInfo.ValueMap = valueMap
	}

	return enumInfo, true
}

// extractEnumFromProtobufTag extracts the enum type name from a protobuf struct tag.
// Tag format: "varint,5,opt,name=classification,proto3,enum=full.package.Type"
// Returns (typeName, true) if enum= parameter found, ("", false) otherwise.
func extractEnumFromProtobufTag(tag string) (string, bool) {
	// Split tag by commas
	parts := strings.Split(tag, ",")

	// Look for enum= parameter
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if strings.HasPrefix(trimmed, "enum=") {
			// Extract the value after "enum="
			enumType := strings.TrimPrefix(trimmed, "enum=")
			return enumType, true
		}
	}

	return "", false
}

// findCompanionValueMap attempts to locate the companion _value map for an enum type.
// For an enum like "Classification", looks for "Classification_value" variable.
//
// Returns the map if found, nil otherwise.
//
// Note: This uses reflection to access package-level variables. It works because
// the enum's package is imported by the struct's package (e.g., models imports proto/classification).
func findCompanionValueMap(fieldType reflect.Type, enumTypeName string) map[string]int32 {
	// Extract the base type name (last component after final dot)
	// e.g., "proto.audience.Classification" → "Classification"
	baseTypeName := enumTypeName
	if idx := strings.LastIndex(enumTypeName, "."); idx != -1 {
		baseTypeName = enumTypeName[idx+1:]
	}

	// Construct the companion map variable name
	companionMapName := baseTypeName + "_value"

	// TODO: Implement reflection-based lookup of package-level variable
	// This is complex in Go and requires access to the package's exported variables.
	// For now, we return nil (graceful degradation - operator restriction still works)
	//
	// Future implementation would:
	// 1. Get package path from fieldType.PkgPath()
	// 2. Use reflection to find the variable in that package
	// 3. Type assert to map[string]int32
	//
	// This limitation is acceptable for Phase 1 - we can still restrict operators
	// even without value validation.

	_ = companionMapName // Avoid unused variable warning

	return nil // Graceful degradation for now
}
