package dsr

// TODO we should be able to maintain cache size so it does not grow to large.

import "container/list"

// This file contains code implementing a path cache for DSR as described in
// section 4.1 of RFC4728. For simplicity we are implementing a path cache
// first, and later this could be swapped for a link cache if desired. In
// particular, the Link Max-Life route cache is recommended.

type cacheEntry struct {
	route []NodeID
	cost  int
}

// A routeCache contains a collection of cached routes. These routes all
// originate at the local node, but do not include the local node. So the first
// node listed in a route is the node to visit after the local node.
type routeCache struct {
	l *list.List
}

func newRouteCache() *routeCache {
	c := new(routeCache)
	c.l = list.New()
	return c
}

// addRoute adds a route to the cache.
func (c *routeCache) addRoute(route []NodeID, cost int) {
	c.l.PushBack(cacheEntry{
		route: route,
		cost:  cost,
	})
}

// getRoute looks into the route cache and returns shortest path. Returns nil if
// no route is found.
// TODO should return lowest cost path
func (c *routeCache) getRoute(dest NodeID) []NodeID {

	var shortestPath []NodeID

	for e := c.l.Front(); e != nil; e = e.Next() {
		curEntry := e.Value.(cacheEntry)
		curRoute := curEntry.route
		i := findNodeIndex(curRoute, dest)
		if i > -1 {
			newRoute := curRoute[:i+1]
			if shortestPath == nil || len(newRoute) < len(shortestPath) {
				shortestPath = newRoute
			}
		}
	}

	return shortestPath
}

// findNodeIndex finds d (destination) in r (route). If it is found, it returns
// the index of the destination, otherwise it returns -1.
func findNodeIndex(r []NodeID, d NodeID) int {
	for i, v := range r {
		if v == d {
			return i
		}
	}
	return -1
}
