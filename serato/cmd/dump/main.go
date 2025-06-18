package main

import (
	"encoding/json"
	"log"
	"os"

	"github.com/autonomouskoi/trackstar/serato"
)

func fatalIfError(err error, msg string) {
	if err != nil {
		log.Fatal("error ", msg, ": ", err)
	}
}

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <.session file>", os.Args[0])
	}

	infh, err := os.Open(os.Args[1])
	fatalIfError(err, "opening session file")
	defer infh.Close()

	je := json.NewEncoder(os.Stdout)
	je.SetIndent("", "  ")

	fatalIfError(serato.ReadSession(infh, func(t serato.Track) {
		je.Encode(t)
	}), "reading session")
}
