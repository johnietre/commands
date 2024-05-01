package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/spf13/cobra"
)

func main() {
	log.SetFlags(0)

	rootCmd := &cobra.Command{
		Use:   "srvr [address to run (default: 127.0.0.1:8000)] [--dir,-d dir]",
		Short: "Serve a directory on the given address",
		Run: func(cmd *cobra.Command, args []string) {
			dir := cmd.Flag("dir").Value.String()
			addr := cmd.Flags().Arg(0)
			if addr == "" {
				addr = "127.0.0.1:8000"
			}

			fmt.Printf("Serving %s on %s\n", dir, addr)
			err := http.ListenAndServe(addr, http.FileServer(http.Dir(dir)))
			if err != nil {
				log.Fatal(err)
			}
		},
	}
	flags := rootCmd.Flags()
	flags.StringP("dir", "d", ".", "Directory to serve")
	if err := rootCmd.Execute(); err != nil {
		log.Fatal(err)
	}
}
