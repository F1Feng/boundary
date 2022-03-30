/*
 * @Author: F1
 * @Date: 2022-03-30 15:14:45
 * @LastEditTime: 2022-03-30 17:51:05
 * @FilePath: /boundary/cmd/boundary-client/main.go
 * @Description:
 *
 * Copyright (c) 2022 by splashtop.com, All Rights Reserved.
 */

package main

import (
	"time"

	"github.com/hashicorp/boundary/client"
)

func main() {
	connectId, err := client.Connect(
		"8eBCVVb3ZbYoEQWFsqQY5kdyRdgRCoFTY7yCriSjJJm9MWuERcvhCHwmSuWaWY4LqdC5V8jr15ziW4NM6J2ZGtZKiSUkfQq6z8PHXxe6RecxqzR37MzVzmLPdXaVUppk6LVLMmZJnZjHEgUqZdyh2KHo5hmG5vvhfYNetvJMgD3b6KciDTUNAN6pq6Vjx2qNG7JNYNSY2VAWYnP6fmtckdH4XAVzDJw3vnV7W3poVow93h8YDRfZPCnFBf42QP5BxjGVviua8aSYJGrkqE3S7WdDagkrz7d2PUPxSrYH4SztsZPuE9qY78D7MvJCdg1AniDcUvzBtyDHqJWTkLheaqYgZu4rk4f6kurQbHN9ZVKN1wr2ZGkWwSfzEPDwQANzQus6TKVTkRtoExnGGr7tkd4vMG2RHKZk7LKzJPipwqBNCwxgYFnsFRpBKnEUvgEVSckwQP7HFUhxSLS4fHMk3RSSv1Ry3t4TZDYTwpipXvajEXtCrWrAV7yy4WnFpEBvHH9xRPinWqDSGGWnWZRBGBCZuR9kZaXcBbSS3vpEAq6T7XMx4cCKeK3Fii9ZB3b2TcekUBJsksu47BEZaHqJBJqApNnHkuSPEZCSmLSRkSXGfAdKZU3UTh8fGfRnVsyrdMqgfgEJ1EsEohR1rWqrv3a3NvXBtGa2YJm2nGENo9fgfvEHAWsXSMBjM7YgJ8Zza6ip6hC6bd6smsCCmTHm5M7c52n8r5aLUpk359qs11tfhsRcRv4nVfwWm6i9pyubNDtKchCNQV5X7pXntg5ovHmwAsoKnc9CmiGr64n65u5aa8XMB4w4r8zHesJ72kPjb1iwQ4PnVGMcYAZbgPqo9iEvpGnbGkh4yGwnSwoLk6qaQ4QsK7eQ5nuxFTbbAQJ7Zpnauqg5xXAJadt4G8TxwyQn5ZKBtGH5MrhtT9CJL9xe87xb7kSj63iReds2oSU9VC9VF52viHvCkTVXA9SDRSVXowPLNThBTEoRPY3hAkQU2eMfX1A9SxVUnnM",
		"databases",
		"", "", "")

	if err != nil {
		println("", err.Error())
	}

	time.Sleep(10 * time.Second)

	client.DisConnect(connectId)

	select {}
}
