package main

import (
	_ "github.com/lib/pq"

	"fmt"
)

type UserInfo struct {
	UserId      string `db:"user_id"`
	NamespaceId string `db:"namespace_id"`
	Label       string `db:"label"`
}

func GetUserInfo(mail string) ([]UserInfo, error) {
	var ns []UserInfo

	if mail == "" {
		return nil, fmt.Errorf("Empty email not good\n")
	}

	query := `SELECT users.id AS user_id, namespaces.id AS namespace_id namespaces.label AS label FROM user
		LEFT JOIN namespaces ON users.id=namespaces.user_id  WHERE users.email=$1 AND namespaces.active=TRUE`
	if err := svc.postgres.Select(&ns, query, mail); err != nil {
		return nil, fmt.Errorf("Can't select from pg - %v\n", err)
	}

	if len(ns) < 1 {
		return nil, fmt.Errorf("No such user\n")
	}

	return ns, nil
}
