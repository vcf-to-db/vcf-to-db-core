package core

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type importerTestTransaction struct {
	inserted        []MappedRow
	returnedByTable map[string]Record
	failTable       string
	insertErr       error
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

	if table == tx.failTable {
		return nil, tx.insertErr
	}

	return tx.returnedByTable[table], nil
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
					{SourceField: "VALUE", TargetField: "address"},
				},
			},
			{
				TargetTable: "contacts",
				Fields: []FieldMapping{
					{SourceField: "FN", TargetField: "name"},
				},
			},
		},
	}

	tx := &importerTestTransaction{
		returnedByTable: map[string]Record{
			"contacts": {"id": int64(42)},
		},
	}

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

func TestImportRecordMissingParentKey(t *testing.T) {
	record := Record{
		"FN": "Alice",
		"EMAIL": []Value{
			map[string]Value{"VALUE": "alice@example.com"},
		},
	}

	mapping := Mapping{
		Tables: []TableMapping{
			{
				TargetTable: "contacts",
				Fields:      []FieldMapping{{SourceField: "FN", TargetField: "name"}},
			},
			{
				TargetTable:      "emails",
				SourceCollection: "EMAIL",
				ParentReference: &ParentReference{
					ParentTable: "contacts",
					ParentField: "id",
					TargetField: "contact_id",
				},
				Fields: []FieldMapping{{SourceField: "VALUE", TargetField: "address"}},
			},
		},
	}

	tx := &importerTestTransaction{}

	err := ImportRecord(context.Background(), tx, record, mapping, nil)
	if err == nil {
		t.Fatal("ImportRecord() expected an error for missing parent key")
	}

	if len(tx.inserted) != 1 || tx.inserted[0].Table != "contacts" {
		t.Fatalf("unexpected inserts: %#v", tx.inserted)
	}
}

func TestImportRecordMultipleParentRows(t *testing.T) {
	record := Record{
		"CONTACTS": []Value{
			map[string]Value{"NAME": "Alice"},
			map[string]Value{"NAME": "Bob"},
		},
		"EMAIL": []Value{
			map[string]Value{"VALUE": "alice@example.com"},
		},
	}

	mapping := Mapping{
		Tables: []TableMapping{
			{
				TargetTable:      "contacts",
				SourceCollection: "CONTACTS",
				Fields: []FieldMapping{
					{SourceField: "NAME", TargetField: "name"},
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
					{SourceField: "VALUE", TargetField: "address"},
				},
			},
		},
	}

	tx := &importerTestTransaction{
		returnedByTable: map[string]Record{
			"contacts": {"id": int64(42)},
		},
	}

	err := ImportRecord(context.Background(), tx, record, mapping, nil)
	if err == nil {
		t.Fatal("ImportRecord() expected an error for multiple parent rows")
	}

	if len(tx.inserted) != 2 {
		t.Fatalf("expected only two parent inserts, got %#v", tx.inserted)
	}

	for _, inserted := range tx.inserted {
		if inserted.Table != "contacts" {
			t.Fatalf("unexpected child insert: %#v", inserted)
		}
	}
}

func TestImportRecordStopsAfterInsertFailure(t *testing.T) {
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
					{SourceField: "VALUE", TargetField: "address"},
				},
			},
		},
	}

	insertErr := errors.New("simulated insert failure")
	tx := &importerTestTransaction{
		returnedByTable: map[string]Record{
			"contacts": {"id": int64(42)},
		},
		failTable: "emails",
		insertErr: insertErr,
	}

	err := ImportRecord(context.Background(), tx, record, mapping, nil)
	if !errors.Is(err, insertErr) {
		t.Fatalf("expected insert error %v, got %v", insertErr, err)
	}

	if len(tx.inserted) != 2 {
		t.Fatalf("expected contact and one attempted email insert, got %#v", tx.inserted)
	}

	if tx.inserted[0].Table != "contacts" || tx.inserted[1].Table != "emails" {
		t.Fatalf("unexpected insert order: %#v", tx.inserted)
	}
}
