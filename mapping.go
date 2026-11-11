package core

type Mapping struct {
	Tables []TableMapping
}

type TableMapping struct {
	TargetTable      string
	SourceCollection string
	ParentReference  *ParentReference
	Fields           []FieldMapping
}

type ParentReference struct {
	ParentTable string
	ParentField string
	TargetField string
}

type FieldMapping struct {
	TargetField string
	SourceField string
	Transform   string
	Default     *Value
	Occurrence  int
}
