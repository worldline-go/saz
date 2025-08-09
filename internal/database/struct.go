package database

import (
	"database/sql"
	"reflect"
	"strings"

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

func Struct2Map(v any) map[string]any {
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
			result[field.Tag.Get("db")] = value.Interface()
		}
	}

	return result
}
