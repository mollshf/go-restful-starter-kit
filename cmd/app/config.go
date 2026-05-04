package main

import (
	"github.com/gin-contrib/cors"
)

var corsConfig = cors.Config{
	AllowOrigins: []string{
		"http://localhost:3000",
	},
	AllowMethods: []string{
		"GET",
		"POST",
		"PUT",
		"DELETE",
		"OPTIONS",
	},
}
