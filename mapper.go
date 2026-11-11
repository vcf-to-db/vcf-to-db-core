package core

import "fmt"

// TransformFunc transforms a value during mapping.
type TransformFunc func(Value) (Value, error)

// MappedRow represents one row to be inserted into a target table.
type MappedRow struct {
	Table  string
	Record Record
}

// MapRecord transforms a source record according to a mapping definition.
//
// The mapping must have passed ValidateMapping before this function is called.
// Collection mappings produce one row per collection item.
// MapRecord does not perform database operations.
func MapRecord(
	record Record,
	mapping Mapping,
	transforms map[string]TransformFunc,
) ([]MappedRow, error) {
	rows := make([]MappedRow, 0, len(mapping.Tables))

	for _, tableMapping := range mapping.Tables {
		if tableMapping.ParentReference != nil {
			return nil, fmt.Errorf("%w: parent references are not implemented",
				ErrUnsupportedFeature)
		}

		sourceRecords, err := resolveSourceRecords(record, tableMapping.SourceCollection)
		if err != nil {
			return nil, err
		}

		for i, sourceRecord := range sourceRecords {
			targetRecord, err := mapTableRecord(sourceRecord, tableMapping, transforms)
			if err != nil {
				if tableMapping.SourceCollection != "" {
					return nil, fmt.Errorf("collection %q item %d: %w",
						tableMapping.SourceCollection, i, err)
				}
				return nil, err
			}

			rows = append(rows, MappedRow{
				Table:  tableMapping.TargetTable,
				Record: targetRecord,
			})
		}
	}

	return rows, nil
}

func resolveSourceRecords(record Record, sourceCollection string) ([]Record, error) {
	if sourceCollection == "" {
		return []Record{record}, nil
	}

	value, exists := record[sourceCollection]
	if !exists {
		return nil, fmt.Errorf("%w: source collection %q is missing",
			ErrInvalidRecord, sourceCollection)
	}

	values, ok := value.([]Value)
	if !ok {
		return nil, fmt.Errorf("%w: source collection %q is not a repeated value",
			ErrInvalidRecord, sourceCollection)
	}

	records := make([]Record, 0, len(values))

	for i, item := range values {
		fields, ok := item.(map[string]Value)
		if !ok {
			return nil, fmt.Errorf("%w: source collection %q item %d is not a structured value",
				ErrInvalidRecord, sourceCollection, i)
		}

		records = append(records, Record(fields))
	}

	return records, nil
}

func mapTableRecord(
	sourceRecord Record,
	tableMapping TableMapping,
	transforms map[string]TransformFunc,
) (Record, error) {
	targetRecord := make(Record, len(tableMapping.Fields))

	for _, field := range tableMapping.Fields {
		value, err := resolveSourceValue(sourceRecord, field)
		if err != nil {
			return nil, err
		}

		if field.Transform != "" {
			transform, exists := transforms[field.Transform]
			if !exists {
				return nil, fmt.Errorf("%w: unknown transformation %q",
					ErrInvalidMapping, field.Transform)
			}

			transformed, err := transform(value)
			if err != nil {
				return nil, fmt.Errorf("%w: transformation %q failed for field %q: %v",
					ErrInvalidRecord, field.Transform, field.SourceField, err)
			}

			value = transformed
		}

		targetRecord[field.TargetField] = value
	}

	return targetRecord, nil
}

func resolveSourceValue(record Record, field FieldMapping) (Value, error) {
	value, exists := record[field.SourceField]
	if !exists {
		return fieldDefaultValue(field)
	}

	values, repeated := value.([]Value)
	if !repeated {
		if field.Occurrence != 0 {
			return fieldDefaultValue(field)
		}
		return value, nil
	}

	if field.Occurrence >= len(values) {
		return fieldDefaultValue(field)
	}

	return values[field.Occurrence], nil
}

func fieldDefaultValue(field FieldMapping) (Value, error) {
	if field.Default != nil {
		return *field.Default, nil
	}

	return nil, fmt.Errorf("%w: source field %q has no value at occurrence %d and has no default",
		ErrInvalidRecord, field.SourceField, field.Occurrence)
}
