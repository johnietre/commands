package main
// Doesn't follow symbolic links

/* IDEAS
* Use Peek method of bufio reader to check first bytes of file
	(what isExecutable function does) in the searchFile function
* Add "replace" feature
* Add option to mute errors
*/

import (
	"bufio"
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"regexp"
  "sort"
	"strings"
)

var (
	what      string
	rexpr     *regexp.Regexp
	foundChan = make(chan string, 5)
	flags     = map[rune]bool{
		'c': false, // Search file contents
		'r': false, // Search directories recursively
		'i': false, // Case-insensitive search
    's': false, // Sort results
	}
	ignoredFiles = map[string]bool{
		".apng":  true,
		".avif":  true,
		".bmp":   true,
		".cur":   true,
		".dat":   true,
		".docx":  true,
		".gif":   true,
		".ico":   true,
		".jfif":  true,
		".jpeg":  true,
		".jpg":   true,
		".pjpeg": true,
		".pjp":   true,
		".png":   true,
		".pdf":   true,
		".svg":   true,
		".tif":   true,
		".tiff":  true,
		".webp":  true,
		".xlsx":  true,
	}
)

func main() {
  log.SetFlags(0)
	// The place to search (file or directory)
	where := "./"
	l := len(os.Args)
	if l < 2 {
		printHelp(1)
	}
	// The "what" must always be the first arg
	what = os.Args[1]
	// Get the flags and where, if provided
	if l > 2 {
		for i, arg := range os.Args[2:] {
			if arg[0] == '-' {
				// Cannot speciy multiple flags at once
				if len(arg) != 2 {
					log.Fatalf("Invalid flag: %s\n", arg)
				}
				switch arg[1] {
				case 'c', 'r', 'i', 's':
					flags[rune(arg[1])] = true
				case 'h':
					printHelp(0)
				default:
					log.Fatalf("Invalid flag: %s\n", arg)
				}
			} else {
				// There where must always be the second argument
				if i != 0 {
					log.Fatalln(`Path must be specified second (after the "what")`)
				}
				where = arg
			}
		}
	}
	where = path.Clean(where)
	// The statement to match against
	regexStmt := what
	if flags['i'] {
		regexStmt = "(?i)" + regexStmt
	}
	var err error
	if rexpr, err = regexp.Compile(regexStmt); err != nil {
		log.Fatalln(err)
	}

	// Check if the initial "where" is a file or directory
	fstat, err := os.Stat(where)
	if err != nil {
		log.Fatalln(err)
	}
	switch mode := fstat.Mode(); {
	case !mode.IsDir():
		f, err := os.Open(where)
		if err != nil {
			log.Fatalln(err)
		}
		defer f.Close()
		linenos, err := searchFile(f)
		if err != nil {
			log.Println(err)
		}
		if len(linenos) != 0 {
			fmt.Printf("%s:%s\n", where, strings.Join(linenos, ","))
		}
		return
	}
	// Iterate through the directory and listen for matches on "foundChan"
	dc := make(chan int, 1)
	go searchDir(where, dc, true)
  var results []string
	for e := range foundChan {
    if flags['s'] {
      results = append(results, e)
    } else {
		  fmt.Println(e)
    }
	}
	<-dc
  if flags['s'] {
    sort.Strings(results)
    for _, res := range results {
      fmt.Println(res)
    }
  }
}

func searchDir(dirPath string, dc chan<- int, initial bool) {
	// Once the function returns, we need to notify the calling routing (dc chan)
	// and close the foundChan if this is the initial "searchDir" call
	defer func() {
		if initial {
			close(foundChan)
		}
		dc <- 1
	}()
	entries, err := ioutil.ReadDir(dirPath)
  for err != nil {
    // If there are too many files open, try until success or different err
    if strings.Contains(err.Error(), "too many open files") {
      entries, err = ioutil.ReadDir(dirPath)
      continue
    }
    log.Println(err)
    return
  }
	// Keep track of the number of directories searched
	nDirs, doneChan := 0, make(chan int, 1)
	for _, de := range entries {
    if de.Mode() & os.ModeSymlink != 0 {
      continue
    }
		fullPath := path.Join(dirPath, de.Name())
		if flags['c'] && !de.IsDir() {
			// If searching file contents and directory entry isn't a directory
			// Continue loop if file is in ignored list
			if ignoredFiles[path.Ext(de.Name())] {
				continue
			}
			// Open file and make sure it's not an executable
TooMany:
			f, err := os.Open(fullPath)
			if err != nil {
        // If there are too many files open, try until success or different err
        if strings.Contains(err.Error(), "too many open files") {
          goto TooMany
        }
				log.Println(err)
				continue
			} else if isExec, err := isExecutable(f); err != nil {
				log.Println(err)
				continue
			} else if isExec {
        f.Close()
				continue
			}
			// Search file, may return lines even with an error
			linenos, err := searchFile(f)
			if err != nil {
				log.Println(err)
			}
			f.Close()
			if len(linenos) != 0 {
				foundChan <- fmt.Sprintf("%s:%s", fullPath, strings.Join(linenos, ","))
			}
			continue
		} else if !flags['c'] {
			// If directory entry names
			if rexpr.MatchString(de.Name()) {
				if de.IsDir() {
					fullPath += "/"
				}
				foundChan <- fullPath
			}
		}
		if flags['r'] && de.IsDir() {
			// Start another search if searching recursively
			nDirs++
			go searchDir(fullPath, doneChan, false)
		}
	}
	// Wait for "searchDir" calls to finish
	for i := 0; i < nDirs; i++ {
		<-doneChan
	}
}

// Only returns empty slice if no mathces were found before returning (even
// if an error was encountered)
func searchFile(f *os.File) (linenos []string, err error) {
	reader := bufio.NewReader(f)
	// Loop forever (until reader encounders error)
	for lineno := 1; true; lineno++ {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err.Error() != "EOF" {
				return linenos, err
			}
			return linenos, nil
		}
		if rexpr.MatchString(line) {
			linenos = append(linenos, fmt.Sprintf("%d", lineno))
		}
	}
	return
}

// Executable byte order mark
var executableBOM = [7]byte{127, 67, 76, 70, 2, 1, 1}

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
	return bytes.Contains(buf[:], executableBOM[:]), nil
}

func printHelp(code int) {
	stmt := "Usage: search {what} {where (optional)} [flags]\n"
	stmt += "    -c contents\t\tSearch contents of file(s)\n"
	stmt += "    -r recursive\t\tRecursively search directories\n"
	stmt += "    -i case-insensitive\t\tCase-insensitive search\n"
  stmt += "    -s sort results\t\tSort the results\n"
	log.Print(stmt)
	os.Exit(code)
}
