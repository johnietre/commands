package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/johnietre/commands/meyerson/cli"
	webs "golang.org/x/net/websocket"
)

var (
	addr, indexPath, jsPath, cssPath string
	procs           sync.Map

  globalEnv []string
  globalEnvMtx sync.RWMutex

	conns           sync.Map

	running  int32
	numConns int32
)

func init() {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		log.Fatal("error getting file")
	}
	indexPath = filepath.Join(filepath.Dir(file), "index.html")
  jsPath = filepath.Join(filepath.Dir(file), "index.js")
  cssPath = filepath.Join(filepath.Dir(file), "index.css")
}

func main() {
	log.SetFlags(0)

	// Run the CLI
	if len(os.Args) != 1 && os.Args[1] == "cli" {
		cli.Run(os.Args[2:])
		return
	}

	usage := flag.CommandLine.Usage
	flag.CommandLine.Usage = func() {
		usage()
		out := flag.CommandLine.Output()
		fmt.Fprint(out, "Commands:\n")
		fmt.Fprintln(out, "  cli\tRun CLI")
	}
	flag.StringVar(&addr, "addr", "127.0.0.1:3350", "Address to run server on")
  configPath := flag.String("config", "", "Config to load")
	flag.Parse()
  
  if *configPath != "" {
    config, err := loadConfig(*configPath)
    if err != nil {
      log.Fatal("error loading config: ", err)
    }
    populateFromConfig(config)
  }

	r := http.NewServeMux()
	r.HandleFunc("/", homeHandler)
	r.HandleFunc("/index.js", jsHandler)
	r.HandleFunc("/index.css", cssHandler)
	r.Handle("/ws", webs.Handler(wsHandler))

	srvr := http.Server{
		Addr:    addr,
		Handler: r,
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	go func() {
		<-c
		srvr.Close()
	}()

	atomic.StoreInt32(&running, 1)
	log.Printf("running on %s", addr)
	if err := srvr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Printf("error running server: %v", err)
	}
	atomic.StoreInt32(&running, 0)

	conns.Range(func(_, iConn any) bool {
		iConn.(*webs.Conn).Close()
		return true
	})
	for atomic.LoadInt32(&numConns) != 0 {
	}
	procs.Range(func(_, iProc any) bool {
		iProc.(*Process).Kill()
		return true
	})
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, indexPath)
}

func jsHandler(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, jsPath)
}

func cssHandler(w http.ResponseWriter, r *http.Request) {
  http.ServeFile(w, r, cssPath)
}

func wsHandler(ws *webs.Conn) {
	defer ws.Close()
	conns.Store(ws.Request().RemoteAddr, ws)
	defer conns.Delete(ws.Request().RemoteAddr)
	atomic.AddInt32(&numConns, 1)
	defer atomic.AddInt32(&numConns, -1)
	for atomic.LoadInt32(&running) == 1 {
		var msg Message
		if err := webs.JSON.Receive(ws, &msg); err != nil {
			if err.Error() != "EOF" && !strings.Contains(err.Error(), "closed") {
				log.Printf("error receiving ws message: %v", err)
			}
			return
		}
		switch msg.Action {
		case ActionAdd:
			if msg.Process.Name != "" && msg.Process.Path != "" {
				if !addProcess(msg.Process) {
					webs.JSON.Send(ws, Message{Action: ActionError, Contents: "process exists"})
				}
			} else {
				webs.JSON.Send(ws, Message{Action: ActionError, Contents: "must have name and path"})
			}
		case ActionStart:
			if msg.Process.Name != "" {
				if proc, ok := procs.Load(msg.Process.Name); ok {
					if err := proc.(*Process).StartWith(msg.Process); err != nil {
						webs.JSON.Send(ws, Message{Action: ActionError, Contents: err.Error()})
					}
				} else {
					if msg.Process.Path != "" {
						if addProcess(msg.Process) {
							if err := msg.Process.Start(); err != nil {
								webs.JSON.Send(ws, Message{Action: ActionError, Contents: err.Error()})
							}
						}
					} else {
						webs.JSON.Send(ws, Message{Action: ActionError, Contents: "must have name and path"})
					}
				}
			}
		case ActionKill:
			if msg.Process.Name != "" {
				if proc, ok := procs.Load(msg.Process.Name); ok {
					if err := proc.(*Process).Kill(); err != nil {
						webs.JSON.Send(ws, Message{Action: ActionError, Contents: err.Error()})
					}
				}
			}
		case ActionInterrupt:
			if msg.Process.Name != "" {
				if proc, ok := procs.Load(msg.Process.Name); ok {
					if err := proc.(*Process).Interrupt(); err != nil {
						webs.JSON.Send(ws, Message{Action: ActionError, Contents: err.Error()})
					}
				}
			}
		case ActionDel:
			delProcess(msg.Process)
		case ActionRefresh:
			processes := make([]*Process, 0, 1)
			procs.Range(func(_, p any) bool {
				processes = append(processes, p.(*Process))
				return true
			})
			webs.JSON.Send(
				ws,
				Message{Action: ActionRefresh, Contents: string(must(json.Marshal(processes)))},
			)
		case ActionError:
			// Shouldn't be received from client
			continue
		default:
			webs.JSON.Send(
				ws,
				&Message{
					Action:   ActionError,
					Contents: fmt.Sprintf("invalid action: %s", msg.Action),
				},
			)
			continue
		}
	}
}

type Message struct {
	Action   string   `json:"action"`
	Process  *Process `json:"process,omitempty"`
	Contents string   `json:"contents,omitempty"`
}

const (
	ActionAdd       string = "add"
	ActionStart     string = "start"
	ActionKill      string = "kill"
	ActionInterrupt string = "interrupt"
	ActionDel       string = "del"
	ActionRefresh   string = "refresh"
  ActionEnv string = "env"
	ActionError     string = "error"
)

const (
	ProcessNotStarted int32 = iota
	ProcessRunning
	ProcessStopping
  ProcessStopped
)

var (
	ErrAlreadyRunning = fmt.Errorf("process already running")
	ErrStopped        = fmt.Errorf("process stopped")
	// Use to store in atomic value when there's no error so nil error isn't stored
	ErrNoErr = fmt.Errorf("")
)

type Process struct {
	Name   string   `json:"name"`
	Path   string   `json:"path"`
	Args   []string `json:"args"`
	Env    []string `json:"env"`
	Dir    string   `json:"dir"`
	Status int32    `json:"status"`
	Error  string   `json:"error"`
	Stderr string   `json:"stderr"`

	stderr Buffer
	cmd    *exec.Cmd
	status int32
	err    atomic.Value
}

func ProcFromCLI(proc *cli.Process) *Process {
  return &Process{
    Name: proc.Name,
    Path: proc.Program,
    Args: proc.Args,
    Env: proc.Env,
    Dir: proc.Dir,
  }
}

func (p *Process) Start() error {
	if !atomic.CompareAndSwapInt32(&p.status, ProcessStopped, ProcessRunning) {
		return ErrAlreadyRunning
	}
	go p.start()
	return nil
}

func (p *Process) StartWith(proc *Process) error {
	if !atomic.CompareAndSwapInt32(&p.status, ProcessStopped, ProcessRunning) {
		return ErrAlreadyRunning
	}
	p.Path = proc.Path
	p.Args = proc.Args
	p.Env = proc.Env
	p.Dir = proc.Dir
	go p.start()
	return nil
}

func (p *Process) start() error {
	p.err.Store(ErrNoErr.Error())
	p.cmd = exec.Command(p.Path, p.Args...)
	if len(p.Env) != 0 {
		p.cmd.Env = p.Env
	}
	p.cmd.Dir = p.Dir
	p.cmd.Stderr = &p.stderr
	p.stderr.Reset()

	notify(p, ActionStart)
	err := error(p.cmd.Run())
	if err == nil {
		p.err.Store(ErrNoErr.Error())
	} else {
		p.err.Store(err.Error())
	}
	atomic.StoreInt32(&p.status, ProcessStopped)
	notify(p, ActionKill)
	return err
}

func (p *Process) Interrupt() error {
	if atomic.LoadInt32(&p.status) != ProcessRunning {
		return ErrStopped
	}
	return p.cmd.Process.Signal(os.Interrupt)
}

func (p *Process) Kill() error {
	if !atomic.CompareAndSwapInt32(&p.status, ProcessRunning, ProcessStopping) {
		return ErrStopped
	}
	return p.cmd.Process.Kill()
}

func (p *Process) MarshalJSON() ([]byte, error) {
	var errStr string
	if iErr := p.err.Load(); iErr != nil {
		errStr = iErr.(string)
	} else {
		errStr = ""
	}
	return []byte(fmt.Sprintf(
		`{`+
			`"name":"%s",`+
			`"path":"%s",`+
			`"args":%s,`+
			`"env":%s,`+
			`"dir":"%s",`+
			`"status":%d,`+
			`"error":%s,`+
			`"stderr":%s`+
			`}`,
		p.Name, p.Path,
		must(json.Marshal(p.Args)), must(json.Marshal(p.Env)),
		p.Dir,
		atomic.LoadInt32(&p.status),
		must(json.Marshal(errStr)),
		must(json.Marshal(p.stderr.String())),
	)), nil
}

func addProcess(p *Process) bool {
	if _, loaded := procs.LoadOrStore(p.Name, p); loaded {
		return false
	}
	notify(p, ActionAdd)
	return true
}

func delProcess(p *Process) bool {
	_, loaded := procs.LoadAndDelete(p.Name)
	if !loaded {
		return false
	}
	p.Kill()
	notify(p, ActionDel)
	return true
}

func notify(proc *Process, action string) {
	msg := &Message{Action: action, Process: proc}
	conns.Range(func(_, ws any) bool {
		err := webs.JSON.Send(ws.(*webs.Conn), msg)
		if err != nil {
			log.Println(err)
		}
		return true
	})
}

func loadConfig(path string) (*cli.Config, error) {
  f, err := os.Open(path)
  if err != nil {
    return nil, err
  }
  defer f.Close()
  config := &cli.Config{}
  if err := json.NewDecoder(f).Decode(config); err != nil {
    return nil, err
  }
  return config, nil
}

func parseConfig(bytes []byte) (*cli.Config, error) {
  config := &cli.Config{}
  if err := json.Unmarshal(bytes, config); err != nil {
    return nil, err
  }
  return config, nil
}

func populateFromConfig(config *cli.Config) {
  for _, cproc := range config.Procs {
    proc := ProcFromCLI(cproc)
    procs.Store(proc.Name, proc)
  }
  globalEnv = config.Env
}

type Buffer struct {
	buffer []byte
	mtx    sync.RWMutex
}

func (b *Buffer) Reset() {
	b.mtx.Lock()
	b.buffer = b.buffer[:0]
	b.mtx.Unlock()
}

func (b *Buffer) String() string {
	b.mtx.RLock()
	defer b.mtx.RUnlock()
	return string(b.buffer)
}

func (b *Buffer) Write(p []byte) (int, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.buffer = append(b.buffer, p...)
	return len(p), nil
}

func must[T any](t T, err error) T {
	if err != nil {
		panic(err)
	}
	return t
}
