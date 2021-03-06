package main

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"github.com/quite/i3gopher/history"
	"github.com/quite/i3gopher/util"
	"github.com/spf13/pflag"
	"go.i3wm.org/i3/v4"
)

func main() {
	_, sway := os.LookupEnv("SWAYSOCK")

	log.SetPrefix("i3gopher ")
	flagExec := pflag.String("exec", "", "cmd to exec on any window event (example: killall -USR1 i3status")
	flagLast := pflag.BoolP("focus-last", "l", false, "focus last focused container on current workspace")
	pflag.Parse()

	socketPath := getSocketPath(sway)

	if *flagLast {
		client, err := rpc.DialHTTP("unix", socketPath)
		if err != nil {
			log.Fatalf("dialing: %s", err)
		}
		err = client.Call("History.FocusLast", struct{}{}, nil)
		if err != nil {
			log.Fatalf("History.FocusLast error: %s", err)
		}
		os.Exit(0)
	}

	quit := make(chan bool, 1)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sig
		quit <- true
	}()

	if sway {
		i3.SocketPathHook = func() (string, error) {
			out, err := exec.Command("sway", "--get-socketpath").CombinedOutput()
			if err != nil {
				return "", fmt.Errorf("getting sway socketpath: %v (output: %s)", err, out)
			}
			return string(out), nil
		}
		i3.IsRunningHook = func() bool {
			out, err := exec.Command("swaymsg", "-t", "get_version").CombinedOutput()
			if err != nil {
				log.Printf("getting sway version: %v (output: %s)", err, out)
				return false
			}
			return true
		}
	}

	var hist = history.NewHistory()

	recv := i3.Subscribe(i3.WorkspaceEventType, i3.WindowEventType)
	go func() {
		current, err := util.GetFocusedCon()
		// just register currently focused container if there is one
		if err == nil {
			ws, err := util.GetWorkspaceByCon(current)
			if err != nil {
				log.Printf("init: error getting workspace of focused container: %s", err)
			}
			hist.Add(ws, current)
		}

		for recv.Next() {
			switch ev := recv.Event().(type) {
			case *i3.WindowEvent:
				if ev.Change == "focus" {
					current := ev.Container.ID
					ws, err := util.GetWorkspaceByCon(current)
					if err != nil {
						log.Printf("warn: error getting workspace of focused container: %s", err)
					}
					hist.Add(ws, current)
				}

				if *flagExec != "" {
					s := strings.Split(*flagExec, " ")
					_ = exec.Command(s[0], s[1:]...).Run() //nolint:gosec // run as user what user says
				}
			}
		}
		quit <- true
	}()

	go func() {
		if err := os.RemoveAll(socketPath); err != nil {
			log.Fatal(err)
		}
		if err := rpc.Register(hist); err != nil {
			log.Fatal(err)
		}
		rpc.HandleHTTP()
		l, e := net.Listen("unix", socketPath)
		if e != nil {
			log.Fatalf("listen error: %s", e)
		}
		err := http.Serve(l, nil)
		if err != nil {
			log.Fatal(err)
		}
	}()

	<-quit

	if err := recv.Close(); err != nil {
		log.Fatal(err)
	}
}

func getSocketPath(sway bool) string {
	dir := os.Getenv("XDG_RUNTIME_DIR")
	if dir == "" {
		dir = "/tmp"
	}
	var disp string
	if sway {
		disp = os.Getenv("WAYLAND_DISPLAY")
	} else {
		disp = os.Getenv("DISPLAY")
	}
	if disp == "" {
		log.Fatalf("Environment variable DISPLAY/WAYLAND_DISPLAY missing")
	}
	return fmt.Sprintf("%s/i3gopher-%d-%s", dir, os.Getuid(), disp)
}
