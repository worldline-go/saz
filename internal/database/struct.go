package database

import (
	"database/sql"
	"errors"
	"reflect"
	"strings"
	"unicode/utf8"

	"github.com/rytsh/mugo/render"
	"github.com/shopspring/decimal"
	"github.com/spf13/cast"
	"github.com/worldline-go/saz/internal/service"
	"github.com/worldline-go/types"
)

func GenerateStruct(columnTypes []*sql.ColumnType, mapType service.MapType) []reflect.StructField {
	dynamicFields := make([]reflect.StructField, 0, len(columnTypes))
	for _, col := range columnTypes {
		field := reflect.StructField{
			Name: strings.ToTitle(col.Name()),
			Type: GetStructType(col, mapType),
			Tag:  reflect.StructTag(`db:"` + col.Name() + `"`),
		}

		dynamicFields = append(dynamicFields, field)
	}

	return dynamicFields
}

func GetStructType(col *sql.ColumnType, mapType service.MapType) reflect.Type {
	if mapType.Enabled {
		if colMapType, ok := mapType.Column[col.Name()]; ok {
			switch colMapType.Type {
			case "string":
				if colMapType.Nullable {
					return reflect.TypeOf(types.Null[string]{})
				}
				return reflect.TypeOf("")
			case "number":
				if colMapType.Nullable {
					return reflect.TypeOf(types.NullDecimal{})
				}
				return reflect.TypeOf(types.Decimal{})
			}
		}
	}

	switch col.ScanType().Kind() {
	case reflect.Float64, reflect.Int64, reflect.Uint64,
		reflect.Float32, reflect.Int32, reflect.Uint32,
		reflect.Int8, reflect.Uint8,
		reflect.Int, reflect.Uint:
		if nullable, _ := col.Nullable(); nullable {
			return reflect.TypeOf(types.NullDecimal{})
		}
		return reflect.TypeOf(types.Decimal{})
	case reflect.String:
		if nullable, _ := col.Nullable(); nullable {
			return reflect.TypeOf(types.Null[string]{})
		}
		return reflect.TypeOf("")
	case reflect.Bool:
		if nullable, _ := col.Nullable(); nullable {
			return reflect.TypeOf(types.Null[bool]{})
		}
		return reflect.TypeOf(false)
	}

	return col.ScanType()
}

func Struct2Map(v any, mapType service.MapType) (map[string]any, error) {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil, errors.New("input is not a struct")
	}

	result := make(map[string]any)
	for i := range val.NumField() {
		field := val.Type().Field(i)
		value := val.Field(i)

		if value.IsValid() && value.CanInterface() {
			// check if interface is string than sanitize utf8
			switch v := value.Interface().(type) {
			case string:
				result[field.Tag.Get("db")] = sanitizeString(v)
				continue
			case types.Null[string]:
				if v.Valid {
					v.V = sanitizeString(v.V)
					result[field.Tag.Get("db")] = v
					continue
				}
			}

			result[field.Tag.Get("db")] = value.Interface()
		}
	}

	result, err := mapDestination(result, mapType)
	if err != nil {
		return nil, err
	}

	return result, nil
}

func mapDestination(result map[string]any, mapType service.MapType) (map[string]any, error) {
	if !mapType.Enabled {
		return result, nil
	}

	mappedResult := make(map[string]any, len(result))
	for k, v := range result {
		if colType, ok := mapType.Destination[k]; ok {
			switch colType.Type {
			case "string":
				vStr, err := getAnyString(v, colType.Template)
				if err != nil {
					return nil, err
				}
				if colType.Nullable {
					mappedResult[k] = vStr
				} else {
					mappedResult[k] = vStr.V
				}
			case "number":
				vNum, err := getAnyNumber(v, colType.Template)
				if err != nil {
					return nil, err
				}
				if colType.Nullable {
					mappedResult[k] = vNum
				} else {
					mappedResult[k] = vNum.Decimal
				}
			default:
				mappedResult[k] = v
			}
		} else {
			mappedResult[k] = v
		}
	}

	return mappedResult, nil
}

func getAnyString(v any, t service.Template) (types.Null[string], error) {
	if v == nil {
		return types.NewNullWithValid("", false), nil
	}

	switch val := v.(type) {
	case string:
		if t.Enabled {
			vRendered, err := render.ExecuteWithData(t.Value, val)
			if err != nil {
				return types.Null[string]{}, err
			}
			return types.NewNull(string(vRendered)), nil
		}
		return types.NewNull(val), nil
	case types.Null[string]:
		if t.Enabled {
			vRendered, err := render.ExecuteWithData(t.Value, val.V)
			if err != nil {
				return types.Null[string]{}, err
			}
			return types.NewNullWithValid(string(vRendered), val.Valid), nil
		}
		return val, nil
	default:
		if t.Enabled {
			vRendered, err := render.ExecuteWithData(t.Value, cast.ToString(v))
			if err != nil {
				return types.Null[string]{}, err
			}
			return types.NewNullWithValid(string(vRendered), true), nil
		}
		return types.NewNull(cast.ToString(v)), nil
	}
}

func getAnyNumber(v any, t service.Template) (types.NullDecimal, error) {
	if v == nil {
		return types.NullDecimal{Valid: false}, nil
	}

	switch val := v.(type) {
	case types.Decimal:
		if t.Enabled {
			vRendered, err := render.ExecuteWithData(t.Value, val.String())
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
		if t.Enabled {
			vRendered, err := render.ExecuteWithData(t.Value, val.Decimal.String())
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
	default:
		if t.Enabled {
			vRendered, err := render.ExecuteWithData(t.Value, cast.ToString(v))
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
}

func sanitizeString(s string) string {
	if utf8.ValidString(s) {
		return s
	}

	var b strings.Builder
	for i, r := range s {
		if r == utf8.RuneError {
			_, size := utf8.DecodeRuneInString(s[i:])
			if size == 1 {
				// skip invalid byte
				continue
			}
		}
		b.WriteRune(r)
	}

	return b.String()
}
