package core

// Capabilities describes the features supported by a target adapter.
type Capabilities struct {
	SupportedDataTypes []DataType

	PrimaryKeys bool
	UniqueKeys  bool
	ForeignKeys bool

	BatchWrites  bool
	Transactions bool
}

// SupportsDataType reports whether the adapter supports the given data type.
func (c Capabilities) SupportsDataType(dataType DataType) bool {
	for _, supported := range c.SupportedDataTypes {
		if supported == dataType {
			return true
		}
	}

	return false
}
