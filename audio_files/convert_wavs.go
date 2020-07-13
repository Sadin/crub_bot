package main

import (
	"io"
	"os"

	"github.com/jonas747/dca"
	//"io/ioutil"
)

func main() {
	// Encoding a file and saving it to disk
	encodeSession, err := dca.EncodeFile("./for_the_day_FULL.mp3", dca.StdEncodeOptions)
	// Make sure everything is cleaned up, that for example the encoding process if any issues happened isnt lingering around
	defer encodeSession.Cleanup()

	output, err := os.Create("./for_the_day_FULL.dca")
	if err != nil {
		// Handle the error
	}

	io.Copy(output, encodeSession)
}
