package dsr

// This file contains code implementing a route cache for DSR as described in
// section 4.1 of RFC4728. For simplicity we are implementing a route cache
// first, and later this could be swapped for a link cache if desired.

type cacheEntry struct {
	route []id
	cost  int
}

type routeCache container.List

// addRoute adds a route to the cache.
func (r *routeCache) addRoute(newRoute Route, cost int) {

}

// getRoute looks into the route cache and returns a route.
func (r *routeCache) getRoute(dest NodeID) {

}
