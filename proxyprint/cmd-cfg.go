package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func makeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:                   "proxyprint",
		Short:                 "Run a proxy which can print out communications in a variety of ways",
		Long:                  "Run a proxy which can print out communications in a variety of ways. A password can be specified using the PROXYPRINT_PWD environment variable (unless it is set otherwise).",
		Run:                   run,
		DisableFlagsInUseLine: true,
	}
	cmd.AddCommand(&cobra.Command{
		Use:   "generate",
		Short: "Generate a blank config file",
		Long:  "Generate a blank config file to populate. NOTE: the file is generated with mostly invalid values which must be changed or deleted.",
		Run:   runCfg,
	})

	flags := cmd.Flags()

	flags.StringVar(&config.Listen, "listen", "", "Network address to listen on")
	flags.StringVar(&config.Connect, "connect", "", "Network address to connect to")
	flags.StringVar(
		&config.Tunnel,
		"tunnel",
		"",
		"Network address of proxyprint session to tunnel to",
	)
	flags.StringVar(
		&config.ListenServers,
		"listen-servers",
		"",
		"Network address to listen for tunneling servers on",
	)
	flags.Var(
		&config.ClientPrint,
		"client-print",
		"Set the client data print (0 = off*, 1 = as string, "+
			"2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
	)
	flags.Var(
		&config.ServerPrint,
		"server-print",
		"Set the server data print (0 = off*, 1 = as string, "+
			"2 = as bytes, 3 = as lower hex bytestring, 4 = as upper hex bytestring)",
	)
	flags.StringVar(
		&config.ClientPrintFile,
		"client-print-file",
		"",
		"The file to write client print output to",
	)
	flags.StringVar(
		&config.ServerPrintFile,
		"server-print-file",
		"",
		"The file to write server print output to",
	)
	flags.Uint64Var(
		&config.Buffer,
		"buffer",
		1<<15,
		"Size of the buffer to use to copy data",
	)
	flags.UintVar(
		&config.MaxWaitingTunnels,
		"max-waiting-tunnels",
		10,
		"Set the number of tunnels (to servers) that can be waiting for servers at once."+
			"Used with --tunnel flag.",
	)
	flags.UintVar(
		&config.MaxAcceptedServers,
		"max-accepted-servers",
		10,
		"Set the number of tunneling servers that can be accepted/handled at once"+
			"Used with --listen-servers flag.",
	)
	flags.StringVar(
		&config.PwdEnvName,
		"pwd-env-name",
		"PROXYPRINT_PWD",
		"The environment variable for reading the tunneling password. "+
			"If the name of the variable starts with the string 'file:', the value "+
			"of the variable is treated as a file path and the pointed-to file is "+
			"read and its content used as the password. "+
			"An empty string means to not read any password "+
			"(will still use empty password). "+
			"An empty environment variable value means no (an empty) password.",
	)
	flags.BoolVar(
		&config.RequirePwdEnvExists,
		"require-pwd-env-exists",
		false,
		"Require environment variable value of pwd-env-name flag exists and, if "+
			"not, throw a fatal error. If false, only warn.",
	)
	flags.StringVar(
		&config.Log,
		"log", "", "File to output logs to (blank is command line)",
	)
	flags.StringVar(
		&config.MonitorServer,
		"monitor-server",
		"",
		"Network address to run HTTP monitor server on "+
			"(blank means not to run it)",
	)
	flags.String("cfg", "", "Path to config file")
	return cmd
}

// TODO: do better (try viper?)
type Config struct {
	Listen              string      `json:"listen,omitempty"`
	Connect             string      `json:"connect,omitempty"`
	Tunnel              string      `json:"tunnel,omitempty"`
	ListenServers       string      `json:"listenServers,omitempty"`
	ClientPrint         printStatus `json:"clientPrint,omitempty"`
	ServerPrint         printStatus `json:"serverPrint,omitempty"`
	ClientPrintFile     string      `json:"clientPrintFile,omitempty"`
	ServerPrintFile     string      `json:"serverPrintFile,omitempty"`
	Buffer              uint64      `json:"buffer,omitempty"`
	MaxWaitingTunnels   uint        `json:"maxOpenTunnels,omitempty"`
	MaxAcceptedServers  uint        `json:"maxAcceptedServers,omitempty"`
	PwdEnvName          string      `json:"pwdEnvName,omitempty"`
	RequirePwdEnvExists bool        `json:"requirePwdEnvExists,omitempty"`
	Log                 string      `json:"log,omitempty"`
	MonitorServer       string      `json:"monitorServer,omitempty"`
}
type ConfigPtrs struct {
	Listen              *string      `json:"listen,omitempty"`
	Connect             *string      `json:"connect,omitempty"`
	Tunnel              *string      `json:"tunnel,omitempty"`
	ListenServers       *string      `json:"listenServers,omitempty"`
	ClientPrint         *printStatus `json:"clientPrint,omitempty"`
	ServerPrint         *printStatus `json:"serverPrint,omitempty"`
	ClientPrintFile     *string      `json:"clientPrintFile,omitempty"`
	ServerPrintFile     *string      `json:"serverPrintFile,omitempty"`
	Buffer              *uint64      `json:"buffer,omitempty"`
	MaxWaitingTunnels   *uint        `json:"maxOpenTunnels,omitempty"`
	MaxAcceptedServers  *uint        `json:"maxAcceptedServers,omitempty"`
	PwdEnvName          *string      `json:"pwdEnvName,omitempty"`
	RequirePwdEnvExists *bool        `json:"requirePwdEnvExists,omitempty"`
	Log                 *string      `json:"log,omitempty"`
	MonitorServer       *string      `json:"monitorServer,omitempty"`
}

func (c *Config) FillEmptyFrom(other *Config) {
	if c.Listen == "" {
		c.Listen = other.Listen
	}
	if c.Connect == "" {
		c.Connect = other.Connect
	}
	if c.Tunnel == "" {
		c.Tunnel = other.Tunnel
	}
	if c.ListenServers == "" {
		c.ListenServers = other.ListenServers
	}
	// NOTE: do something else?
	if c.ClientPrint == noPrint {
		c.ClientPrint = other.ClientPrint
	}
	if c.ServerPrint == noPrint {
		c.ServerPrint = other.ServerPrint
	}
	if c.ClientPrintFile == "" {
		c.ClientPrintFile = other.ClientPrintFile
	}
	if c.ServerPrintFile == "" {
		c.ServerPrintFile = other.ServerPrintFile
	}
	if c.Buffer == 0 {
		c.Buffer = other.Buffer
	}
	if c.MaxWaitingTunnels == 0 {
		c.MaxWaitingTunnels = other.MaxWaitingTunnels
	}
	if c.MaxAcceptedServers == 0 {
		c.MaxAcceptedServers = other.MaxAcceptedServers
	}
	// NOTE: do something else?
	if c.PwdEnvName == "" {
		c.PwdEnvName = other.PwdEnvName
	}
	if c.RequirePwdEnvExists == false {
		c.RequirePwdEnvExists = other.RequirePwdEnvExists
	}
	if c.Log == "" {
		c.Log = other.Log
	}
	if c.MonitorServer == "" {
		c.MonitorServer = other.MonitorServer
	}
}

func checkFlagSet(flags *pflag.FlagSet, name string) bool {
	flag := flags.Lookup(name)
	if flag == nil {
		log.Fatal("invalid flag lookup: ", name)
	}
	return flag.Changed
}

func (c *Config) PopulateCheckFlags(other *ConfigPtrs, flags *pflag.FlagSet) {
	if other.Listen != nil && !checkFlagSet(flags, "listen") {
		c.Listen = *other.Listen
	}
	if other.Connect != nil && !checkFlagSet(flags, "connect") {
		c.Connect = *other.Connect
	}
	if other.Tunnel != nil && !checkFlagSet(flags, "tunnel") {
		c.Tunnel = *other.Tunnel
	}
	if other.ListenServers != nil && !checkFlagSet(flags, "listen-servers") {
		c.ListenServers = *other.ListenServers
	}
	// NOTE: do something else?
	if other.ClientPrint != nil && !checkFlagSet(flags, "client-print") {
		c.ClientPrint = *other.ClientPrint
	}
	if other.ServerPrint != nil && !checkFlagSet(flags, "server-print") {
		c.ServerPrint = *other.ServerPrint
	}
	if other.ClientPrintFile != nil && !checkFlagSet(flags, "client-print-file") {
		c.ClientPrintFile = *other.ClientPrintFile
	}
	if other.ServerPrintFile != nil && !checkFlagSet(flags, "server-print-file") {
		c.ServerPrintFile = *other.ServerPrintFile
	}
	if other.Buffer != nil && !checkFlagSet(flags, "buffer") {
		c.Buffer = *other.Buffer
	}
	if other.MaxWaitingTunnels != nil && !checkFlagSet(flags, "max-waiting-tunnels") {
		c.MaxWaitingTunnels = *other.MaxWaitingTunnels
	}
	if other.MaxAcceptedServers != nil && !checkFlagSet(flags, "max-accepted-servers") {
		c.MaxAcceptedServers = *other.MaxAcceptedServers
	}
	// NOTE: do something else?
	if other.PwdEnvName != nil && !checkFlagSet(flags, "pwd-env-name") {
		c.PwdEnvName = *other.PwdEnvName
	}
	if other.RequirePwdEnvExists != nil && !checkFlagSet(flags, "require-pwd-env-exists") {
		c.RequirePwdEnvExists = *other.RequirePwdEnvExists
	}
	if other.Log != nil && !checkFlagSet(flags, "log") {
		c.Log = *other.Log
	}
	if other.MonitorServer != nil && !checkFlagSet(flags, "monitor-server") {
		c.MonitorServer = *other.MonitorServer
	}
}

func runCfg(_ *cobra.Command, args []string) {
	path := "proxyprint.conf.json"
	if len(args) > 0 {
		info, err := os.Stat(args[0])
		if err != nil && !os.IsNotExist(err) {
			log.Fatal("error checking output path: ", err)
		} else if err == nil && info.IsDir() {
			path = filepath.Join(args[0], path)
		} else {
			path = args[0]
		}
	}
	f, err := os.Create(path)
	if err != nil {
		log.Fatal("error creating config file: ", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	config := Config{
		Listen:             "IP:PORT",
		Connect:            "IP:PORT",
		Tunnel:             "IP:PORT",
		ListenServers:      "IP:PORT",
		ClientPrint:        -1,
		ServerPrint:        -1,
		Buffer:             1 << 15,
		MaxAcceptedServers: 10,
		PwdEnvName:         "PROXYPRINT_PWD",
		Log:                "PATH",
		MonitorServer:      "IP:PORT",
	}
	if err := enc.Encode(config); err != nil {
		log.Fatal("error writing config file: ", err)
	}
}
