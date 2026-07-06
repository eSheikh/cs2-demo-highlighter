package main

import (
	"log"
	"os"

	"github.com/eSheikh/cs2-demo-highlighter/internal/engine"
	"github.com/eSheikh/cs2-demo-highlighter/internal/parser/demoinfocs"
	"github.com/eSheikh/cs2-demo-highlighter/internal/service"
	"github.com/eSheikh/cs2-demo-highlighter/internal/tui"
)

func main() {
	demoArg := ""
	if len(os.Args) > 1 {
		demoArg = os.Args[1]
	}

	eng := engine.New(demoinfocs.NewParser(), service.NewHighlightService())
	if err := tui.Run(eng, demoArg); err != nil {
		log.Fatal(err)
	}
}
