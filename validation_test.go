package core

import (
	"errors"
	"testing"
)

func validSchema() Schema {
	return Schema{
		Tables: []Table{
			{
				Name: "contacts",
				Columns: []Column{
					{Name: "id", Type: DataTypeInteger, Generation: GenerationAuto},
					{Name: "name", Type: DataTypeString},
					{Name: "email", Type: DataTypeString, Nullable: true},
				},
				PrimaryKey: []string{"id"},
				UniqueKeys: [][]string{{"email"}},
			},
		},
	}
}

func TestValidateSchema(t *testing.T) {
	tests := []struct {
		name    string
		schema  Schema
		wantErr error
	}{
		{
			name:   "valid schema",
			schema: validSchema(),
		},
		{
			name:    "no tables",
			schema:  Schema{},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "empty table name",
			schema: Schema{
				Tables: []Table{{Columns: []Column{
					{Name: "id", Type: DataTypeInteger},
				}}},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "duplicate table",
			schema: Schema{
				Tables: []Table{
					{Name: "contacts", Columns: []Column{{Name: "id", Type: DataTypeInteger}}},
					{Name: "contacts", Columns: []Column{{Name: "id", Type: DataTypeInteger}}},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "table without columns",
			schema: Schema{
				Tables: []Table{{Name: "contacts"}},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "duplicate column",
			schema: Schema{
				Tables: []Table{
					{
						Name: "contacts",
						Columns: []Column{
							{Name: "id", Type: DataTypeInteger},
							{Name: "id", Type: DataTypeString},
						},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "unsupported data type",
			schema: Schema{
				Tables: []Table{
					{
						Name: "contacts",
						Columns: []Column{
							{Name: "id", Type: "unknown"},
						},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "unsupported generation strategy",
			schema: Schema{
				Tables: []Table{
					{
						Name: "contacts",
						Columns: []Column{
							{
								Name:       "id",
								Type:       DataTypeInteger,
								Generation: "unsupported",
							},
						},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "unknown primary key column",
			schema: Schema{
				Tables: []Table{
					{
						Name:       "contacts",
						Columns:    []Column{{Name: "id", Type: DataTypeInteger}},
						PrimaryKey: []string{"missing"},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "unknown unique key column",
			schema: Schema{
				Tables: []Table{
					{
						Name:       "contacts",
						Columns:    []Column{{Name: "id", Type: DataTypeInteger}},
						UniqueKeys: [][]string{{"missing"}},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "unknown foreign key table",
			schema: Schema{
				Tables: []Table{
					{
						Name: "contacts",
						Columns: []Column{
							{Name: "company_id", Type: DataTypeInteger},
						},
						ForeignKeys: []ForeignKey{
							{
								Columns:           []string{"company_id"},
								ReferencedTable:   "companies",
								ReferencedColumns: []string{"id"},
							},
						},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "foreign key column count mismatch",
			schema: Schema{
				Tables: []Table{
					{
						Name: "companies",
						Columns: []Column{
							{Name: "id", Type: DataTypeInteger},
						},
					},
					{
						Name: "contacts",
						Columns: []Column{
							{Name: "company_id", Type: DataTypeInteger},
						},
						ForeignKeys: []ForeignKey{
							{
								Columns:           []string{"company_id"},
								ReferencedTable:   "companies",
								ReferencedColumns: []string{"id", "other_id"},
							},
						},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "unknown foreign key source column",
			schema: Schema{
				Tables: []Table{
					{
						Name: "companies",
						Columns: []Column{
							{Name: "id", Type: DataTypeInteger},
						},
					},
					{
						Name: "contacts",
						Columns: []Column{
							{Name: "company_id", Type: DataTypeInteger},
						},
						ForeignKeys: []ForeignKey{
							{
								Columns:           []string{"missing"},
								ReferencedTable:   "companies",
								ReferencedColumns: []string{"id"},
							},
						},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
		{
			name: "unknown foreign key target column",
			schema: Schema{
				Tables: []Table{
					{
						Name: "companies",
						Columns: []Column{
							{Name: "id", Type: DataTypeInteger},
						},
					},
					{
						Name: "contacts",
						Columns: []Column{
							{Name: "company_id", Type: DataTypeInteger},
						},
						ForeignKeys: []ForeignKey{
							{
								Columns:           []string{"company_id"},
								ReferencedTable:   "companies",
								ReferencedColumns: []string{"missing"},
							},
						},
					},
				},
			},
			wantErr: ErrInvalidSchema,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSchema(tt.schema)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateSchema() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMapping(t *testing.T) {
	schema := validSchema()

	tests := []struct {
		name    string
		mapping Mapping
		wantErr error
	}{
		{
			name: "valid mapping",
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{SourceField: "FN", TargetField: "name"},
							{SourceField: "EMAIL", TargetField: "email"},
						},
					},
				},
			},
		},
		{
			name: "unknown target table",
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "unknown",
						Fields: []FieldMapping{
							{SourceField: "FN", TargetField: "name"},
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name: "unknown target field",
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{SourceField: "FN", TargetField: "unknown"},
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name: "empty source field",
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{TargetField: "name"},
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name: "empty target field",
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{SourceField: "FN"},
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name: "duplicate target field",
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{SourceField: "FN", TargetField: "name"},
							{SourceField: "N", TargetField: "name"},
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name: "duplicate target table mapping",
			mapping: Mapping{
				Tables: []TableMapping{
					{TargetTable: "contacts"},
					{TargetTable: "contacts"},
				},
			},
			wantErr: ErrInvalidMapping,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMapping(tt.mapping, schema)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateMapping() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMappingParentReference(t *testing.T) {
	schema := Schema{
		Tables: []Table{
			{
				Name: "contacts",
				Columns: []Column{
					{Name: "id", Type: DataTypeInteger, Generation: GenerationAuto},
					{Name: "name", Type: DataTypeString},
				},
				PrimaryKey: []string{"id"},
			},
			{
				Name: "emails",
				Columns: []Column{
					{Name: "id", Type: DataTypeInteger, Generation: GenerationAuto},
					{Name: "contact_id", Type: DataTypeInteger},
					{Name: "address", Type: DataTypeString},
				},
				PrimaryKey: []string{"id"},
				ForeignKeys: []ForeignKey{
					{
						Columns:           []string{"contact_id"},
						ReferencedTable:   "contacts",
						ReferencedColumns: []string{"id"},
					},
				},
			},
		},
	}

	validMapping := Mapping{
		Tables: []TableMapping{
			{
				TargetTable: "contacts",
				Fields: []FieldMapping{
					{SourceField: "FN", TargetField: "name"},
				},
			},
			{
				TargetTable:      "emails",
				SourceCollection: "EMAIL",
				ParentReference: &ParentReference{
					ParentTable: "contacts",
					ParentField: "id",
					TargetField: "contact_id",
				},
				Fields: []FieldMapping{
					{SourceField: "value", TargetField: "address"},
				},
			},
		},
	}

	tests := []struct {
		name    string
		schema  Schema
		mapping Mapping
		wantErr error
	}{
		{
			name:    "valid parent reference",
			schema:  schema,
			mapping: validMapping,
		},
		{
			name:   "missing parent mapping",
			schema: schema,
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "emails",
						ParentReference: &ParentReference{
							ParentTable: "contacts",
							ParentField: "id",
							TargetField: "contact_id",
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name: "missing foreign key",
			schema: Schema{
				Tables: []Table{
					schema.Tables[0],
					{
						Name: "emails",
						Columns: []Column{
							{Name: "id", Type: DataTypeInteger},
							{Name: "contact_id", Type: DataTypeInteger},
						},
						PrimaryKey: []string{"id"},
					},
				},
			},
			mapping: validMapping,
			wantErr: ErrInvalidMapping,
		},
		{
			name:   "parent field also explicitly mapped",
			schema: schema,
			mapping: Mapping{
				Tables: []TableMapping{
					validMapping.Tables[0],
					{
						TargetTable: "emails",
						ParentReference: &ParentReference{
							ParentTable: "contacts",
							ParentField: "id",
							TargetField: "contact_id",
						},
						Fields: []FieldMapping{
							{SourceField: "CONTACT_ID", TargetField: "contact_id"},
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name: "cyclic parent dependencies",
			schema: Schema{
				Tables: []Table{
					{
						Name: "contacts",
						Columns: []Column{
							{Name: "id", Type: DataTypeInteger},
							{Name: "email_id", Type: DataTypeInteger},
						},
						PrimaryKey: []string{"id"},
						ForeignKeys: []ForeignKey{
							{
								Columns:           []string{"email_id"},
								ReferencedTable:   "emails",
								ReferencedColumns: []string{"id"},
							},
						},
					},
					{
						Name: "emails",
						Columns: []Column{
							{Name: "id", Type: DataTypeInteger},
							{Name: "contact_id", Type: DataTypeInteger},
						},
						PrimaryKey: []string{"id"},
						ForeignKeys: []ForeignKey{
							{
								Columns:           []string{"contact_id"},
								ReferencedTable:   "contacts",
								ReferencedColumns: []string{"id"},
							},
						},
					},
				},
			},
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						ParentReference: &ParentReference{
							ParentTable: "emails",
							ParentField: "id",
							TargetField: "email_id",
						},
					},
					{
						TargetTable: "emails",
						ParentReference: &ParentReference{
							ParentTable: "contacts",
							ParentField: "id",
							TargetField: "contact_id",
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMapping(tt.mapping, tt.schema)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateMapping() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateMappingRequiredColumns(t *testing.T) {
	defaultValue := Value("Unknown")

	tests := []struct {
		name    string
		mapping Mapping
		schema  Schema
		wantErr error
	}{
		{
			name:   "missing required column",
			schema: validSchema(),
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{SourceField: "EMAIL", TargetField: "email"},
						},
					},
				},
			},
			wantErr: ErrInvalidMapping,
		},
		{
			name:   "required column with default",
			schema: validSchema(),
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{
								SourceField: "FN",
								TargetField: "name",
								Default:     &defaultValue,
							},
						},
					},
				},
			},
		},
		{
			name:   "nullable column without mapping",
			schema: validSchema(),
			mapping: Mapping{
				Tables: []TableMapping{
					{
						TargetTable: "contacts",
						Fields: []FieldMapping{
							{SourceField: "FN", TargetField: "name"},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateMapping(tt.mapping, tt.schema)

			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("ValidateMapping() error = %v, want %v",
					err, tt.wantErr)
			}
		})
	}
}

func TestValidateMappingRejectsNegativeOccurrence(t *testing.T) {
	schema := Schema{
		Tables: []Table{
			{
				Name: "contacts",
				Columns: []Column{
					{
						Name:     "email",
						Type:     DataTypeString,
						Nullable: true,
					},
				},
			},
		},
	}

	mapping := Mapping{
		Tables: []TableMapping{
			{
				TargetTable: "contacts",
				Fields: []FieldMapping{
					{
						SourceField: "EMAIL",
						TargetField: "email",
						Occurrence:  -1,
					},
				},
			},
		},
	}

	err := ValidateMapping(mapping, schema)
	if !errors.Is(err, ErrInvalidMapping) {
		t.Fatalf("expected ErrInvalidMapping, got %v", err)
	}
}
