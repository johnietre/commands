// Multiple sort (e.g., by size, then name)
package main

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	sortpkg "sort"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/johnietre/utils/go"
	"github.com/spf13/cobra"
)

var (
	sizeDenom             = B
	sort                  = SortNone
	recursive             = false
	prec                  = -1
	numWorkers NumWorkers = 1

	matchFunc = func(string) bool { return true }
	size      uint64
	totalSize uint64
	jobChan   chan Job
	wg        sync.WaitGroup
	// Should be added to when sending to jobChan and `Done`ed only ni doJob
	jobsWg sync.WaitGroup
)

func main() {
	log.SetFlags(0)

	cmd := &cobra.Command{
		Use:   "finfo",
		Short: "A program to get a bit of file info (like size and last mod)",
		Run:   run,
	}
	flags := cmd.Flags()
	flags.VarP(
		&sizeDenom, "size", "s",
		"Size denomination (B/KB/KiB/MB/MiB/GB/GiB/Custom)",
	)
	flags.StringArray(
		"path", nil,
		"Explicitly specify path (useful for passing a path that looks like a flag; can be passed multiple times)",
	)
	flags.BoolVarP(
		&recursive, "recursive", "r", false,
		"If directory, get size of all subdirectories (recursive)",
	)
	flags.IntVarP(
		&prec, "precision", "p", -1,
		"Precision of output decimals (<0 for no limit)",
	)
	flags.BoolP(
		"total", "t", false,
		"Calculate the total size of all args",
	)
	flags.StringP(
		"regex", "x", "",
		"Regex expression to match paths against when calculate size",
	)
	flags.Bool(
		"excl", false,
		"Exclude regular expression matches",
	)
	flags.Bool(
		"files", false, "Match only files with regular expressions",
	)
	flags.Bool(
		"dirs", false, "Match only directories with regular expressions",
	)
	flags.VarP(&sort, "sort", "S", "Sort output")
	flags.Bool("rev", false, "Reverse the sort (no effect if unsorted)")
	flags.VarP(&numWorkers, "num-workers", "N", `Number of workers to spawn; "cpus" to spawn for each CPU core, negative to use original method`)
	flags.Bool("stdin", false, "Read from stdin")
	cmd.MarkFlagsMutuallyExclusive("files", "dirs")

	if err := cmd.Execute(); err != nil {
		log.Fatal(err)
	}
}

func run(cmd *cobra.Command, args []string) {
	flags := cmd.Flags()
	regexStr := utils.Must(flags.GetString("regex"))
	excludeMatch := utils.Must(flags.GetBool("excl"))
	filesOnly := utils.Must(flags.GetBool("files"))
	dirsOnly := utils.Must(flags.GetBool("dirs"))
	total := utils.Must(flags.GetBool("total"))
	reverse := utils.Must(flags.GetBool("rev"))
	stdin := utils.Must(flags.GetBool("stdin"))

	if regexStr != "" {
		regex, err := regexp.Compile(regexStr)
		if err != nil {
			log.Fatal("invalid regex: ", err)
		}
		matchFunc = regex.MatchString
	}
	if excludeMatch {
		f := matchFunc
		matchFunc = func(s string) bool { return !f(s) }
	}
	if filesOnly {
		f := matchFunc
		matchFunc = func(s string) bool {
			l := len(s)
			return l == 0 || s[l-1] != '/' || f(s)
		}
	}
	if dirsOnly {
		f := matchFunc
		matchFunc = func(s string) bool {
			l := len(s)
			return l == 0 || s[l-1] == '/' || f(s)
		}
	}

	paths := append(args, utils.Must(flags.GetStringArray("path"))...)

	if stdin {
		r := bufio.NewReader(os.Stdin)
		for {
			line, err := r.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				log.Fatal("error reading from stdin: ", err)
			}
			l := len(line)
			if l == 0 {
				continue
			}
			paths = append(paths, line[:l-1])
		}
	}

	if len(paths) == 0 {
		log.Fatal("must provide path")
	}

	var outputs []*Output

	if numWorkers >= 0 {
		chanLen := numWorkers
		if chanLen == 0 {
			chanLen = 100
		}
		jobChan = make(chan Job, chanLen)

		go runJobs()
		for i := NumWorkers(1); i < numWorkers; i++ {
			go runJobs()
		}

		var jobs []Job
		for _, path := range paths {
			job := Job{
				path:    path,
				size:    NewAU64(0),
				started: NewABool(false),
				output:  &Output{name: path},
			}
			jobsWg.Add(1)
			jobChan <- job
			jobs = append(jobs, job)
		}

		jobsWg.Wait()
		close(jobChan)

		for _, job := range jobs {
			job.output.size = job.size.Load()
			totalSize += job.output.size
			outputs = append(outputs, job.output)
		}
	} else {
		for i, path := range paths {
			if sort == SortNone && i != 0 {
				fmt.Println(strings.Repeat("=", 40))
			}
			info, err := os.Stat(path)
			if err != nil {
				//log.Fatalf("error getting info: %v", err)
				log.Printf("error getting info: %v", err)
				continue
			}
			if !info.IsDir() {
				size = uint64(info.Size())
			} else {
				wg.Add(1)
				walkDir(path)
				wg.Wait()
			}
			if sort == SortNone {
				printInfo(path, info, size)
			} else {
				outputs = append(outputs, &Output{name: path, size: size, info: info})
			}
			if total {
				totalSize += size
			}
			size = 0
		}
	}

	if sort != SortNone {
		sortpkg.Slice(outputs, func(i, j int) bool {
			var res bool
			switch sort {
			case SortName:
				res = outputs[i].name < outputs[j].name
			case SortSize:
				res = outputs[i].size < outputs[j].size
			case SortTime:
				res = outputs[i].info.ModTime().Before(outputs[j].info.ModTime())
			default:
				res = outputs[i].name < outputs[j].name
			}
			return res != reverse
		})
	}
	for i, out := range outputs {
		if i != 0 {
			fmt.Println(strings.Repeat("=", 40))
		}
		printInfo(out.name, out.info, out.size)
	}
	if total {
		fmt.Println(strings.Repeat("=", 40))
		fmt.Println("Total Size:", makeSizeStr(totalSize))
	}
}

type Output struct {
	name string
	size uint64
	info os.FileInfo
}

type Job struct {
	path    string
	size    *AU64
	started *ABool
	output  *Output
}

func runJobs() {
	for job := range jobChan {
		if numWorkers == 0 {
			go func(job Job) {
				doJob(job)
			}(job)
			continue
		}
		doJob(job)
	}
}

func doJob(job Job) {
	defer jobsWg.Done()
	if job.started.Swap(true) {
		doJobDir(job)
		return
	}
	info, err := os.Stat(job.path)
	if err != nil {
		//log.Fatalf("error getting info: %v", err)
		log.Printf("error getting info: %v", err)
		return
	}
	job.output.info = info
	if !info.IsDir() {
		size = uint64(info.Size())
	} else {
		doJobDir(job)
	}
}

func doJobDir(job Job) {
	ents, err := os.ReadDir(job.path)
	if err != nil {
		log.Printf("error opening %s: %v", job.path, err)
		return
	}
	for _, ent := range ents {
		info, err := ent.Info()
		if err != nil {
			log.Printf(
				"error getting info for %s: %v",
				filepath.Join(job.path, ent.Name()), err,
			)
			continue
		}
		if !info.IsDir() {
			if matchFunc(ent.Name()) {
				job.size.Add(uint64(info.Size()))
			}
			continue
		} else if recursive {
			name := ent.Name()
			if l := len(name); l == 0 || name[l-1] != '/' {
				name += "/"
			}
			if matchFunc(name) {
				jobsWg.Add(1)
				jobChan <- Job{
					path: filepath.Join(job.path, ent.Name()),
					size: job.size,
				}
			}
		}
	}
}

func walkDir(path string) {
	defer wg.Done()
	ents, err := os.ReadDir(path)
	if err != nil {
		log.Printf("error opening %s: %v", path, err)
		return
	}
	for _, ent := range ents {
		info, err := ent.Info()
		if err != nil {
			log.Printf("error getting info for %s: %v", filepath.Join(path, ent.Name()), err)
			continue
		}
		if !info.IsDir() {
			if matchFunc(ent.Name()) {
				atomic.AddUint64(&size, uint64(info.Size()))
			}
			continue
		} else if recursive {
			name := ent.Name()
			if l := len(name); l == 0 || name[l-1] != '/' {
				name += "/"
			}
			if matchFunc(name) {
				wg.Add(1)
				go walkDir(filepath.Join(path, ent.Name()))
			}
		}
	}
}

func printInfo(path string, info fs.FileInfo, size uint64) {
	sizeStr := makeSizeStr(size)
	fmt.Printf(
		"%s\nSize: %s\nLast Mod: %s\n",
		//info.Name(),
		path,
		sizeStr,
		info.ModTime().Format("15:04 Jan 02, 2006"),
	)
}

func makeSizeStr(size uint64) string {
	return makeSizeStrFor(size, sizeDenom)
}

func makeSizeStrFor(size uint64, denom SizeDenom) (sizeStr string) {
	switch denom {
	case B:
		sizeStr = commas(size) + " B"
	case KB:
		sizeStr = commas(size / KB)
		if rem := size % KB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(KB), 'f', prec, 64)[1:]
		}
		sizeStr += " KB"
	case KiB:
		sizeStr = commas(size / KiB)
		if rem := size % KiB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(KiB), 'f', prec, 64)[1:]
		}
		sizeStr += " KiB"
	case MB:
		sizeStr = commas(size / MB)
		if rem := size % MB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(MB), 'f', prec, 64)[1:]
		}
		sizeStr += " MB"
	case MiB:
		sizeStr = commas(size / MiB)
		if rem := size % MiB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(MiB), 'f', prec, 64)[1:]
		}
		sizeStr += " MiB"
	case GB:
		sizeStr = commas(size / GB)
		if rem := size % GB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(GB), 'f', prec, 64)[1:]
		}
		sizeStr += " GB"
	case GiB:
		sizeStr = commas(size / GiB)
		if rem := size % GiB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(GiB), 'f', prec, 64)[1:]
		}
		sizeStr += " GiB"
	case TB:
		sizeStr = commas(size / TB)
		if rem := size % TB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(TB), 'f', prec, 64)[1:]
		}
		sizeStr += " TB"
	case TiB:
		sizeStr = commas(size / TiB)
		if rem := size % TiB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(TiB), 'f', prec, 64)[1:]
		}
		sizeStr += " TiB"
	case PB:
		sizeStr = commas(size / PB)
		if rem := size % PB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(PB), 'f', prec, 64)[1:]
		}
		sizeStr += " PB"
	case PiB:
		sizeStr = commas(size / PiB)
		if rem := size % PiB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(PiB), 'f', prec, 64)[1:]
		}
		sizeStr += " PiB"
	case EB:
		sizeStr = commas(size / EB)
		if rem := size % EB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(EB), 'f', prec, 64)[1:]
		}
		sizeStr += " EB"
	case EiB:
		sizeStr = commas(size / EiB)
		if rem := size % EiB; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(EiB), 'f', prec, 64)[1:]
		}
		sizeStr += " EiB"
	case XB:
		ssize := SizeDenom(size)
		if ssize >= EB {
			return makeSizeStrFor(size, EB)
		} else if ssize >= PB {
			return makeSizeStrFor(size, PB)
		} else if ssize >= TB {
			return makeSizeStrFor(size, TB)
		} else if ssize >= GB {
			return makeSizeStrFor(size, GB)
		} else if ssize >= MB {
			return makeSizeStrFor(size, MB)
		} else if ssize >= KB {
			return makeSizeStrFor(size, KB)
		} else {
			return makeSizeStrFor(size, B)
		}
	case XiB:
		ssize := SizeDenom(size)
		if ssize >= EiB {
			return makeSizeStrFor(size, EiB)
		} else if ssize >= PiB {
			return makeSizeStrFor(size, PiB)
		} else if ssize >= TiB {
			return makeSizeStrFor(size, TiB)
		} else if ssize >= GiB {
			return makeSizeStrFor(size, GiB)
		} else if ssize >= MiB {
			return makeSizeStrFor(size, MiB)
		} else if ssize >= KiB {
			return makeSizeStrFor(size, KiB)
		} else {
			return makeSizeStrFor(size, B)
		}
	default:
		sd := uint64(sizeDenom)
		sizeStr = commas(size / sd)
		if rem := size % sd; rem != 0 {
			sizeStr += strconv.FormatFloat(float64(rem)/float64(sd), 'f', prec, 64)[1:]
		}
		sizeStr += fmt.Sprintf(" /%d B", sd)
	}
	return
}

func commas(u uint64) string {
	numStr := strconv.FormatUint(u, 10)
	str := ""
	// Track when to place comma with cc
	for i, cc := len(numStr)-1, -1; i >= 0; i-- {
		cc++
		if cc == 3 {
			str, cc = ","+str, 0
		}
		str = string(numStr[i]) + str
	}
	return str
}

type SizeDenom uint64

func (sd *SizeDenom) Set(s string) error {
	switch strings.ToLower(s) {
	case "b":
		*sd = B
	case "kb":
		*sd = KB
	case "kib":
		*sd = KiB
	case "mb":
		*sd = MB
	case "mib":
		*sd = MiB
	case "gb":
		*sd = GB
	case "gib":
		*sd = GiB
	case "tb":
		*sd = TB
	case "tib":
		*sd = TiB
	case "pb":
		*sd = PB
	case "pib":
		*sd = PiB
	case "eb":
		*sd = EB
	case "eib":
		*sd = EiB
	case "x", "xb":
		*sd = XB
	case "xib":
		*sd = XiB
	default:
		n, err := strconv.ParseUint(s, 0, 64)
		if err != nil {
			return err
		}
		*sd = SizeDenom(n)
	}
	return nil
}

func (sd SizeDenom) String() string {
	switch sd {
	case B:
		return "B"
	case KB:
		return "KB"
	case KiB:
		return "KiB"
	case MB:
		return "MB"
	case MiB:
		return "MiB"
	case GB:
		return "GB"
	case GiB:
		return "GiB"
	case TB:
		return "TB"
	case TiB:
		return "TiB"
	case PB:
		return "PB"
	case PiB:
		return "PiB"
	case EB:
		return "EB"
	case EiB:
		return "EiB"
	case XB:
		return "XB"
	case XiB:
		return "XiB"
	}
	return fmt.Sprintf("%d B", sd)
}

func (sd SizeDenom) Type() string {
	return "size"
}

const (
	B   SizeDenom = 1
	KB            = 1_000
	KiB           = 1 << 10
	MB            = 1_000_000
	MiB           = 1 << 20
	GB            = 1_000_000_000
	GiB           = 1 << 30
	TB            = 1_000_000_000_000
	TiB           = 1 << 40
	PB            = 1_000_000_000_000_000
	PiB           = 1 << 50
	EB            = 1_000_000_000_000_000_000
	EiB           = 1 << 60
	XB            = (1 << 64) - 1
	XiB           = (1 << 64) - 2
)

type Sort int

func (s *Sort) Set(str string) error {
	switch strings.ToLower(str) {
	case "", "none":
		*s = SortNone
	case "name":
		*s = SortName
	case "size":
		*s = SortSize
	case "time":
		*s = SortTime
	default:
		return fmt.Errorf("invalid sort")
	}
	return nil
}

func (s Sort) String() string {
	switch s {
	case SortNone:
		return "none"
	case SortName:
		return "name"
	case SortSize:
		return "size"
	case SortTime:
		return "time"
	}
	return "???"
}

func (s Sort) Type() string {
	return "sort"
}

const (
	SortNone Sort = iota
	SortName
	SortSize
	SortTime
)

type NumWorkers int

// NOTE: settings negative uses non-worker way
func (nw *NumWorkers) Set(s string) error {
	if s == "cpus" {
		*nw = NumWorkers(runtime.NumCPU())
		return nil
	}
	u, err := strconv.Atoi(s)
	if err == nil {
		*nw = NumWorkers(u)
	}
	return err
}

func (nw NumWorkers) String() string {
	return fmt.Sprintf("%d", nw)
}

func (NumWorkers) Type() string {
	return "num_cpu"
}

type AU64 struct {
	val uint64
}

func NewAU64(val uint64) *AU64 {
	return &AU64{val: val}
}

func (au *AU64) Add(a uint64) uint64 {
	return atomic.AddUint64(&au.val, a)
}

func (au *AU64) Load() uint64 {
	return atomic.LoadUint64(&au.val)
}

type ABool struct {
	val uint32
}

func NewABool(b bool) *ABool {
	val := uint32(0)
	if b {
		val = 1
	}
	return &ABool{val: val}
}

func (au *ABool) Store(b bool) {
	if b {
		atomic.StoreUint32(&au.val, 1)
	} else {
		atomic.StoreUint32(&au.val, 0)
	}
}

func (au *ABool) Swap(b bool) bool {
	if b {
		return atomic.SwapUint32(&au.val, 1) != 0
	} else {
		return atomic.SwapUint32(&au.val, 0) != 0
	}
}

func (au *ABool) Load() bool {
	return atomic.LoadUint32(&au.val) != 0
}
