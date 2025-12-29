package service

func DataToMap(columns []string, data [][]any) []map[string]any {
	result := make([]map[string]any, 0, len(data))
	for _, row := range data {
		rowMap := make(map[string]any, len(columns))
		for i, col := range columns {
			if i < len(row) {
				rowMap[col] = row[i]
			} else {
				rowMap[col] = nil
			}
		}

		result = append(result, rowMap)
	}

	return result
}
