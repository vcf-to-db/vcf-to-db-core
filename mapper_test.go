package core

import (
	"errors"
	"reflect"
	"strings"
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

func TestMapRecordCollection(t *testing.T) {
	defaultValue := Value("unknown")

	tests := []struct {
		name         string
		record       Record
		defaultValue *Value
		want         []MappedRow
		wantErr      error
	}{
		{
			name: "maps each collection item",
			record: Record{
				"EMAILS": []Value{
					map[string]Value{"VALUE": "alice@example.com"},
					map[string]Value{"VALUE": "alice.work@example.com"},
				},
			},
			want: []MappedRow{
				{Table: "emails", Record: Record{"address": "alice@example.com"}},
				{Table: "emails", Record: Record{"address": "alice.work@example.com"}},
			},
		},
		{
			name: "empty collection produces no rows",
			record: Record{
				"EMAILS": []Value{},
			},
			want: []MappedRow{},
		},
		{
			name:    "missing collection",
			record:  Record{},
			wantErr: ErrInvalidRecord,
		},
		{
			name: "collection is not repeated",
			record: Record{
				"EMAILS": "alice@example.com",
			},
			wantErr: ErrInvalidRecord,
		},
		{
			name: "collection item is not structured",
			record: Record{
				"EMAILS": []Value{"alice@example.com"},
			},
			wantErr: ErrInvalidRecord,
		},
		{
			name: "missing item field uses default",
			record: Record{
				"EMAILS": []Value{map[string]Value{}},
			},
			defaultValue: &defaultValue,
			want: []MappedRow{
				{Table: "emails", Record: Record{"address": "unknown"}},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mapping := Mapping{
				Tables: []TableMapping{
					{
						TargetTable:      "emails",
						SourceCollection: "EMAILS",
						Fields: []FieldMapping{
							{
								SourceField: "VALUE",
								TargetField: "address",
								Default:     tt.defaultValue,
							},
						},
					},
				},
			}

			got, err := MapRecord(tt.record, mapping, nil)

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

			for i := range tt.want {
				if got[i].Table != tt.want[i].Table {
					t.Errorf("row %d: expected table %q, got %q",
						i, tt.want[i].Table, got[i].Table)
				}

				for key, wantValue := range tt.want[i].Record {
					if gotValue := got[i].Record[key]; gotValue != wantValue {
						t.Errorf("row %d field %q: expected %v, got %v",
							i, key, wantValue, gotValue)
					}
				}

				if len(got[i].Record) != len(tt.want[i].Record) {
					t.Errorf("row %d: expected record %v, got %v",
						i, tt.want[i].Record, got[i].Record)
				}
			}
		})
	}
}

func TestMapRecordRootAndCollectionMappings(t *testing.T) {
	record := Record{
		"FN": "Alice",
		"EMAIL": []Value{
			map[string]Value{"VALUE": "alice@example.com"},
			map[string]Value{"VALUE": "alice.work@example.com"},
		},
	}

	mapping := Mapping{
		Tables: []TableMapping{
			{
				TargetTable: "contacts",
				Fields: []FieldMapping{
					{
						SourceField: "FN",
						TargetField: "name",
					},
				},
			},
			{
				TargetTable:      "emails",
				SourceCollection: "EMAIL",
				Fields: []FieldMapping{
					{
						SourceField: "VALUE",
						TargetField: "address",
						Transform:   "uppercase",
					},
				},
			},
		},
	}

	transforms := map[string]TransformFunc{
		"uppercase": func(value Value) (Value, error) {
			text, ok := value.(string)
			if !ok {
				return nil, ErrInvalidRecord
			}
			return strings.ToUpper(text), nil
		},
	}

	want := []MappedRow{
		{
			Table:  "contacts",
			Record: Record{"name": "Alice"},
		},
		{
			Table:  "emails",
			Record: Record{"address": "ALICE@EXAMPLE.COM"},
		},
		{
			Table:  "emails",
			Record: Record{"address": "ALICE.WORK@EXAMPLE.COM"},
		},
	}

	got, err := MapRecord(record, mapping, transforms)
	if err != nil {
		t.Fatalf("MapRecord() unexpected error: %v", err)
	}

	if len(got) != len(want) {
		t.Fatalf("expected %d rows, got %d", len(want), len(got))
	}

	for i := range want {
		if got[i].Table != want[i].Table {
			t.Errorf("row %d: expected table %q, got %q",
				i, want[i].Table, got[i].Table)
		}

		if len(got[i].Record) != len(want[i].Record) {
			t.Errorf("row %d: expected record %v, got %v",
				i, want[i].Record, got[i].Record)
			continue
		}

		for key, wantValue := range want[i].Record {
			if gotValue := got[i].Record[key]; gotValue != wantValue {
				t.Errorf("row %d field %q: expected %v, got %v",
					i, key, wantValue, gotValue)
			}
		}
	}
}
