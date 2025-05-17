package main

import (
	"ej_final/api"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	api.InitRoutes(r)
	r.Run() // 0.0.0.0:8080
}
