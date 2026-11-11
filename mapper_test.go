package core

import (
	"errors"
	"reflect"
	"testing"
)

func TestMapRecord(t *testing.T) {
	defaultValue := Value("Unknown")
	nullValue := Value(nil)

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
					Default:     &nullValue,
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"name": nil},
			}},
		},
		{
			name: "applies transformation",
			record: Record{
				"FN": "alice",
			},
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
			name: "maps first occurrence",
			record: Record{
				"EMAIL": []Value{"first@example.com", "second@example.com"},
			},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "EMAIL",
					TargetField: "email",
					Occurrence:  0,
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"email": "first@example.com"},
			}},
		},
		{
			name: "maps second occurrence",
			record: Record{
				"EMAIL": []Value{"first@example.com", "second@example.com"},
			},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "EMAIL",
					TargetField: "email",
					Occurrence:  1,
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"email": "second@example.com"},
			}},
		},
		{
			name: "uses default when occurrence is out of range",
			record: Record{
				"EMAIL": []Value{"first@example.com"},
			},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "EMAIL",
					TargetField: "email",
					Occurrence:  1,
					Default:     &defaultValue,
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"email": "Unknown"},
			}},
		},
		{
			name: "errors when occurrence is out of range without default",
			record: Record{
				"EMAIL": []Value{"first@example.com"},
			},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "EMAIL",
					TargetField: "email",
					Occurrence:  1,
				}},
			}}},
			wantErr: ErrInvalidRecord,
		},
		{
			name: "maps scalar value at occurrence zero",
			record: Record{
				"EMAIL": "first@example.com",
			},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "EMAIL",
					TargetField: "email",
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"email": "first@example.com"},
			}},
		},
		{
			name: "uses default for scalar value at nonzero occurrence",
			record: Record{
				"EMAIL": "first@example.com",
			},
			mapping: Mapping{Tables: []TableMapping{{
				TargetTable: "contacts",
				Fields: []FieldMapping{{
					SourceField: "EMAIL",
					TargetField: "email",
					Occurrence:  1,
					Default:     &defaultValue,
				}},
			}}},
			want: []MappedRow{{
				Table:  "contacts",
				Record: Record{"email": "Unknown"},
			}},
		},
		{
			name:   "errors for missing source without default",
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
			name: "errors for unknown transformation",
			record: Record{
				"FN": "Alice",
			},
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
			name: "errors when transformation fails",
			record: Record{
				"FN": "Alice",
			},
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

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("expected %#v, got %#v", tt.want, got)
			}
		})
	}
}
