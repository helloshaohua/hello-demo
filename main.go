package main

import (
	"hello-demo/api"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	markdownHandler := api.NewMarkdown("./static/markdown")
	r := gin.Default()
	r.GET("/:filename", markdownHandler.GetMarkdown)
	log.Fatal(r.Run(":8859"))
}
