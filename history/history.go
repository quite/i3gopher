package history

import (
	"fmt"
	"sync"

	"github.com/quite/i3gopher/util"
	"go.i3wm.org/i3/v4"
)

type History struct {
	wsNodes map[i3.NodeID][]i3.NodeID
	mu      sync.Mutex
}

func NewHistory() *History {
	return &History{
		wsNodes: make(map[i3.NodeID][]i3.NodeID),
		mu:      sync.Mutex{},
	}
}

func (h *History) FocusLast(_, _ *struct{}) error {
	var err error
	h.mu.Lock()
	defer h.mu.Unlock()

	focusedWS, err := util.GetFocusedWS()
	if err != nil {
		return fmt.Errorf("getFocusedWS: %w", err)
	}

	nodes := h.wsNodes[focusedWS]
	for {
		last := peek(nodes, 1)
		if last == 0 {
			break
		}
		if ws, _ := util.GetWorkspaceByCon(last); ws != focusedWS {
			// container is gone, has moved ws, or so
			nodes = drop(nodes, 1)
			continue
		}
		if _, err = i3.RunCommand(fmt.Sprintf("[con_id=%d] focus", last)); err != nil {
			err = fmt.Errorf("i3.RunCommand: %w", err)
			break
		}
		break
	}
	h.wsNodes[focusedWS] = nodes
	return err
}

func (h *History) Add(ws i3.NodeID, e *i3.NodeID) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	nodes := h.wsNodes[ws]
	if peek(nodes, 0) == *e {
		return nil
	}
	nodes = push(nodes, *e)
	h.wsNodes[ws] = nodes
	return nil
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

// Replace consecutive runs of same nodeID with singles.
func compact(s []i3.NodeID) []i3.NodeID {
	o := make([]i3.NodeID, 0)
	for i, e := range s {
		if i == 0 || e != s[i-1] {
			o = append(o, e)
		}
	}
	return o
}

// Remove top pair if equal to pair below.
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
