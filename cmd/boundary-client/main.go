/*
 * @Author: F1
 * @Date: 2022-03-29 16:32:24
 * @LastEditTime: 2022-03-29 16:48:54
 * @FilePath: /boundary/cmd/boundary-client/main.go
 * @Description:
 *
 * Copyright (c) 2022 by splashtop.com, All Rights Reserved.
 */
package main

import (
	"math/rand"
	"os"
	"time"

	"github.com/hashicorp/boundary/client"
)

func main() {
	rand.Seed(time.Now().UnixNano())
	os.Exit(client.Run(os.Args[1:]))
}
