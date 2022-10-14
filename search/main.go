package main

// Doesn't follow symbolic links

/* IDEAS
 * Look at sync.Pool
 */

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
  "io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

var (
	what           string
	rexpr          *regexp.Regexp
	running        int32 = 0
	replaceWithPtr *string
	foundChan      = make(chan string, 5)
	flags          = make(map[byte]bool)
	ignoredFiles   = map[string]bool{
		".apng": true, ".avif": true, ".bmp": true, ".cur": true,
    ".dat": true, ".docx": true, ".exe": true, ".gif": true, ".ico": true,
		".jfif": true, ".jpeg": true, ".jpg": true, ".pjpeg": true,
		".pjp": true, ".png": true, ".pdf": true, ".svg": true,
		".tif": true, ".tiff": true, ".webp": true, ".xlsx": true,
	}
)

func main() {
	log.SetFlags(0)

	// Create the flagset
	flagSet := flag.NewFlagSet("search", flag.ExitOnError)
	flagSet.Bool("c", false, "Search file contents")
	flagSet.Bool("r", false, "Search directories recursively")
	flagSet.Bool("i", false, "Case-insensitive search")
	flagSet.Bool("s", false, "Sort results")
	flagSet.Bool("m", false, "Mute non-fatal errors")
	flagSet.Bool("n", false, "Count the number of occurrences")
	flagSet.Bool("x", false, "Use regex")
  flagSet.Bool("l", false, "Print line numbers of occurrences")
	replaceWithPtr = flagSet.String("p", "", "String to replace the \"what\" with")
	flagSet.Usage = func() {
		fmt.Fprintf(
			flagSet.Output(),
			"Usage of search what [where (optional)] [flags (optional)]\n",
		)
		flagSet.PrintDefaults()
	}

	// Make sure the minimum amount of arguments was passed
	l := len(os.Args)
	if l < 2 {
		flagSet.Usage()
		return
	}

	// The "what" must always be the first argument
	what = os.Args[1]
	var (
		wheres    []string // A list of places to search
		flagStart = 0      // Where the flags start in the argument list
	)
	// Get the list of "wheres"
	for i, arg := range os.Args[2:] {
		if arg[0] == '-' {
			flagStart = i + 2
			break
		}
		wheres = append(wheres, arg)
	}
	if len(wheres) == 0 {
		wheres = []string{"."}
	}

	// Parse the flags if there are any
	if flagStart != 0 {
		flagSet.Parse(os.Args[flagStart:])
		flagSet.Visit(func(f *flag.Flag) {
			flags[f.Name[0]] = true
		})
	}

	// The statement to match against
	regexStmt := what
	if !flags['x'] {
    regexStmt = regexp.QuoteMeta(regexStmt)
	}
	if flags['i'] {
		regexStmt = "(?i)" + regexStmt
	}
  var err error
  if rexpr, err = regexp.Compile(regexStmt); err != nil {
		log.Fatal(err)
	}

  if flags['m'] {
    log.SetOutput(io.Discard)
  }

	// Loop through the "wheres"
	for _, where := range wheres {
		// Check to make sure the "where" exists
		fstat, err := os.Stat(where)
		if err != nil {
			log.Printf("%v", err)
			continue
		}
		if mode := fstat.Mode(); mode.IsRegular() {
			// Run searchFile if the where is a regular file
			go searchFile(where, false)
		} else if mode.IsDir() {
			// Run searchDir if the where is a directory
			go searchDir(where)
		}
	}
	// Wait for at least one goroutine to spin up
	time.Sleep(time.Millisecond * 50)
	var results []string
	// Get the results from the channel and wait for num running to hit 0
	count := 0
Loop:
	for {
		select {
		case e := <-foundChan:
			if flags['s'] {
				if flags['n'] {
					count += strings.Count(e, ",") + 1
				}
				results = append(results, e)
			} else if !flags['n'] {
				fmt.Println(e)
			} else {
				count += strings.Count(e, ",") + 1
			}
		default:
			if atomic.LoadInt32(&running) == 0 {
				break Loop
			}
		}
	}
	// Print the count, if requested
	if flags['n'] {
		fmt.Println(count)
	}
	// If sorting, sort the results
	if flags['s'] {
		sort.Strings(results)
		for _, res := range results {
			fmt.Println(res)
		}
	}
}

func searchDir(dirPath string) {
	atomic.AddInt32(&running, 1)
	defer atomic.AddInt32(&running, -1)
	entries, err := ioutil.ReadDir(dirPath)
	for err != nil {
		// If there are too many files open, try until success or different err
		if strings.Contains(err.Error(), "too many open files") {
			entries, err = ioutil.ReadDir(dirPath)
			continue
		}
		log.Printf("%v", err)
		return
	}
	for _, de := range entries {
		if de.Mode()&os.ModeSymlink != 0 {
			continue
		}
		fullPath := path.Join(dirPath, de.Name())
		if flags['c'] && !de.IsDir() {
			// If searching file contents and directory entry isn't a directory
			// Continue loop if file is in ignored list
			if ignoredFiles[path.Ext(de.Name())] {
				continue
			}
			// Search the file
			searchFile(fullPath, true)
			continue
		} else if !flags['c'] && !flags['p'] {
			// If directory entry names
			if isMatch(de.Name()) {
				if de.IsDir() {
					fullPath += "/"
				}
				foundChan <- fullPath
			}
		}
		if flags['r'] && de.IsDir() {
			// Start another search if searching recursively
			go searchDir(fullPath)
		}
	}
}

func searchFile(filepath string, checkIfExec bool) {
	atomic.AddInt32(&running, 1)
	defer atomic.AddInt32(&running, -1)
	f, err := os.Open(filepath)
	for err != nil {
		if !strings.Contains(err.Error(), "too many open files") {
			log.Printf("%v", err)
			return
		}
		f, err = os.Open(filepath)
	}
	defer f.Close()
	if checkIfExec {
		// Check if the file is an executable
		if isExec, err := isExecutable(f); err != nil {
			log.Printf("%v", err)
			return
		} else if isExec {
			return
		}
	} else if flags['p'] {
		// This should only be called with explicit single files
		// checkIfExec should only be false when not being called from searchDir
		// i.e., when being called on a single file only
		writePath := filepath + ".search"
		if writeFile, err := os.Create(writePath); err != nil {
			log.Printf("%v", err)
		} else if err := replaceFileContents(f, writeFile); err != nil {
			log.Printf("%v", err)
		} else if err := os.Rename(writePath, filepath); err != nil {
			log.Printf("%v", err)
		}
		return
	}
	// Search file
	linenos, err := searchFileContents(f)
	if err != nil {
		log.Printf("%v", err)
	}
	if len(linenos) != 0 {
    if flags['l'] {
		  foundChan <- fmt.Sprintf("%s:%s", filepath, strings.Join(linenos, ","))
    } else {
		  foundChan <- fmt.Sprintf("%s", filepath)
    }
	}
}

// Only returns empty slice if no mathces were found before returning (even
// if an error was encountered)
func searchFileContents(f *os.File) (linenos []string, err error) {
	reader, line := bufio.NewReader(f), ""
	// Loop forever (until reader encounders error)
	for lineno := 1; true; lineno++ {
		line, err = reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				return
			}
      err = nil
			return
		}
		if isMatch(strings.TrimSpace(line)) {
			linenos = append(linenos, fmt.Sprintf("%d", lineno))
      if !flags['l'] {
        return
      }
		}
	}
	return
}

func replaceFileContents(readFile, writeFile *os.File) error {
	r, w := bufio.NewReader(readFile), bufio.NewWriter(writeFile)
	/* IDEA: Defer flush */
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				return err
			}
			return w.Flush()
		}
		if _, err := w.WriteString(replace(line)); err != nil {
			return err
		}
	}
}

func isMatch(text string) bool {
	return rexpr.MatchString(text)
}

func replace(text string) string {
	/* IDEA: Look at ReplaceAllString and ReplaceAllLiteralString */
	return rexpr.ReplaceAllString(text, *replaceWithPtr)
}

// Executable byte order mark sequeences
// Ex) Possible sequences: 127 69 76 70 2 1 1 0 or 127 69 76 70 2 1 1 3
var executableBOMSeqs = [][]byte{
	{127},
	{67, 69},
	{76},
	{70},
	{2},
	{1},
	{1},
	{0, 3},
}

// Checks where a file is an execuable based on the file's first seven bytes
func isExecutable(f *os.File) (bool, error) {
	var buf [8]byte
	_, err := f.Read(buf[:])
	if err != nil {
		if err.Error() == "EOF" {
			return false, nil
		}
		return false, err
	}
	// Return the file to it's original state
	if _, err := f.Seek(0, 0); err != nil {
		return false, err
	}
	for i, b := range buf {
		if !bytes.Contains(executableBOMSeqs[i], []byte{b}) {
			return false, nil
		}
	}
	return true, nil
}
