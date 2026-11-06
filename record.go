package core

// Record represents a single logical source record.
//
// Field names are source-independent identifiers. Their interpretation
// and mapping to target fields are defined by the mapping configuration.
type Record map[string]Value

// Value represents a generic source or target field value.
//
// Supported concrete values:
//   - string
//   - int64
//   - float64
//   - bool
//   - time.Time
//   - []byte
//   - []Value for repeated values
//   - map[string]Value for structured values
//   - nil for an explicit null value
//
// Source adapters are responsible for converting parsed values into
// these types. Target adapters are responsible for converting them into
// database-compatible values.
type Value any
