package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"go.i3wm.org/i3"
)

func main() {
	quit := make(chan bool, 1)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		quit <- true
	}()

	recv := i3.Subscribe(i3.WorkspaceEventType, i3.WindowEventType)
	go func() {
		for recv.Next() {
			switch ev := recv.Event().(type) {
			case *i3.WindowEvent:
				if ev.Change == "focus" {
					log.Println("WIND", ev.Container.Window, ev.Container.Name)
				}
			case *i3.WorkspaceEvent:
				if ev.Change == "focus" {
					log.Println("WORK", ev.Current.ID)
				}
			}
		}
	}()

	<-quit

	if err := recv.Close(); err != nil {
		log.Fatal(err)
	}
}
