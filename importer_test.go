package core

import (
	"context"
	"reflect"
	"testing"
)

type importerTestTransaction struct {
	inserted []MappedRow
}

func (tx *importerTestTransaction) Insert(
	_ context.Context,
	table string,
	record Record,
) (Record, error) {
	copyRecord := make(Record, len(record))
	for key, value := range record {
		copyRecord[key] = value
	}

	tx.inserted = append(tx.inserted, MappedRow{
		Table:  table,
		Record: copyRecord,
	})

	if table == "contacts" {
		return Record{"id": int64(42)}, nil
	}

	return Record{}, nil
}

func (tx *importerTestTransaction) Commit(context.Context) error {
	return nil
}

func (tx *importerTestTransaction) Rollback(context.Context) error {
	return nil
}

func TestImportRecordParentAndCollection(t *testing.T) {
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
				TargetTable:      "emails",
				SourceCollection: "EMAIL",
				ParentReference: &ParentReference{
					ParentTable: "contacts",
					ParentField: "id",
					TargetField: "contact_id",
				},
				Fields: []FieldMapping{
					{
						SourceField: "VALUE",
						TargetField: "address",
					},
				},
			},
			{
				TargetTable: "contacts",
				Fields: []FieldMapping{
					{
						SourceField: "FN",
						TargetField: "name",
					},
				},
			},
		},
	}

	tx := &importerTestTransaction{}

	err := ImportRecord(context.Background(), tx, record, mapping, nil)
	if err != nil {
		t.Fatalf("ImportRecord() unexpected error: %v", err)
	}

	want := []MappedRow{
		{
			Table:  "contacts",
			Record: Record{"name": "Alice"},
		},
		{
			Table: "emails",
			Record: Record{
				"address":    "alice@example.com",
				"contact_id": int64(42),
			},
		},
		{
			Table: "emails",
			Record: Record{
				"address":    "alice.work@example.com",
				"contact_id": int64(42),
			},
		},
	}

	if !reflect.DeepEqual(tx.inserted, want) {
		t.Fatalf("unexpected inserts:\n got: %#v\nwant: %#v", tx.inserted, want)
	}
}
