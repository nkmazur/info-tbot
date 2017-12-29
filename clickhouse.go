package main

import (
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
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
		log.WithFields(log.Fields{
			"Error": err,
		}).Error("Can't select from clickhouse")
		return 0, fmt.Errorf("Can't select from clickhouse - %v\n", err)
	}
	for result.Next() {
		result.Scan(&count)
	}
	return count, nil
}

func getHistoryFromClickhouse(id string) (string, error) {

	var history []struct {
		Created   time.Time `db: "created"`
		Kind      string    `db: "kind"`
		Namespace string    `db: "kind"`
		Name      string    `db: "kind"`
		Method    string    `db: "kind"`
	}

	var info string
	history, err := svc.clickhouse.Query("SELECT created,kind,namespace,name,method FROM results WHERE user_id=? limit 10", id)
	if err != nil {
		return "", err
	}

	for result.Next() {
		result.Scan(&info)
	}
	fmt.Println(info)
	return info, err
}
