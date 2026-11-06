package core

import "errors"

var (
	// ErrInvalidConfiguration indicates invalid application configuration.
	ErrInvalidConfiguration = errors.New("invalid configuration")

	// ErrInvalidSchema indicates an invalid target schema.
	ErrInvalidSchema = errors.New("invalid schema")

	// ErrInvalidMapping indicates an invalid field mapping.
	ErrInvalidMapping = errors.New("invalid mapping")

	// ErrInvalidRecord indicates a record that violates the core record contract.
	ErrInvalidRecord = errors.New("invalid record")

	// ErrUnsupportedFeature indicates that an adapter does not support
	// a requested operation or schema feature.
	ErrUnsupportedFeature = errors.New("unsupported feature")
)
