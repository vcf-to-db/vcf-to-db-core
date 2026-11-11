package core

import (
	"fmt"
	"strings"
)

// ValidateSchema checks the logical structure of a schema.
func ValidateSchema(schema Schema) error {
	if len(schema.Tables) == 0 {
		return fmt.Errorf("%w: schema must contain at least one table", ErrInvalidSchema)
	}

	tables := make(map[string]Table, len(schema.Tables))

	for _, table := range schema.Tables {
		if strings.TrimSpace(table.Name) == "" {
			return fmt.Errorf("%w: table name cannot be empty", ErrInvalidSchema)
		}

		if _, exists := tables[table.Name]; exists {
			return fmt.Errorf("%w: duplicate table %q", ErrInvalidSchema, table.Name)
		}

		tables[table.Name] = table

		if len(table.Columns) == 0 {
			return fmt.Errorf("%w: table %q has no columns", ErrInvalidSchema, table.Name)
		}

		columns := make(map[string]Column, len(table.Columns))

		for _, column := range table.Columns {
			if strings.TrimSpace(column.Name) == "" {
				return fmt.Errorf("%w: empty column name in table %q", ErrInvalidSchema, table.Name)
			}

			if _, exists := columns[column.Name]; exists {
				return fmt.Errorf("%w: duplicate column %q in table %q",
					ErrInvalidSchema, column.Name, table.Name)
			}

			if !isValidDataType(column.Type) {
				return fmt.Errorf("%w: unsupported data type %q for %s.%s",
					ErrInvalidSchema, column.Type, table.Name, column.Name)
			}

			if !isValidGenerationStrategy(column.Generation) {
				return fmt.Errorf("%w: unsupported generation strategy %q for %s.%s",
					ErrInvalidSchema, column.Generation, table.Name, column.Name)
			}

			columns[column.Name] = column
		}

		if err := validateColumnReferences(table.Name, columns, table.PrimaryKey); err != nil {
			return err
		}

		for _, uniqueKey := range table.UniqueKeys {
			if err := validateColumnReferences(table.Name, columns, uniqueKey); err != nil {
				return err
			}
		}
	}

	for _, table := range schema.Tables {
		for _, foreignKey := range table.ForeignKeys {
			referencedTable, exists := tables[foreignKey.ReferencedTable]
			if !exists {
				return fmt.Errorf("%w: table %q references unknown table %q",
					ErrInvalidSchema, table.Name, foreignKey.ReferencedTable)
			}

			if len(foreignKey.Columns) == 0 ||
				len(foreignKey.Columns) != len(foreignKey.ReferencedColumns) {
				return fmt.Errorf("%w: invalid foreign key in table %q",
					ErrInvalidSchema, table.Name)
			}

			sourceColumns := columnNames(table.Columns)
			targetColumns := columnNames(referencedTable.Columns)

			for _, name := range foreignKey.Columns {
				if _, exists := sourceColumns[name]; !exists {
					return fmt.Errorf("%w: foreign key references unknown column %q in table %q",
						ErrInvalidSchema, name, table.Name)
				}
			}

			for _, name := range foreignKey.ReferencedColumns {
				if _, exists := targetColumns[name]; !exists {
					return fmt.Errorf("%w: foreign key references unknown column %q in table %q",
						ErrInvalidSchema, name, foreignKey.ReferencedTable)
				}
			}
		}
	}

	return nil
}

// ValidateMapping checks whether mappings reference existing target fields
// and describe valid parent-child relationships.
func ValidateMapping(mapping Mapping, schema Schema) error {
	if err := ValidateSchema(schema); err != nil {
		return err
	}

	schemaTables := make(map[string]Table, len(schema.Tables))
	for _, table := range schema.Tables {
		schemaTables[table.Name] = table
	}

	mappedTables := make(map[string]TableMapping, len(mapping.Tables))

	// First pass: validate target tables and reject duplicate mappings.
	for _, tableMapping := range mapping.Tables {
		if strings.TrimSpace(tableMapping.TargetTable) == "" {
			return fmt.Errorf("%w: target table cannot be empty", ErrInvalidMapping)
		}

		if _, exists := schemaTables[tableMapping.TargetTable]; !exists {
			return fmt.Errorf("%w: unknown target table %q",
				ErrInvalidMapping, tableMapping.TargetTable)
		}

		if _, exists := mappedTables[tableMapping.TargetTable]; exists {
			return fmt.Errorf("%w: duplicate mapping for table %q",
				ErrInvalidMapping, tableMapping.TargetTable)
		}

		if tableMapping.SourceCollection != "" &&
			strings.TrimSpace(tableMapping.SourceCollection) == "" {
			return fmt.Errorf("%w: source collection cannot be whitespace in table %q",
				ErrInvalidMapping, tableMapping.TargetTable)
		}

		mappedTables[tableMapping.TargetTable] = tableMapping
	}

	// Second pass: validate fields and parent relationships.
	for _, tableMapping := range mapping.Tables {
		table := schemaTables[tableMapping.TargetTable]
		columns := columnNames(table.Columns)

		seenFields := make(map[string]struct{}, len(tableMapping.Fields))

		for _, field := range tableMapping.Fields {
			if strings.TrimSpace(field.SourceField) == "" {
				return fmt.Errorf("%w: source field cannot be empty in table %q",
					ErrInvalidMapping, tableMapping.TargetTable)
			}

			if field.Occurrence < 0 {
				return fmt.Errorf("%w: occurrence cannot be negative for field %q",
					ErrInvalidMapping, field.SourceField)
			}

			if strings.TrimSpace(field.TargetField) == "" {
				return fmt.Errorf("%w: target field cannot be empty in table %q",
					ErrInvalidMapping, tableMapping.TargetTable)
			}

			if _, exists := columns[field.TargetField]; !exists {
				return fmt.Errorf("%w: unknown target field %q in table %q",
					ErrInvalidMapping, field.TargetField, tableMapping.TargetTable)
			}

			if _, exists := seenFields[field.TargetField]; exists {
				return fmt.Errorf("%w: duplicate target field %q in table %q",
					ErrInvalidMapping, field.TargetField, tableMapping.TargetTable)
			}

			seenFields[field.TargetField] = struct{}{}
		}

		if parent := tableMapping.ParentReference; parent != nil {
			if strings.TrimSpace(parent.ParentTable) == "" ||
				strings.TrimSpace(parent.ParentField) == "" ||
				strings.TrimSpace(parent.TargetField) == "" {
				return fmt.Errorf("%w: incomplete parent reference in table %q",
					ErrInvalidMapping, tableMapping.TargetTable)
			}

			if parent.ParentTable == tableMapping.TargetTable {
				return fmt.Errorf("%w: table %q cannot reference itself as its parent",
					ErrInvalidMapping, tableMapping.TargetTable)
			}

			parentTable, exists := schemaTables[parent.ParentTable]
			if !exists {
				return fmt.Errorf("%w: unknown parent table %q",
					ErrInvalidMapping, parent.ParentTable)
			}

			if _, exists := mappedTables[parent.ParentTable]; !exists {
				return fmt.Errorf("%w: parent table %q has no mapping",
					ErrInvalidMapping, parent.ParentTable)
			}

			if _, exists := columnNames(parentTable.Columns)[parent.ParentField]; !exists {
				return fmt.Errorf("%w: unknown parent field %q in table %q",
					ErrInvalidMapping, parent.ParentField, parent.ParentTable)
			}

			if _, exists := columns[parent.TargetField]; !exists {
				return fmt.Errorf("%w: unknown parent reference target field %q in table %q",
					ErrInvalidMapping, parent.TargetField, tableMapping.TargetTable)
			}

			if _, exists := seenFields[parent.TargetField]; exists {
				return fmt.Errorf("%w: parent reference target field %q must not also be explicitly mapped in table %q",
					ErrInvalidMapping, parent.TargetField, tableMapping.TargetTable)
			}

			if !hasSingleColumnForeignKey(
				table,
				parent.TargetField,
				parent.ParentTable,
				parent.ParentField,
			) {
				return fmt.Errorf("%w: %s.%s has no matching foreign key to %s.%s",
					ErrInvalidMapping,
					tableMapping.TargetTable,
					parent.TargetField,
					parent.ParentTable,
					parent.ParentField,
				)
			}

			seenFields[parent.TargetField] = struct{}{}
		}

		// Check that every required column has a value source.
		for _, column := range table.Columns {
			if column.Nullable || column.Generation == GenerationAuto {
				continue
			}

			if _, supplied := seenFields[column.Name]; !supplied {
				return fmt.Errorf("%w: required column %q in table %q has no mapping or default",
					ErrInvalidMapping, column.Name, tableMapping.TargetTable)
			}
		}
	}

	if err := validateParentDependencies(mappedTables); err != nil {
		return err
	}

	return nil
}

// hasSingleColumnForeignKey checks whether a table has a matching
// single-column foreign key.
func hasSingleColumnForeignKey(
	table Table,
	sourceColumn string,
	referencedTable string,
	referencedColumn string,
) bool {
	for _, foreignKey := range table.ForeignKeys {
		if len(foreignKey.Columns) != 1 ||
			len(foreignKey.ReferencedColumns) != 1 {
			continue
		}

		if foreignKey.Columns[0] == sourceColumn &&
			foreignKey.ReferencedTable == referencedTable &&
			foreignKey.ReferencedColumns[0] == referencedColumn {
			return true
		}
	}

	return false
}

// validateParentDependencies rejects cycles in the parent mapping graph.
func validateParentDependencies(mappedTables map[string]TableMapping) error {
	// 0 = unvisited, 1 = visiting, 2 = visited.
	state := make(map[string]uint8, len(mappedTables))

	var visit func(string) error
	visit = func(tableName string) error {
		switch state[tableName] {
		case 1:
			return fmt.Errorf("%w: cyclic parent dependency involving table %q",
				ErrInvalidMapping, tableName)
		case 2:
			return nil
		}

		state[tableName] = 1

		tableMapping := mappedTables[tableName]
		if parent := tableMapping.ParentReference; parent != nil {
			if err := visit(parent.ParentTable); err != nil {
				return err
			}
		}

		state[tableName] = 2
		return nil
	}

	for tableName := range mappedTables {
		if err := visit(tableName); err != nil {
			return err
		}
	}

	return nil
}

func validateColumnReferences(
	tableName string,
	columns map[string]Column,
	references []string,
) error {
	seen := make(map[string]struct{}, len(references))

	for _, name := range references {
		if _, exists := columns[name]; !exists {
			return fmt.Errorf("%w: unknown column %q in table %q",
				ErrInvalidSchema, name, tableName)
		}

		if _, exists := seen[name]; exists {
			return fmt.Errorf("%w: duplicate column %q in table %q",
				ErrInvalidSchema, name, tableName)
		}

		seen[name] = struct{}{}
	}

	return nil
}

func columnNames(columns []Column) map[string]struct{} {
	names := make(map[string]struct{}, len(columns))

	for _, column := range columns {
		names[column.Name] = struct{}{}
	}

	return names
}

func isValidDataType(dataType DataType) bool {
	switch dataType {
	case DataTypeString,
		DataTypeInteger,
		DataTypeDecimal,
		DataTypeBoolean,
		DataTypeTimestamp,
		DataTypeBytes:
		return true
	default:
		return false
	}
}

func isValidGenerationStrategy(strategy GenerationStrategy) bool {
	switch strategy {
	case GenerationNone, GenerationAuto:
		return true
	default:
		return false
	}
}
