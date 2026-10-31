package core

// Mapping defines how source records are transformed into target records.
type Mapping struct {
	Tables []TableMapping
}

// TableMapping defines the field mappings for a target table.
type TableMapping struct {
	TargetTable string
	Fields      []FieldMapping
}

// FieldMapping defines how a single target field is populated.
type FieldMapping struct {
	TargetField string
	SourceField string
	Transform   string
	Default     Value
}
