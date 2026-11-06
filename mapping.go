package core

// Mapping defines how source records are transformed into target records.
type Mapping struct {
	Tables []TableMapping
}

// TableMapping defines how source data populates a target table.
type TableMapping struct {
	TargetTable string

	// SourceCollection identifies a repeated collection in the source record.
	//
	// When empty, at most one row is produced per source record.
	// When set, the mapper iterates over the collection's []Value items
	// and produces one row per item.
	SourceCollection string

	// ParentReference describes the relationship to a parent row.
	// The parent row must be inserted before the child row.
	ParentReference *ParentReference

	Fields []FieldMapping
}

// ParentReference identifies the parent mapping and the columns
// used to establish the relationship between parent and child rows.
type ParentReference struct {
	ParentTable string
	ParentField string
	TargetField string
}

// FieldMapping defines how a target field is populated.
//
// SourceField is resolved against the root Record when SourceCollection
// is empty. When SourceCollection is set, it is resolved against each
// collection item, which must be a map[string]Value.
//
// Transform identifies an optional transformation registered by the
// application. An empty value means no transformation is applied.
//
// Default is used when the source field is absent. A nil Default means
// no default was configured; a pointer to a nil Value represents an
// explicitly configured null default.
type FieldMapping struct {
	TargetField string
	SourceField string
	Transform   string
	Default     *Value
}
