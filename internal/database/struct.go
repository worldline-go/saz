package database

import (
	"database/sql"
	"reflect"
)

func GenerateStruct(columnTypes []*sql.ColumnType) []reflect.StructField {
	dynamicFields := make([]reflect.StructField, 0, len(columnTypes))
	for _, col := range columnTypes {
		field := reflect.StructField{
			Name: col.Name(),
			Type: reflect.TypeOf(col.ScanType()),
			Tag:  reflect.StructTag(`db:"` + col.Name() + `"`),
		}

		dynamicFields = append(dynamicFields, field)
	}

	return dynamicFields
}

// func GetStructType(t reflect.Type) reflect.Type {

// }
