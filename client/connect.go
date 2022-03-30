/*
 * @Author: F1
 * @Date: 2022-03-30 14:52:04
 * @LastEditTime: 2022-03-30 14:55:54
 * @FilePath: /client/connect.go
 * @Description:
 *
 * Copyright (c) 2022 by splashtop.com, All Rights Reserved.
 */
package client

import (
	"fmt"

	"github.com/google/uuid"
)

/**
 * @description:
 *
 *		according boundary client connect to
 *
 * @param {[]string} args
 * @return {*}
 */
func connectto(args []string) {
	Run(args)
}

func Connect(authzToken string, scopeName string, scopeId string) (connectId string, err error) {
	targetScope := "-target-scope-name"
	uid, _ := uuid.NewUUID()
	if len(scopeName) == 0 && len(scopeId) == 0 {
		return connectId, fmt.Errorf("invalid param scopeName or scopeId can't empty")
	}
	if len(scopeId) > 0 {
		targetScope = "-target-scope-id"
	}
	connectId = uid.String()
	connectto([]string{
		connectId,
		"connect",
		"-authz-token",
		authzToken,
		targetScope,
		scopeId,
	})

	return connectId, nil
}

func DisConnect(connectId string) error {
	return disconnect(connectId)
}
