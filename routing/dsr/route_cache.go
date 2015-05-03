package dsr

// TODO we should be able to maintain cache size so it does not grow to
// large.

// TODO don't add duplicates, instead update timeout

import (
	"container/list"

	"github.com/SlugCam/SCmesh/packet/header"
)

// This file contains code implementing a path cache for DSR as
// described in section 4.1 of RFC4728. For simplicity we are
// implementing a path cache first, and later this could be swapped for
// a link cache if desired. In particular, the Link Max-Life route cache
// is recommended.

type cachedNode struct {
	address uint32
	cost    uint32
}

type route struct {
	nodes []uint32
	cost  int
}

// A routeCache contains a collection of cached routes. These routes all
// originate at the local node, but do not include the local node. So
// the first node listed in a route is the node to visit after the local
// node.
type routeCache struct {
	l *list.List // holds []cachedNode
}

// newRouteCache initialized an empty routeCache.
func newRouteCache() *routeCache {
	c := new(routeCache)
	c.l = list.New()
	return c
}

// addRoute adds a route to the cache. Note that we do not track the
// last nodes cost this will be sent to zero.
// TODO this uses the existing pointer, may not be safe. Check this.
func (c *routeCache) addRoute(route []*header.DSRHeader_Node, target uint32) {
	cachedRoute := make([]cachedNode, len(route)+1)
	for i, n := range route {
		cachedRoute[i] = cachedNode{
			address: *n.Address,
			cost:    *n.Cost,
		}
	}
	cachedRoute[len(cachedRoute)-1] = cachedNode{
		address: target,
		cost:    uint32(0),
	}
	c.l.PushBack(cachedRoute)
}

func (c *routeCache) removeLink(a, b uint32) {
	for e := c.l.Front(); e != nil; e = e.Next() {
		curCached := e.Value.([]cachedNode)
		lastWasA := false
		for _, n := range curCached {
			if n.address == a {
				lastWasA = true
			} else {
				if n.address == b && lastWasA {
					c.l.Remove(e)
					break
				}
				lastWasA = false
			}
		}
	}
}

func (c *routeCache) removeNeighbor(neighbor uint32) {
	for e := c.l.Front(); e != nil; e = e.Next() {
		curCached := e.Value.([]cachedNode)
		if len(curCached) > 0 && curCached[0].address == neighbor {
			c.l.Remove(e)
		}
	}
}

// getRoute looks into the route cache and returns shortest path. Returns nil if
// no route is found. The route is returned as specified by the DSR specs of
// what a route in a source route should look like, meaning the source and
// destination are not included. In other words this function will return an
// array of the intermediate nodes to reach node dest.
// TODO should return lowest cost path
func (c *routeCache) getRoute(dest uint32) []uint32 {

	var routes []*route

	// Find all valid routes
	for e := c.l.Front(); e != nil; e = e.Next() {
		curCached := e.Value.([]cachedNode)
		curRoute := findInnerRoute(curCached, dest)
		if curRoute != nil {
			routes = append(routes, curRoute)
		}
	}

	// Choose a route
	if len(routes) > 0 {
		return routes[0].nodes
	} else {
		return nil
	}
}

// findNodeIndex finds d (destination) in r (route). If it is found, it returns
// a route (excluding destination), otherwise it returns nil.
func findInnerRoute(r []cachedNode, d uint32) *route {
	addresses := make([]uint32, 0)
	cost := 0
	for _, v := range r {
		if v.address == d {
			return &route{
				nodes: addresses,
				cost:  cost,
			}
		}
		addresses = append(addresses, v.address)
		cost += int(v.cost)
	}
	return nil
}
