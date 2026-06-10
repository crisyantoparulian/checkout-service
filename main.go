package main

import (
	"github.com/crisyantoparulian/checkout-service/cmd"
	_ "github.com/crisyantoparulian/checkout-service/docs"
)

// @title Go-Boilerplate
// @version 1.0.0
// @description This is boilerplate code for golang project

// @contact.name Crisyanto Parulian
// @contact.email crisyanto.p@gmail.com

// @host localhost:9000
func main() {
	cmd.Execute()
}
