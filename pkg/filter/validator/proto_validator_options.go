package validator

// proto_validator_options.go contains configuration options for ProtoValidator.
//
// This file defines the options struct and functional option pattern for
// configuring ProtoValidator behavior. Options can be extended in the future
// without breaking the API.

// ProtoValidatorOptions holds configuration for ProtoValidator behavior.
// This struct can be extended with new options in the future without
// breaking the API.
type ProtoValidatorOptions struct {
	// EnableEnumPrefixStripping allows enum values to match with prefix stripped.
	// When true (default): "ACTIVE" matches "STATUS_ACTIVE"
	// When false: only "STATUS_ACTIVE" matches
	EnableEnumPrefixStripping bool

	// Future options can be added here:
	// AllowCaseInsensitiveFields bool
	// StrictModeEnabled bool
	// CustomValidators map[string]func(...) error
}

// ProtoValidatorOption is a functional option for configuring ProtoValidator.
type ProtoValidatorOption func(*ProtoValidatorOptions)

// WithEnumPrefixStripping controls whether enum values can be matched with their
// prefix stripped. When enabled (default), both forms are accepted:
//   - status = "STATUS_ACTIVE" (exact match)
//   - status = "ACTIVE" (prefix-stripped: STATUS_ + ACTIVE)
//
// When disabled, only exact matches are accepted:
//   - status = "STATUS_ACTIVE" (only this works)
//   - status = "ACTIVE" (fails)
//
// Default: true (enabled for user convenience and AIP-160 "non-technical audience" principle)
//
// Example:
//
//	validator := NewProtoValidator(msgDesc, WithEnumPrefixStripping(false))
func WithEnumPrefixStripping(enable bool) ProtoValidatorOption {
	return func(opts *ProtoValidatorOptions) {
		opts.EnableEnumPrefixStripping = enable
	}
}
