package util

import (
	"fmt"
	"log"

	"go.i3wm.org/i3/v4"
)

func GetFocusedCon() (i3.NodeID, error) {
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

func GetFocusedWS() (i3.NodeID, error) {
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
		return 0, fmt.Errorf("could not find container: %d", con)
	}
	if ws == 0 {
		return 0, fmt.Errorf("could not get workspace of container: %d", con)
	}
	return ws, nil
}
