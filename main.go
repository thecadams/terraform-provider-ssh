package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/stefansundin/terraform-provider-ssh/provider"
	"github.com/stefansundin/terraform-provider-ssh/ssh"
)

func main() {
	log.Printf("[DEBUG] pid=%d in main()", os.Getpid())
	go logSignals()

	if _, ok := os.LookupEnv(plugin.Handshake.MagicCookieKey); ok {
		plugin.Serve(&plugin.ServeOpts{
			ProviderFunc: func() *schema.Provider {
				return provider.SSHProvider()
			},
		})
	} else {
		var addr string
		var ppid int
		var proto string
		var err error

		log.SetFlags(0)

		flag.IntVar(&ppid, "ppid", 0, "parent process pid")
		flag.StringVar(&addr, "addr", os.Getenv("TF_SSH_PROVIDER_TUNNEL_ADDR"), "set rpc server address")
		flag.StringVar(&proto, "proto", os.Getenv("TF_SSH_PROVIDER_TUNNEL_PROTO"), "set rpc server protocol")
		flag.Parse()
		if ppid == 0 {
			if ppid, err = strconv.Atoi(os.Getenv("TF_SSH_PROVIDER_TUNNEL_PPID")); err != nil {
				log.Fatalf("[ERROR] parent process pid wasn't set")
			}
		}
		if addr == "" {
			log.Fatalf("[ERROR] RPC server address wasn't set")
		}
		var sshTunnel ssh.SSHTunnel
		if err := sshTunnel.Run(proto, addr, ppid); err != nil {
			log.Fatalf("[ERROR] failed to start SSH Tunnel:\n%s", err)
		}
	}
}

func logSignals() {
	signals := make(chan os.Signal, 1)
	for {
		signal.Notify(signals, syscall.SIGHUP, syscall.SIGINT, syscall.SIGKILL, syscall.SIGTERM, syscall.SIGQUIT)

		s := <-signals
		log.Printf("pid=%d received signal %s, passing it through to default handler", os.Getpid(), s)

		signal.Reset()

		currentProcess, err := os.FindProcess(os.Getpid()); if err != nil {
			log.Fatalf("[ERROR] failed to find current process: %s", err)
		}
		currentProcess.Signal(s)
	}
}
