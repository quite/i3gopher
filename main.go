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
	"regexp"
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
	flagExclude := pflag.String("exclude", "", "ignore container from history if its instance name matches this regexp")
	flagLast := pflag.BoolP("focus-last", "l", false, "focus last focused container on current workspace")
	pflag.Parse()

	var excludeRE *regexp.Regexp
	if *flagExclude != "" {
		excludeRE = regexp.MustCompile(*flagExclude)
	}

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
				return "", fmt.Errorf("getting sway socketpath: %w (output: %s)", err, out)
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

	hist := history.NewHistory()

	recv := i3.Subscribe(i3.WindowEventType)
	go func() {
		current, err := util.GetFocusedCon()
		// just register currently focused container if there is one
		if err == nil {
			add(hist, excludeRE, current)
		}

		for recv.Next() {
			if ev, ok := recv.Event().(*i3.WindowEvent); ok {
				if ev.Change == "focus" {
					add(hist, excludeRE, &ev.Container)
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
		//nolint:gosec // not worrying about
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

func add(hist *history.History, excludeRE *regexp.Regexp, con *i3.Node) {
	ws, err := util.GetWorkspaceByCon(con.ID)
	if err != nil {
		log.Printf("init: error getting workspace of focused container: %s", err)
	}
	if excludeRE != nil {
		for _, s := range []string{con.WindowProperties.Instance, con.AppID} {
			if len(s) != 0 && excludeRE.MatchString(s) {
				return
			}
		}
	}
	_ = hist.Add(ws, &con.ID)
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
