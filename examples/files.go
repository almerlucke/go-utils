package examples

import (
	"log"
	"os"

	"github.com/almerlucke/go-utils/files"
)

// TestFilesScan test files scan
func TestFilesScan() {
	lines, err := files.ScanFile("examples/files.go")
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	for line := range lines {
		if line.Error != nil {
			log.Fatalf("err: %v", line.Error)
		}

		log.Printf("line: %v\n", line.Line)
	}
}

// TestReadEnv test read .env file
func TestReadEnv() {
	_, err := files.ReadDotEnvFile("examples/resources/.env", true)
	if err != nil {
		log.Fatalf("err: %v", err)
	}

	log.Printf("hallo %v", os.Getenv("HALLO"))
	log.Printf("url %v", os.Getenv("APP_URL"))
}
