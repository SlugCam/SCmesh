package dsr

import (
	"reflect"
	"testing"
)

func TestRouteCache(t *testing.T) {
	cache := newRouteCache()
	cache.addRoute([]NodeID{NodeID(1), NodeID(2), NodeID(5)}, 1)
	cache.addRoute([]NodeID{NodeID(1), NodeID(5), NodeID(3)}, 0)
	cache.addRoute([]NodeID{NodeID(8), NodeID(6), NodeID(5)}, 3)

	cases := []struct {
		dest NodeID
		want []NodeID
	}{
		{NodeID(1), []NodeID{NodeID(1)}},
		{NodeID(9), nil},
		{NodeID(5), []NodeID{NodeID(1), NodeID(5)}},
		{NodeID(3), []NodeID{NodeID(1), NodeID(5), NodeID(3)}},
	}

	for _, c := range cases {
		got := cache.getRoute(c.dest)
		if !reflect.DeepEqual(got, c.want) {
			t.Errorf("getRoute(%v) == %v, want %v", c.dest, got, c.want)
		}
	}
}

func TestFindNodeIndex(t *testing.T) {
	cases := []struct {
		r    []NodeID
		d    NodeID
		want int
	}{
		{
			r:    []NodeID{NodeID(1), NodeID(5), NodeID(2)},
			d:    NodeID(2),
			want: 2,
		},
		{
			r:    []NodeID{NodeID(1), NodeID(5), NodeID(2)},
			d:    NodeID(1),
			want: 0,
		},
		{
			r:    []NodeID{},
			d:    NodeID(8),
			want: -1,
		},
		{
			r:    []NodeID{NodeID(1), NodeID(7), NodeID(2), NodeID(90)},
			d:    NodeID(4),
			want: -1,
		},
	}

	for _, c := range cases {
		got := findNodeIndex(c.r, c.d)
		if got != c.want {
			t.Errorf("findNodeIndex(%v, %v) == %v, want %v", c.r, c.d, got, c.want)
		}
	}

}
