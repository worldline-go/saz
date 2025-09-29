package database

import (
	"fmt"
	"strings"
)

func QueryBuilder(table string, columns []string, placeHolder string) func(batchCount int) string {
	return func(batchCount int) string {
		queryBuilderValues := strings.Builder{}
		for batchIndex := range batchCount {
			queryBuilderValues.WriteString("(")
			if placeHolder == "$" || placeHolder == ":" {
				for i := range (len(columns)) - 1 {
					queryBuilderValues.WriteString(fmt.Sprintf("%s%d,", placeHolder, (batchIndex*len(columns))+i+1))
				}

				queryBuilderValues.WriteString(fmt.Sprintf("%s%d", placeHolder, (batchIndex+1)*len(columns)))
			} else {
				queryBuilderValues.WriteString(strings.Repeat("?,", (len(columns))-1))
				queryBuilderValues.WriteString("?")
			}

			queryBuilderValues.WriteString(")")

			if batchIndex < batchCount-1 {
				queryBuilderValues.WriteString(", ")
			}
		}

		queryBuilder := strings.Builder{}

		queryBuilder.WriteString("INSERT INTO ")
		queryBuilder.WriteString(table)
		queryBuilder.WriteString(" (")
		queryBuilder.WriteString(strings.Join(columns, ","))
		queryBuilder.WriteString(") VALUES ")
		queryBuilder.WriteString(queryBuilderValues.String())

		return queryBuilder.String()
	}
}
