package database

import (
	"database/sql"
	"reflect"
	"strings"
	"unicode/utf8"

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

func Struct2Map(v any, mapType service.MapType) map[string]any {
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
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

	return mapDestination(result, mapType)
}

func mapDestination(result map[string]any, mapType service.MapType) map[string]any {
	if !mapType.Enabled {
		return result
	}

	mappedResult := make(map[string]any, len(result))
	for k, v := range result {
		if colType, ok := mapType.Destination[k]; ok {
			switch colType.Type {
			case "string":
				vStr := getAnyString(v)
				if colType.Nullable {
					mappedResult[k] = vStr
				} else {
					mappedResult[k] = vStr.V
				}
			case "number":
				vNum := getAnyNumber(v)
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

	return mappedResult
}

func getAnyString(v any) types.Null[string] {
	if v == nil {
		return types.NewNullWithValid("", false)
	}

	switch val := v.(type) {
	case string:
		return types.NewNull(val)
	case types.Null[string]:
		return val
	default:
		return types.NewNull(cast.ToString(v))
	}
}

func getAnyNumber(v any) types.NullDecimal {
	if v == nil {
		return types.NullDecimal{Valid: false}
	}

	switch val := v.(type) {
	case types.Decimal:
		return types.NullDecimal{Decimal: val, Valid: true}
	case types.NullDecimal:
		return val
	default:
		return types.NullDecimal{Decimal: decimal.RequireFromString(cast.ToString(v)), Valid: true}
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
