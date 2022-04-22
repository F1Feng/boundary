/*
 * @Author: F1
 * @Date: 2022-03-30 14:52:04
 * @LastEditTime: 2022-04-22 15:24:43
 * @FilePath: /boundary/client/connect.go
 * @Description:
 *
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

func Connect(authzToken string, scopeName string, scopeId string, listenAddr string, listenPort string) (connectId string, err error) {
	targetScope := "-target-scope-name"
	uid, _ := uuid.NewUUID()
	if len(scopeName) == 0 && len(scopeId) == 0 {
		return connectId, fmt.Errorf("invalid param scopeName or scopeId can't empty")
	}
	if len(scopeId) > 0 {
		targetScope = "-target-scope-id"
	}
	connectId = uid.String()

	listenAddrStr := ""
	if len(listenAddr) > 0 {
		listenAddrStr = fmt.Sprintf("-listen-addr=%s", listenAddr)
	}
	listenPortStr := ""
	if len(listenPort) > 0 {
		listenPortStr = fmt.Sprintf("-listen-port=%s", listenPort)
	}

	args := []string{
		connectId,
		"connect",
		"-authz-token",
		authzToken,
		targetScope,
		scopeId,
		"-format=",
	}

	if len(listenAddrStr) > 0 {
		args = append(args, listenAddrStr)
	}
	if len(listenPortStr) > 0 {
		args = append(args, listenPortStr)
	}
	connectto(args)

	return connectId, nil
}

func DisConnect(connectId string) error {
	return disconnect(connectId)
}
