package database

import (
	"reflect"
	"testing"
)

func TestQueryBuilder(t *testing.T) {
	type args struct {
		table       string
		columns     []string
		placeHolder string
		batchCount  int
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "Single Column",
			args: args{
				table:       "users",
				columns:     []string{"id"},
				placeHolder: "$",
				batchCount:  1,
			},
			want: "INSERT INTO users (id) VALUES ($1)",
		},
		{
			name: "Multiple Columns",
			args: args{
				table:       "users",
				columns:     []string{"id", "name"},
				placeHolder: "$",
				batchCount:  1,
			},
			want: "INSERT INTO users (id,name) VALUES ($1,$2)",
		},
		{
			name: "Batch Insert",
			args: args{
				table:       "users",
				columns:     []string{"id", "name"},
				placeHolder: "$",
				batchCount:  2,
			},
			want: "INSERT INTO users (id,name) VALUES ($1,$2), ($3,$4)",
		},
		{
			name: "Single Column ?",
			args: args{
				table:       "users",
				columns:     []string{"id"},
				placeHolder: "?",
				batchCount:  1,
			},
			want: "INSERT INTO users (id) VALUES (?)",
		},
		{
			name: "Multiple Columns ?",
			args: args{
				table:       "users",
				columns:     []string{"id", "name"},
				placeHolder: "?",
				batchCount:  1,
			},
			want: "INSERT INTO users (id,name) VALUES (?,?)",
		},
		{
			name: "Batch Insert ?",
			args: args{
				table:       "users",
				columns:     []string{"id", "name"},
				placeHolder: "?",
				batchCount:  2,
			},
			want: "INSERT INTO users (id,name) VALUES (?,?), (?,?)",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := QueryBuilder(tt.args.table, tt.args.columns, tt.args.placeHolder)

			if result := got(tt.args.batchCount); !reflect.DeepEqual(result, tt.want) {
				t.Errorf("QueryBuilder() = %v, want %v", result, tt.want)
			}
		})
	}
}
