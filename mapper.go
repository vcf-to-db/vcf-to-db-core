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
// Collection expansion and parent references are not supported yet.
// MapRecord does not perform database operations.
func MapRecord(
	record Record,
	mapping Mapping,
	transforms map[string]TransformFunc,
) ([]MappedRow, error) {
	rows := make([]MappedRow, 0, len(mapping.Tables))

	for _, tableMapping := range mapping.Tables {
		if tableMapping.SourceCollection != "" {
			return nil, fmt.Errorf("%w: collection mapping is not implemented",
				ErrUnsupportedFeature)
		}

		if tableMapping.ParentReference != nil {
			return nil, fmt.Errorf("%w: parent references are not implemented",
				ErrUnsupportedFeature)
		}

		targetRecord := make(Record, len(tableMapping.Fields))

		for _, field := range tableMapping.Fields {
			value, exists := record[field.SourceField]

			if !exists {
				if field.Default == nil {
					return nil, fmt.Errorf("%w: source field %q is missing and has no default",
						ErrInvalidRecord, field.SourceField)
				}

				value = *field.Default
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

		rows = append(rows, MappedRow{
			Table:  tableMapping.TargetTable,
			Record: targetRecord,
		})
	}

	return rows, nil
}
