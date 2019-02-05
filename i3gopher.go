package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"go.i3wm.org/i3"
)

const markPrefix = "_i3gopher-last-on-"

func getFocusedCon() i3.NodeID {
	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}
	con := tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Focused && n.Type == i3.Con
	})
	if con == nil {
		log.Fatalf("could not find a focused container")
	}
	return con.ID
}

func getWorkspaceByCon(con i3.NodeID) i3.NodeID {
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
		log.Fatalf("could not find container")
	}
	if ws == 0 {
		log.Fatalf("could not get workspace")
	}
	return ws
}

func main() {
	flagExec := flag.String("exec", "", "cmd to exec on any window event (example: killall -USR1 i3status")
	flagLast := flag.Bool("focus-last", false, "focus last container on current workspace")
	flag.Parse()

	if *flagLast {
		ws := getWorkspaceByCon(getFocusedCon())
		i3.RunCommand(fmt.Sprintf("[con_mark=%s%d] focus", markPrefix, ws))
		os.Exit(0)
	}

	quit := make(chan bool, 1)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		quit <- true
	}()

	recv := i3.Subscribe(i3.WorkspaceEventType, i3.WindowEventType)
	go func() {
		var focusedcon = make(map[i3.NodeID]i3.NodeID)
		current := getFocusedCon()
		focusedcon[getWorkspaceByCon(current)] = current

		for recv.Next() {
			switch ev := recv.Event().(type) {
			case *i3.WindowEvent:
				if ev.Change == "focus" {
					current := ev.Container.ID
					ws := getWorkspaceByCon(current)
					if last, ok := focusedcon[ws]; ok {
						if last != current {
							cmd := fmt.Sprintf("[con_id=%d] mark --add %s%d",
								last, markPrefix, ws)
							i3.RunCommand(cmd)
						}
					}
					focusedcon[ws] = current
				}

				if *flagExec != "" {
					s := strings.Split(*flagExec, " ")
					cmd := exec.Command(s[0], s[1:]...)
					_ = cmd.Run()
				}
			}
		}
		quit <- true
	}()

	<-quit

	if err := recv.Close(); err != nil {
		log.Fatal(err)
	}
}
