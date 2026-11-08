package core

import (
	"errors"
	"testing"
)

func TestMapRecord(t *testing.T) {
	defaultValue := Value("Unknown")

	tests := []struct {
		name       string
		record     Record
		mapping    Mapping
		transforms map[string]TransformFunc
		want       []MappedRow
		wantErr    error
	}{
		{
			name: "maps source fields",
			record: Record{
				"FN":    "Alice",
				"EMAIL": "alice@example.com",
			},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{
					{SourceField: "FN", TargetField: "name"},
					{SourceField: "EMAIL", TargetField: "email"},
				},
			}}},
			want: []MappedRow{{
				Table: "contacts",
				Record: Record{
					"name":  "Alice",
					"email": "alice@example.com",
				},
			}},
		},
		{
			name:   "uses default for missing source field",
			record: Record{},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "FN",
					TargetField: "name",
					Default:     &defaultValue,
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"name": "Unknown"},
			}},
		},
		{
			name:   "preserves explicit null default",
			record: Record{},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "FN",
					TargetField: "name",
					Default:     func() *Value { v := Value(nil); return &v }(),
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"name": nil},
			}},
		},
		{
			name:   "applies transformation",
			record: Record{"FN": "alice"},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "FN",
					TargetField: "name",
					Transform:   "uppercase",
				}},
			}}},
			transforms: map[string]TransformFunc{
				"uppercase": func(value Value) (Value, error) {
					return "ALICE", nil
				},
			},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"name": "ALICE"},
			}},
		},
		{
			name:   "missing source without default",
			record: Record{},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "FN",
					TargetField: "name",
				}},
			}}},
			wantErr: ErrInvalidRecord,
		},
		{
			name:   "unknown transformation",
			record: Record{"FN": "Alice"},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "FN",
					TargetField: "name",
					Transform:   "unknown",
				}},
			}}},
			wantErr: ErrInvalidMapping,
		},
		{
			name:   "transformation error",
			record: Record{"FN": "Alice"},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "FN",
					TargetField: "name",
					Transform:   "fail",
				}},
			}}},
			transforms: map[string]TransformFunc{
				"fail": func(Value) (Value, error) {
					return nil, errors.New("invalid value")
				},
			},
			wantErr: ErrInvalidRecord,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := MapRecord(tt.record, tt.mapping, tt.transforms)

			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error %v, got %v", tt.wantErr, err)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(got) != len(tt.want) {
				t.Fatalf("expected %d rows, got %d", len(tt.want), len(got))
			}

			for i := range got {
				if got[i].Table != tt.want[i].Table {
					t.Errorf("row %d: expected table %q, got %q",
						i, tt.want[i].Table, got[i].Table)
				}

				if len(got[i].Record) != len(tt.want[i].Record) {
					t.Errorf("row %d: expected %d fields, got %d",
						i, len(tt.want[i].Record), len(got[i].Record))
					continue
				}

				for key, wantValue := range tt.want[i].Record {
					if gotValue, exists := got[i].Record[key]; !exists || gotValue != wantValue {
						t.Errorf("row %d field %q: expected %#v, got %#v",
							i, key, wantValue, gotValue)
					}
				}
			}
		})
	}
}
