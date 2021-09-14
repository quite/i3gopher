package history

import (
	"fmt"
	"reflect"
	"testing"

	"go.i3wm.org/i3/v4"
)

func TestDrop(t *testing.T) {
	tests := []struct {
		in    []i3.NodeID
		depth int
		out   []i3.NodeID
	}{
		{[]i3.NodeID{}, 0, []i3.NodeID{}},
		{[]i3.NodeID{1}, 0, []i3.NodeID{}},
		{[]i3.NodeID{1}, 1, []i3.NodeID{1}},
		{[]i3.NodeID{1, 2}, 0, []i3.NodeID{1}},
		{[]i3.NodeID{1, 2}, 1, []i3.NodeID{2}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v, %d", tt.in, tt.depth), func(t *testing.T) {
			got := drop(tt.in, tt.depth)
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("got %v, wanted %v", got, tt.out)
			}
		})
	}
}

func TestCompact(t *testing.T) {
	tests := []struct {
		in  []i3.NodeID
		out []i3.NodeID
	}{
		{[]i3.NodeID{}, []i3.NodeID{}},
		{[]i3.NodeID{1}, []i3.NodeID{1}},
		{[]i3.NodeID{1, 1}, []i3.NodeID{1}},
		{[]i3.NodeID{1, 2, 2}, []i3.NodeID{1, 2}},
		{[]i3.NodeID{1, 1, 1, 1, 2, 2, 2}, []i3.NodeID{1, 2}},
		{[]i3.NodeID{1, 1, 2, 2, 3, 1}, []i3.NodeID{1, 2, 3, 1}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
			got := compact(tt.in)
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("got %v, wanted %v", got, tt.out)
			}
		})
	}
}

func TestDropPair(t *testing.T) {
	tests := []struct {
		in  []i3.NodeID
		out []i3.NodeID
	}{
		{[]i3.NodeID{}, []i3.NodeID{}},
		{[]i3.NodeID{1, 2, 1, 2}, []i3.NodeID{1, 2}},
		{[]i3.NodeID{3, 1, 2, 1, 2}, []i3.NodeID{3, 1, 2}},
		{[]i3.NodeID{1, 2, 1, 3}, []i3.NodeID{1, 2, 1, 3}},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.in), func(t *testing.T) {
			got := dropPair(tt.in)
			if !reflect.DeepEqual(got, tt.out) {
				t.Errorf("got %v, wanted %v", got, tt.out)
			}
		})
	}
}
