package main

import (
	"encoding/json"
	"fmt"
	"time"

	log "github.com/sirupsen/logrus"
)

type historyItem struct {
	Created   time.Time `db:"created"`
	Kind      string    `db:"kind"`
	Namespace string    `db:"namespace"`
	Name      string    `db:"name"`
	Method    string    `db:"method"`
	Data      string    `db:"data"`
}

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

	var history []historyItem
	var text string
	err := svc.clickhouse.Select(&history, "SELECT created,kind,namespace,name,method,data FROM results WHERE user_id=? AND (method='create' OR method='delete') limit 20", id)
	if err != nil {
		log.Error(err)
		return "", err
	}
	for _, event := range history {
		if event.Method == "create" {
			if event.Kind == "deployments" {
				var data []struct {
					Data struct {
						Spec struct {
							Template struct {
								Spec struct {
									Containers []struct {
										Image string `json:"image"`
									} `json:"containers"`
								} `json:"spec"`
							} `json:"template"`
						} `json:"spec"`
						Metadata struct {
							Name string `json:"name"`
						} `json:"metadata"`
					} `json:"data"`
				}
				err := json.Unmarshal([]byte(event.Data), &data)
				if err != nil {
					log.Error("Unable to parse json from clickhouse!", err)
				}
				if event.Data != "null" {
					text += fmt.Sprintf("%v create deploy \"%v\" in ns \"%v\", image \"%v\"\n", event.Created.Format(time.RFC822), data[0].Data.Metadata.Name, event.Namespace, data[0].Data.Spec.Template.Spec.Containers[0].Image)
				} else {
					text += fmt.Sprintf("%v create deploy in ns \"%v\" NO JSON IN DATABASE!!!\n", event.Created.Format(time.RFC822), event.Namespace)
					log.WithFields(log.Fields{
						"JSON":   event.Data,
						"Kind":   event.Kind,
						"Method": event.Method,
					}).Error("No JSON in database!!!")
				}
			} else { // END OF DEPLOYMENTS SECTION
				var data []struct {
					Data struct {
						Metadata struct {
							Name string `json:"name"`
						} `json:"metadata"`
					} `json:"data"`
				}
				err := json.Unmarshal([]byte(event.Data), &data)
				if err != nil {
					log.Error("Unable to parse json from clickhouse!", err)
				}
				if event.Data != "null" {
					text += fmt.Sprintf("%v, create %v \"%v\" in ns: \"%v\"\n", event.Created.Format(time.RFC822), event.Kind, data[0].Data.Metadata.Name, event.Namespace)
				} else {
					text += fmt.Sprintf("%v create %v in ns \"%v\" JSON NOT FOUND IN DATABASE!!!\n", event.Created.Format(time.RFC822), event.Kind, event.Namespace)
					log.WithFields(log.Fields{
						"JSON":   event.Data,
						"Kind":   event.Kind,
						"Method": event.Method,
					}).Error("No JSON in database!!!")
				}
			}
		} else { //END OF CREATE SECTION
			text += fmt.Sprintf("%v, delete %v \"%v\" in ns: \"%v\"\n", event.Created.Format(time.RFC822), event.Kind, event.Name, event.Namespace)
		}
	}
	return text, nil
}
