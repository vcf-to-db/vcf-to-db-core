package core

// Schema describes the logical structure of the target database.
type Schema struct {
	Tables []Table
}

// Table describes a logical database table.
type Table struct {
	Name        string
	Columns     []Column
	PrimaryKey  []string
	UniqueKeys  [][]string
	ForeignKeys []ForeignKey
}

// Column describes a table column.
type Column struct {
	Name       string
	Type       DataType
	Nullable   bool
	Generation GenerationStrategy
}

// GenerationStrategy describes how a column value is generated.
type GenerationStrategy string

const (
	// GenerationNone means the value must be supplied by the ETL process.
	GenerationNone GenerationStrategy = ""

	// GenerationAuto means the target database generates the value.
	GenerationAuto GenerationStrategy = "auto"
)

// DataType identifies a database-independent logical data type.
type DataType string

const (
	DataTypeString    DataType = "string"
	DataTypeInteger   DataType = "integer"
	DataTypeDecimal   DataType = "decimal"
	DataTypeBoolean   DataType = "boolean"
	DataTypeTimestamp DataType = "timestamp"
	DataTypeBytes     DataType = "bytes"
)

// ForeignKey describes a relationship between two tables.
type ForeignKey struct {
	Columns           []string
	ReferencedTable   string
	ReferencedColumns []string
}
