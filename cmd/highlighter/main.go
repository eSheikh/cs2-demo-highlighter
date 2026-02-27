package main

import (
	"context"
	"log"
	"os"

	"cs2-demo-highlighter/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(context.Background(), os.Args[1:], log.Default()); err != nil {
		log.Fatal(err)
	}
}
