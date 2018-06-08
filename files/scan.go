package files

import (
	"bufio"
	"os"
)

// ScanLine holds a scanned line and a possible error
type ScanLine struct {
	Line  string
	Count int
	Error error
}

// ScanFile scans all lines in the give file
func ScanFile(filePath string) (chan ScanLine, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}

	lineChannel := make(chan ScanLine)

	go func() {
		defer file.Close()

		scanner := bufio.NewScanner(file)
		lineCnt := 0

		for scanner.Scan() {
			lineChannel <- ScanLine{
				Line:  scanner.Text(),
				Count: lineCnt,
			}

			lineCnt++
		}

		if err := scanner.Err(); err != nil {
			lineChannel <- ScanLine{
				Error: err,
				Count: lineCnt,
			}
		}

		close(lineChannel)
	}()

	return lineChannel, nil
}
