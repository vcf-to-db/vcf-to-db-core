package core

import (
	"context"
	"fmt"
)

// ImportRecord maps and inserts all rows produced from one source record.
//
// The mapping must have passed ValidateMapping before this function is called.
// Parent tables are inserted before their child tables.
//
// ImportRecord does not commit or roll back the transaction. The caller
// retains responsibility for transaction lifecycle.
func ImportRecord(
	ctx context.Context,
	tx Transaction,
	record Record,
	mapping Mapping,
	transforms map[string]TransformFunc,
) error {
	orderedMappings, err := orderTableMappings(mapping)
	if err != nil {
		return err
	}

	insertedByTable := make(map[string][]Record, len(mapping.Tables))

	for _, tableMapping := range orderedMappings {
		sourceRecords, err := resolveSourceRecords(record, tableMapping.SourceCollection)
		if err != nil {
			return fmt.Errorf("table %q: %w", tableMapping.TargetTable, err)
		}

		insertedRows := make([]Record, 0, len(sourceRecords))

		for i, sourceRecord := range sourceRecords {
			targetRecord, err := mapTableRecord(sourceRecord, tableMapping, transforms)
			if err != nil {
				if tableMapping.SourceCollection != "" {
					return fmt.Errorf("table %q, collection %q item %d: %w",
						tableMapping.TargetTable, tableMapping.SourceCollection, i, err)
				}
				return fmt.Errorf("table %q: %w", tableMapping.TargetTable, err)
			}

			if parent := tableMapping.ParentReference; parent != nil {
				parentRows := insertedByTable[parent.ParentTable]
				if len(parentRows) != 1 {
					return fmt.Errorf("%w: parent table %q must produce exactly one row; got %d",
						ErrInvalidMapping, parent.ParentTable, len(parentRows))
				}

				parentValue, exists := parentRows[0][parent.ParentField]
				if !exists || parentValue == nil {
					return fmt.Errorf("%w: parent table %q did not provide field %q",
						ErrInvalidRecord, parent.ParentTable, parent.ParentField)
				}

				targetRecord[parent.TargetField] = parentValue
			}

			insertedRecord, err := tx.Insert(ctx, tableMapping.TargetTable, targetRecord)
			if err != nil {
				return fmt.Errorf("insert into table %q: %w", tableMapping.TargetTable, err)
			}

			combinedRecord := make(Record, len(targetRecord)+len(insertedRecord))
			for key, value := range targetRecord {
				combinedRecord[key] = value
			}
			for key, value := range insertedRecord {
				combinedRecord[key] = value
			}

			insertedRows = append(insertedRows, combinedRecord)
		}

		insertedByTable[tableMapping.TargetTable] = insertedRows
	}

	return nil
}

func orderTableMappings(mapping Mapping) ([]TableMapping, error) {
	mappingsByTable := make(map[string]TableMapping, len(mapping.Tables))

	for _, tableMapping := range mapping.Tables {
		mappingsByTable[tableMapping.TargetTable] = tableMapping
	}

	const (
		unvisited = iota
		visiting
		visited
	)

	state := make(map[string]int, len(mapping.Tables))
	ordered := make([]TableMapping, 0, len(mapping.Tables))

	var visit func(string) error
	visit = func(table string) error {
		switch state[table] {
		case visiting:
			return fmt.Errorf("%w: cyclic parent dependency involving table %q",
				ErrInvalidMapping, table)
		case visited:
			return nil
		}

		tableMapping, exists := mappingsByTable[table]
		if !exists {
			return fmt.Errorf("%w: table %q is not mapped",
				ErrInvalidMapping, table)
		}

		state[table] = visiting

		if parent := tableMapping.ParentReference; parent != nil {
			if _, exists := mappingsByTable[parent.ParentTable]; !exists {
				return fmt.Errorf("%w: parent table %q is not mapped",
					ErrInvalidMapping, parent.ParentTable)
			}

			if err := visit(parent.ParentTable); err != nil {
				return err
			}
		}

		state[table] = visited
		ordered = append(ordered, tableMapping)
		return nil
	}

	for _, tableMapping := range mapping.Tables {
		if err := visit(tableMapping.TargetTable); err != nil {
			return nil, err
		}
	}

	return ordered, nil
}
