package main

import (
	"bufio"
	"flag"
	"fmt"
  "io"
	logpkg "log"
	"os"
	"path"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var (
	logFile *os.File
)

type TimeValue struct {
	timePtr *int64
}

func (t TimeValue) String() string {
  if t.timePtr != nil {
	  return fmt.Sprintf("%d", *t.timePtr)
  }
  return ""
}

var dateRegex = regexp.MustCompile(`(\d{2})/(\d{2})/(\d{2}) (\d{2}):(\d{2})`)
var invalidFormatErr = fmt.Errorf("invalid date format")
var invalidDateErr = fmt.Errorf("invalid date value")

func (t TimeValue) Set(s string) error {
	matches := dateRegex.FindStringSubmatch(s)
	if len(matches) == 0 {
		return invalidFormatErr
	}
  var date time.Time
	if year, err := strconv.Atoi(matches[3]); err != nil {
		return invalidFormatErr
	} else if month, err := strconv.Atoi(matches[1]); err != nil {
		return invalidFormatErr
	} else if month > 12 || month < 1 {
		return invalidDateErr
	} else if day, err := strconv.Atoi(matches[2]); err != nil {
		return invalidFormatErr
	} else if !validDay(year, month, day) {
		return invalidDateErr
	} else if hour, err := strconv.Atoi(matches[4]); err != nil {
		return invalidFormatErr
	} else if hour > 23 {
		return invalidDateErr
	} else if minute, err := strconv.Atoi(matches[5]); err != nil {
		return invalidFormatErr
	} else if minute > 59 {
		return invalidDateErr
	} else {
	  date = time.Date(
		  2000+year, time.Month(month), day, hour, minute, 0, 0, time.Now().Location(),
	  )
  }
  if time.Now().Before(date) {
    return fmt.Errorf("cannot pass future time")
  }
  *t.timePtr = date.UTC().Unix()
	return nil
}

func main() {
	logpkg.SetFlags(0)
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		logpkg.Fatal("error getting daylogs")
	}

	var err error
	logFile, err = os.OpenFile(
		path.Join(path.Dir(thisFile), "daylog.log"),
		os.O_CREATE|os.O_APPEND|os.O_RDWR,
		0644,
	)
	if err != nil {
		logpkg.Fatal(err)
	}
	defer logFile.Close()

	var t, start, end int64
	flag.Var(&TimeValue{&t}, "t", "Time of log message")
	flag.Var(&TimeValue{&start}, "start", "Earliest log time to search for")
	flag.Var(&TimeValue{&end}, "end", "Latest log time to search for")
	msgPtr := flag.String("m", "", "Log message")
	flag.Parse()

  if t != 0 {
    logpkg.Fatal("-t not implemented")
  }

	if *msgPtr != "" {
		if err := log(t, *msgPtr); err != nil {
      logpkg.Fatal(err)
    }
		return
	} else if args := flag.Args(); len(args) != 0 {
		*msgPtr = strings.Join(args, " ")
		if err := log(t, *msgPtr); err != nil {
      logpkg.Fatal(err)
    }
		return
	}

	r, lineno := bufio.NewReader(logFile), 0
	for {
		line, err := r.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			logpkg.Fatal(err)
		}
		lineno++
		parts := strings.SplitN(strings.TrimSpace(line), "|", 2)
		if len(parts) != 2 {
			logpkg.Fatalf("invalid log format on line %d", lineno)
		}
		t, err := strconv.ParseInt(parts[0], 10, 64)
		if err != nil {
			logpkg.Fatalf("invalid time format on line %d", lineno)
		}
		if t >= start && (t <= end || end == 0) {
			fmt.Printf("%s\t%s\n", time.Unix(t, 0).Format(time.ANSIC), parts[1])
		} else if t > end && end != 0 {
			break
		}
	}
}

func log(timestamp int64, msg string) error {
	if timestamp == 0 {
    timestamp = time.Now().UTC().Unix()
    logMsg := fmt.Sprintf("%d|%s\n", timestamp, msg)
    _, err := logFile.WriteString(logMsg)
    if err == nil {
      err = logFile.Sync()
    }
    return err
  }
  return nil
}

func validDay(year, month, day int) bool {
	if day > 31 || day < 1 {
		return false
	}
	switch month {
	case 4, 6, 9, 11:
		return day <= 30
	case 2:
		if year%400 == 0 || (year%100 != 0 && year%4 == 0) {
			return day <= 29
		}
		return day <= 28
	}
	return true
}
