// Copyright 2021 Wayback Archiver. All rights reserved.
// Use of this source code is governed by the GNU GPL v3
// license that can be found in the LICENSE file.

package http // import "github.com/wabarc/ipfs-pinner/http"

import (
	"net/http"
	"time"

	"github.com/ybbus/httpretry"
)

func NewClient(client *http.Client) *http.Client {
	if client == nil {
		client = &http.Client{}
	}
	return httpretry.NewCustomClient(
		client,
		// retry 5 times
		httpretry.WithMaxRetryCount(5),
		// retry on status == 429, if status >= 500, if err != nil, or if response was nil (status == 0)
		httpretry.WithRetryPolicy(func(statusCode int, err error) bool {
			return err != nil || statusCode == http.StatusTooManyRequests || statusCode >= http.StatusInternalServerError || statusCode == 0
		}),
		// every retry should wait one more 10 second
		httpretry.WithBackoffPolicy(func(attemptNum int) time.Duration {
			return time.Duration(attemptNum+1) * 10 * time.Second
		}),
	)
}
