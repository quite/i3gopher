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
	"sync"
	"syscall"

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

	var history = newHistory()

	recv := i3.Subscribe(i3.WorkspaceEventType, i3.WindowEventType)
	go func() {
		current, err := getFocusedCon()
		// just register currently focused container if there is one
		if err == nil {
			ws, err := getWorkspaceByCon(current)
			if err != nil {
				log.Printf("init: error getting workspace of focused container: %s", err)
			}
			history.add(ws, current)
		}

		for recv.Next() {
			switch ev := recv.Event().(type) {
			case *i3.WindowEvent:
				if ev.Change == "focus" {
					current := ev.Container.ID
					ws, err := getWorkspaceByCon(current)
					if err != nil {
						log.Printf("warn: error getting workspace of focused container: %s", err)
					}
					history.add(ws, current)
				}

				if *flagExec != "" {
					s := strings.Split(*flagExec, " ")
					_ = exec.Command(s[0], s[1:]...).Run() //nolint:gosec
				}
			}
		}
		quit <- true
	}()

	go func() {
		if err := os.RemoveAll(socketPath); err != nil {
			log.Fatal(err)
		}
		if err := rpc.Register(history); err != nil {
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

func getFocusedCon() (i3.NodeID, error) {
	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}
	con := tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Focused && n.Type == i3.Con
	})
	if con == nil {
		return 0, fmt.Errorf("could not find a focused container")
	}
	return con.ID, nil
}

func getFocusedWS() (i3.NodeID, error) {
	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}
	ws := tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Type == i3.WorkspaceNode
	})
	if ws == nil {
		return 0, fmt.Errorf("could not find a focused workspace")
	}
	return ws.ID, nil
}

func getWorkspaceByCon(con i3.NodeID) (i3.NodeID, error) {
	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}
	var ws i3.NodeID
	foundcon := tree.Root.FindChild(func(n *i3.Node) bool {
		// pick up workspace along the way
		if n.Type == i3.WorkspaceNode {
			ws = n.ID
		}
		return n.ID == con
	})
	if foundcon == nil {
		return 0, fmt.Errorf("could not find container: %d", con)
	}
	if ws == 0 {
		return 0, fmt.Errorf("could not get workspace of container: %d", con)
	}
	return ws, nil
}

type History struct {
	wsNodes map[i3.NodeID][]i3.NodeID
	mu      sync.Mutex
}

func newHistory() *History {
	return &History{
		wsNodes: make(map[i3.NodeID][]i3.NodeID),
		mu:      sync.Mutex{},
	}
}

func (h *History) FocusLast(_, _ *struct{}) error {
	var err error
	h.mu.Lock()
	defer h.mu.Unlock()

	focusedWS, err := getFocusedWS()
	if err != nil {
		return fmt.Errorf("getFocusedWS: %s", err)
	}

	nodes := h.wsNodes[focusedWS]
	for {
		last := peek(nodes, 1)
		if last == 0 {
			break
		}
		if ws, _ := getWorkspaceByCon(last); ws != focusedWS {
			// container is gone, has moved ws, or so
			nodes = drop(nodes, 1)
			continue
		}
		if _, err = i3.RunCommand(fmt.Sprintf("[con_id=%d] focus", last)); err != nil {
			err = fmt.Errorf("i3.RunCommand: %s", err)
			break
		}
		break
	}
	h.wsNodes[focusedWS] = nodes
	return err
}

func (h *History) add(ws i3.NodeID, e i3.NodeID) {
	h.mu.Lock()
	defer h.mu.Unlock()

	nodes := h.wsNodes[ws]
	if peek(nodes, 0) == e {
		return
	}
	nodes = push(nodes, e)
	h.wsNodes[ws] = nodes
}

func push(s []i3.NodeID, e i3.NodeID) []i3.NodeID {
	return prune(append(s, e))
}

func drop(s []i3.NodeID, depth int) []i3.NodeID {
	if s == nil || depth < 0 {
		return s
	}
	i := len(s) - 1 - depth
	if i < 0 {
		return s
	}
	return prune(append(s[:i], s[i+1:]...))
}

func peek(s []i3.NodeID, depth int) i3.NodeID {
	if s == nil || depth < 0 {
		return 0
	}
	i := len(s) - 1 - depth
	if i < 0 {
		return 0
	}
	return s[i]
}

func prune(s []i3.NodeID) []i3.NodeID {
	return dropPair(compact(s))
}

// Replace consecutive runs of same nodeID with singles
func compact(s []i3.NodeID) []i3.NodeID {
	new := make([]i3.NodeID, 0)
	for i, e := range s {
		if i == 0 || e != s[i-1] {
			new = append(new, e)
		}
	}
	return new
}

// Remove top pair if equal to pair below
func dropPair(s []i3.NodeID) []i3.NodeID {
	if s == nil || len(s) < 4 {
		return s
	}
	if peek(s, 0) == peek(s, 2) &&
		peek(s, 1) == peek(s, 3) {
		s = drop(s, 0)
		s = drop(s, 0)
	}
	return s
}
