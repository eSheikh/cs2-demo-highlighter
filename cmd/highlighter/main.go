package main

import (
	"context"
	"log"
	"os"

	"github.com/eSheikh/cs2-demo-highlighter/internal/bootstrap"
)

func main() {
	if err := bootstrap.Run(context.Background(), os.Args[1:], log.Default()); err != nil {
		log.Fatal(err)
	}
}
