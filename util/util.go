package util

import (
	"errors"
	"log"

	"go.i3wm.org/i3/v4"
)

var (
	ErrNoFocusedContainer   = errors.New("could not find a focused container")
	ErrNoFocusedWorkspace   = errors.New("could not find a focused workspace")
	ErrContainerNotFound    = errors.New("could not find container")
	ErrContainerNoWorkspace = errors.New("could not get workspace of container")
)

func GetFocusedCon() (*i3.Node, error) {
	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}
	con := tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Focused && n.Type == i3.Con
	})
	if con == nil {
		return nil, ErrNoFocusedContainer
	}
	return con, nil
}

func GetFocusedWS() (i3.NodeID, error) {
	tree, err := i3.GetTree()
	if err != nil {
		log.Fatal(err)
	}
	ws := tree.Root.FindFocused(func(n *i3.Node) bool {
		return n.Type == i3.WorkspaceNode
	})
	if ws == nil {
		return 0, ErrNoFocusedWorkspace
	}
	return ws.ID, nil
}

func GetWorkspaceByCon(con i3.NodeID) (i3.NodeID, error) {
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
		return 0, ErrContainerNotFound
	}
	if ws == 0 {
		return 0, ErrContainerNoWorkspace
	}
	return ws, nil
}
