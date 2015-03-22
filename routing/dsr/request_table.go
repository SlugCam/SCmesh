package dsr

import "container/list"

type requestTableEntry struct {
	//TTL   uint32    // TTL for last route request send for this target
	time  time.Time // Time of last request
	count int       // Number of consecutive route discoveries since last valid reply
}

// A routeCache contains a collection of cached routes. These routes all
// originate at the local node, but do not include the local node. So the first
// node listed in a route is the node to visit after the local node.
type routeCache struct {
	l *list.List
}

// newRouteCache initialized an empty routeCache.
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
// no route is found. The route is returned as specified by the DSR specs of
// what a route in a source route should look like, meaning the source and
// destination are not included. In other words this function will return an
// array of the intermediate nodes to reach node dest.
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
