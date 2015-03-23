package dsr

// TODO handle timeouts

// This file implements a Route Request Table as described in section 4.3 of
// RFC4728.

import "container/list"

type sentEntry struct {
	//TTL   uint32    // TTL for last route request send for this target
	time  time.Time // Time of last request
	count int       // Number of consecutive route discoveries since last valid reply, if 0 this means reply has been received
}

type receivedEntry struct {
}

// A requestTable contains a collection of cached routes. These routes all
// originate at the local node, but do not include the local node. So the first
// node listed in a route is the node to visit after the local node.
type requestTable struct {
	sentRequests     map[NodeID]requestTableEntry
	receivedRequests map[NodeID]list.List // Map initiator to list of requests received
}

// newRouteCache initialized an empty requestTable.
func newRequestTable() *requestTable {
	c := new(requestTable)
	return c
}

func (c *requestTable) sentRequest(target []NodeID) {
	// TODO what about 0 values
	v, ok := c.sentRequests[target]
	if ok {
		// update
		v.count = v.count + 1
		v.time = time.Now()
	} else {
		// new request issued
		c.sentRequests[target] = sentEntry{
			time:  time.Now(),
			count: 1,
		}
	}
}

// receivedReply updates the request table to bring the count of requests sent
// for a target to 0. Should be called whenever a reply is found. If no entry
// exists in the request table nothing happens.
func (c *requestTable) receivedReply(target NodeID) {
	// TODO what about 0 values
	v, ok := c.sentRequests[target]
	if ok {
		v.count = 0
	}
}

// hasReceivedReply returns true if a reply has been received since our last
// request was sent. Used to see if we should resend a route request when a
// timeout occurs.
func (c *requestTable) discoveryInProcess(target NodeID) bool {
	v, ok := c.sentRequestss
	if ok {
		return v.count > 0
	} else {
		return false
	}
}
