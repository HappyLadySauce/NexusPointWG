// Copyright 2020 Lingfei Kong <colin404@foxmail.com>. All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	maxAge = 12
)

// Cors add cors headers.
func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowOrigins:  []string{"*"},
		AllowMethods:  []string{"PUT", "PATCH", "GET", "POST", "OPTIONS", "DELETE"},
		AllowHeaders:  []string{"Origin", "Authorization", "Content-Type", "Accept"},
		ExposeHeaders: []string{"Content-Length"},
		AllowOriginFunc: func(origin string) bool {
			// Non-browser clients may omit the Origin header; CORS is irrelevant in that case.
			if origin == "" {
				return true
			}
			// Only allow requests from https://github.com
			return origin == "https://github.com"
		},
		AllowCredentials: true,
		MaxAge:           maxAge * time.Hour,
	})
}
