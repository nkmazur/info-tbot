package main

import (
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"

	"fmt"
)

type UserInfo struct {
	UserId      string `db:"user_id"`
	NamespaceId string `db:"namespace_id"`
	Label       string `db:"label"`
}

type UserID struct {
	Id string `db:"id"`
}

func GetUserInfo(mail string) ([]UserInfo, error) {
	var ns []UserInfo

	if mail == "" {
		//		log.Info("Empty email in '/info' command")
		return nil, fmt.Errorf("Empty email not good\n")
	}

	query := `SELECT users.id AS user_id, namespaces.id AS namespace_id, namespaces.label AS label FROM users
		LEFT JOIN namespaces ON users.id=namespaces.user_id  WHERE users.email=$1 AND namespaces.active=TRUE`
	if err := svc.postgres.Select(&ns, query, mail); err != nil {
		log.WithFields(log.Fields{
			"query": query,
		}).Error("Unable to select from Postgresql")
		return nil, fmt.Errorf("Can't select from pg - %v\n", err)
	}

	if len(ns) < 1 {
		//		log.WithFields(log.Fields{
		//			"user": mail,
		//		}).Info("No such user in database")
		return nil, fmt.Errorf("No such user\n")
	}

	return ns, nil
}

func GetUserID(mail string) (string, error) {
	if mail == "" {
		return "", fmt.Errorf("Empty email not good\n")
	}
	id := []UserID{}
	query := `SELECT id FROM users WHERE email=$1`
	if err := svc.postgres.Select(&id, query, mail); err != nil {
		log.WithFields(log.Fields{
			"query": query,
		}).Error("Unable to select from Postgresql")
		return "", fmt.Errorf("Can't select from pg - %v\n", err)
	}
	if len(id) == 0 {
		return "", fmt.Errorf("No such user!")
	}


	return id[0].Id, nil
}
