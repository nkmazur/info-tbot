package main

import (
	"fmt"
)

func selectFromClickhouse(kind, method, date, queryType string) (int, error) {
	var query string
	var count int

	switch queryType {
	case "date":
		query = fmt.Sprintf("SELECT count(*) as count FROM results where method = '%v' and kind = '%v' "+
			"and created_date = toDate(?) and namespace !='default'", method, kind)
	case "last":
		query = fmt.Sprintf("SELECT count(*) as count FROM results where method = '%v' "+
			"and kind = '%v' and created_date > toDate(?) and namespace !='default'", method, kind)
	}

	result, err := svc.clickhouse.Query(query, date)
	if err != nil {
		return 0, fmt.Errorf("Can't select from clickhouse - %v\n", err)
	}
	for result.Next() {
		result.Scan(&count)
	}
	return count, nil
}
