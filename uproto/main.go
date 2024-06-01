package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	nameExpr, numExpr *regexp.Regexp
	wg                sync.WaitGroup
)

func main() {
	// Set log flags
	log.SetFlags(0)

	// Set command-line flags
	namePtr := flag.String("name", "", "name of the enum or struct")
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		log.Fatal("Usage: uproto [flags] filenames...")
	}

	// Create the regular expressions
	if *namePtr == "" {
		*namePtr = `\w+`
	}
	var err error
	nameExpr, err = regexp.Compile(fmt.Sprintf("(message|enum) %s {", *namePtr))
	if err != nil {
		log.Fatal("Error compiling regex: ", err)
	}
	numExpr = regexp.MustCompile(`(= \d+)`)

	// Get and open the files passed
	for _, filename := range args {
		if !strings.HasSuffix(filename, ".proto") {
			log.Printf("%s: file must have .proto extension", filename)
		}
		wg.Add(1)
		go func(filename string) {
			defer wg.Done()
			fix(filename)
		}(filename)
	}
	wg.Wait()
}

func fix(filename string) {
	// Get the file stat to open with same permissions
	s, err := os.Stat(filename)
	if err != nil {
		log.Printf("%s: %v", filename, err)
	}
	f, err := os.OpenFile(filename, os.O_RDWR, s.Mode())
	if err != nil {
		log.Printf("%s: %v", filename, err)
	}

	// Read the file and replace what's necessary
	reader := bufio.NewReader(f)
	var (
		newContents string
		// Holds the current numberings (slice to keep track of nested types)
		currents []int
		// A pointer to the current "current" in the "currents" slice
		current *int
		// Keeps track of which "current" we're on in the "currents" slice
		n = -1
		// Keeps track of how many levels we are in
		// If it's different from n, the contents won't be changed
		level = -1
	)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			// Close the file
			f.Close()
			if err.Error() == "EOF" {
				break
			}
			log.Printf("%s: %v", filename, err)
		}
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "//") {
			newContents += line
			continue
		} else if strings.HasSuffix(trimmed, "{") {
			level++
		}
		if nameExpr.MatchString(line) {
			currents = append(currents, 0)
			n++
			current = &currents[n]
			if !strings.HasPrefix(trimmed, "enum") {
				*current = 1
			}
		} else if match := numExpr.FindString(line); n != -1 && level == n && match != "" {
			line = strings.ReplaceAll(line, match, fmt.Sprintf("= %d", *current))
			(*current)++
		} else if strings.HasPrefix(trimmed, "}") {
			if n == level && n != -1 {
				currents = currents[:n]
				n--
				if n != -1 {
					current = &currents[n]
				}
			}
			if level != -1 {
				level--
			}
		}
		newContents += line
	}
	// Write the new contents to the file
	if err := os.WriteFile(filename, []byte(newContents), s.Mode()); err != nil {
		log.Printf("%s: %v", filename, err)
	}
}
