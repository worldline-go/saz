package database

import (
	"database/sql"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"github.com/worldline-go/saz/internal/render"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/types"
)

func ScanSlice(columnsLen int, r *sql.Rows) ([]any, error) {
	values := make([]any, columnsLen)
	for i := range values {
		values[i] = new(any)
	}

	if err := r.Scan(values...); err != nil {
		return nil, err
	}

	for i := range columnsLen {
		values[i] = *(values[i].(*any))
	}

	return values, nil
}

func ScanSliceWithValues(columnsLen int, r *sql.Rows, valueTypes []any) ([]any, error) {
	if len(valueTypes) != columnsLen {
		return nil, fmt.Errorf("values length %d does not match columns length %d", len(valueTypes), columnsLen)
	}

	values := slices.Clone(valueTypes)

	if err := r.Scan(values...); err != nil {
		return nil, err
	}

	for i := range values {
		if _, ok := (values[i].(*any)); ok {
			values[i] = *(values[i].(*any))
		}
	}

	return values, nil
}

func GenerateSlice(columnTypes []*sql.ColumnType, mapType service.MapType) []any {
	dynamicFields := make([]any, 0, len(columnTypes))
	for _, col := range columnTypes {
		dynamicFields = append(dynamicFields, GetType(col, mapType))
	}

	return dynamicFields
}

func GetType(col *sql.ColumnType, mapType service.MapType) any {
	if mapType.Enabled {
		if colMapType, ok := mapType.Column[col.Name()]; ok {
			switch colMapType.Type {
			case "string":
				if colMapType.Nullable {
					return new(types.Null[string])
				}
				return new(string)
			case "number":
				if colMapType.Nullable {
					return new(types.NullDecimal)
				}
				return new(types.Decimal)
			case "date":
				if colMapType.Nullable {
					return new(types.Null[types.Time])
				}
				return new(types.Time)
			}
		}
	}

	switch col.ScanType().Kind() {
	case reflect.Float64, reflect.Int64, reflect.Uint64,
		reflect.Float32, reflect.Int32, reflect.Uint32,
		reflect.Int8, reflect.Uint8,
		reflect.Int, reflect.Uint:
		if nullable, _ := col.Nullable(); nullable {
			return new(types.NullDecimal)
		}
		return new(types.Decimal)
	case reflect.String:
		if nullable, _ := col.Nullable(); nullable {
			return new(types.Null[string])
		}
		return new(string)
	case reflect.Bool:
		if nullable, _ := col.Nullable(); nullable {
			return new(types.Null[bool])
		}
		return new(bool)
	}

	return new(any)
}

func Map(columnsIndex map[string]int, mapType service.MapType, values []any) error {
	for i := range values {
		// check if interface is string than sanitize utf8
		switch v := values[i].(type) {
		case *string:
			if v == nil {
				continue
			}

			sanitized := SanitizeString(*v)
			values[i] = &sanitized

			continue
		case *types.Null[string]:
			if v == nil {
				continue
			}

			if v.Valid {
				v.V = SanitizeString(v.V)
				values[i] = v

				continue
			}
		}
	}

	if err := mapDestination(columnsIndex, mapType, values); err != nil {
		return err
	}

	return nil
}

func mapDestination(columnsIndex map[string]int, mapType service.MapType, result []any) error {
	if !mapType.Enabled {
		return nil
	}

	for k, colType := range mapType.Destination {
		if idx, ok := columnsIndex[k]; ok {
			switch colType.Type {
			case "string":
				vStr, err := getAnyString(result[idx], colType)
				if err != nil {
					return err
				}

				if colType.Nullable {
					result[idx] = vStr
				} else {
					result[idx] = vStr.V
				}
			case "number":
				vNum, err := getAnyNumber(result[idx], colType)
				if err != nil {
					return err
				}

				if colType.Nullable {
					result[idx] = vNum
				} else {
					result[idx] = vNum.Decimal
				}
			case "date":
				vDate, err := getAnyDate(result[idx], colType)
				if err != nil {
					return err
				}

				if colType.Nullable {
					result[idx] = vDate
				} else {
					result[idx] = vDate.V
				}
			}
		}
	}

	return nil
}

func getAnyString(v any, t service.ColumnTypeTemplate) (types.Null[string], error) {
	switch val := v.(type) {
	case string:
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val)
			if err != nil {
				return types.Null[string]{}, err
			}
			return types.NewNull(string(vRendered)), nil
		}
		return types.NewNull(val), nil
	case types.Null[string]:
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val.V)
			if err != nil {
				return types.Null[string]{}, err
			}
			return types.NewNullWithValid(string(vRendered), val.Valid), nil
		}
		return val, nil
	case []byte:
		var v string
		if t.Encoding.Enabled {
			var err error
			switch t.Encoding.Coding {
			case EncodingISO88591:
				v, err = ConvertISO88591ToUTF8(val)
				if err != nil {
					return types.Null[string]{}, err
				}
			default:
				v = string(val)
			}
		} else {
			v = string(val)
		}

		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, v)
			if err != nil {
				return types.Null[string]{}, err
			}
			return types.NewNull(string(vRendered)), nil
		}

		return types.NewNull(v), nil
	case nil:
		if t.Nullable {
			return types.NewNullWithValid("", false), nil
		}
	}

	if t.Template.Enabled {
		vRendered, err := render.ExecuteWithData(t.Template.Value, cast.ToString(v))
		if err != nil {
			return types.Null[string]{}, err
		}
		return types.NewNull(string(vRendered)), nil
	}
	return types.NewNull(cast.ToString(v)), nil
}

func getAnyNumber(v any, t service.ColumnTypeTemplate) (types.NullDecimal, error) {
	switch val := v.(type) {
	case types.Decimal:
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val.String())
			if err != nil {
				return types.NullDecimal{}, err
			}
			decimalVal, err := decimal.NewFromString(string(vRendered))
			if err != nil {
				return types.NullDecimal{}, err
			}
			return types.NullDecimal{Decimal: decimalVal, Valid: true}, nil
		}
		return types.NullDecimal{Decimal: val, Valid: true}, nil
	case types.NullDecimal:
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val.Decimal.String())
			if err != nil {
				return types.NullDecimal{}, err
			}
			decimalVal, err := decimal.NewFromString(string(vRendered))
			if err != nil {
				return types.NullDecimal{}, err
			}
			return types.NullDecimal{Decimal: decimalVal, Valid: val.Valid}, nil
		}
		return val, nil
	case nil:
		if t.Nullable {
			return types.NullDecimal{Valid: false}, nil
		}
	}

	if t.Template.Enabled {
		vRendered, err := render.ExecuteWithData(t.Template.Value, cast.ToString(v))
		if err != nil {
			return types.NullDecimal{}, err
		}
		decimalVal, err := decimal.NewFromString(string(vRendered))
		if err != nil {
			return types.NullDecimal{}, err
		}
		return types.NullDecimal{Decimal: decimalVal, Valid: true}, nil
	}
	decimalVal, err := decimal.NewFromString(cast.ToString(v))
	if err != nil {
		return types.NullDecimal{}, err
	}
	return types.NullDecimal{Decimal: decimalVal, Valid: true}, nil
}

func getAnyDate(v any, t service.ColumnTypeTemplate) (types.Null[types.Time], error) {
	switch val := v.(type) {
	case time.Time:
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val.Format(time.RFC3339))
			if err != nil {
				return types.Null[types.Time]{}, err
			}

			var parsedTime types.Time
			if err := parsedTime.Parse(string(vRendered)); err != nil {
				return types.Null[types.Time]{}, err
			}
			return types.NewTimeNullWithValid(parsedTime.Time, true), nil
		}
		return types.NewTimeNullWithValid(val, true), nil
	case types.Time:
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val.String())
			if err != nil {
				return types.Null[types.Time]{}, err
			}

			var parsedTime types.Time
			if err := parsedTime.Parse(string(vRendered)); err != nil {
				return types.Null[types.Time]{}, err
			}
			return types.NewTimeNullWithValid(parsedTime.Time, true), nil
		}
		return types.NewTimeNullWithValid(val.Time, true), nil
	case types.Null[types.Time]:
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val.V.String())
			if err != nil {
				return types.Null[types.Time]{}, err
			}

			var parsedTime types.Time
			if err := parsedTime.Parse(string(vRendered)); err != nil {
				return types.Null[types.Time]{}, err
			}
			return types.NewTimeNullWithValid(parsedTime.Time, val.Valid), nil
		}
		return val, nil
	case string:
		var parsedTime types.Time
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val)
			if err != nil {
				return types.Null[types.Time]{}, err
			}

			if err := parsedTime.Parse(string(vRendered)); err != nil {
				return types.Null[types.Time]{}, err
			}
			return types.NewTimeNullWithValid(parsedTime.Time, true), nil
		}

		if err := parsedTime.Parse(val); err != nil {
			return types.Null[types.Time]{}, err
		}
		return types.NewTimeNullWithValid(parsedTime.Time, true), nil
	case types.Null[string]:
		if !val.Valid {
			if t.Nullable {
				return types.NewTimeNullWithValid(time.Time{}, false), nil
			}
		}

		var parsedTime types.Time
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, val.V)
			if err != nil {
				return types.Null[types.Time]{}, err
			}
			if err := parsedTime.Parse(string(vRendered)); err != nil {
				return types.Null[types.Time]{}, err
			}
			return types.NewTimeNullWithValid(parsedTime.Time, val.Valid), nil
		}
		if err := parsedTime.Parse(val.V); err != nil {
			return types.Null[types.Time]{}, err
		}
		return types.NewTimeNullWithValid(parsedTime.Time, val.Valid), nil
	case []byte:
		var v string
		if t.Encoding.Enabled {
			var err error
			switch t.Encoding.Coding {
			case EncodingISO88591:
				v, err = ConvertISO88591ToUTF8(val)
				if err != nil {
					return types.Null[types.Time]{}, err
				}
			default:
				v = string(val)
			}
		} else {
			v = string(val)
		}

		var parsedTime types.Time
		if t.Template.Enabled {
			vRendered, err := render.ExecuteWithData(t.Template.Value, v)
			if err != nil {
				return types.Null[types.Time]{}, err
			}
			if err := parsedTime.Parse(string(vRendered)); err != nil {
				return types.Null[types.Time]{}, err
			}
			return types.NewTimeNullWithValid(parsedTime.Time, true), nil
		}
		if err := parsedTime.Parse(v); err != nil {
			return types.Null[types.Time]{}, err
		}
		return types.NewTimeNullWithValid(parsedTime.Time, true), nil
	case nil:
		if t.Nullable {
			return types.NewTimeNullWithValid(time.Time{}, false), nil
		}
	}

	// Default case: try to convert to string and parse
	var parsedTime types.Time
	if t.Template.Enabled {
		vRendered, err := render.ExecuteWithData(t.Template.Value, cast.ToString(v))
		if err != nil {
			return types.Null[types.Time]{}, err
		}
		if err := parsedTime.Parse(string(vRendered)); err != nil {
			return types.Null[types.Time]{}, err
		}
		return types.NewTimeNullWithValid(parsedTime.Time, true), nil
	}
	if err := parsedTime.Parse(cast.ToString(v)); err != nil {
		return types.Null[types.Time]{}, err
	}
	return types.NewTimeNullWithValid(parsedTime.Time, true), nil
}
