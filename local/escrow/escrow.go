/*
The package escrow includes the reliability layer for the SCmesh system.
TODO will the filesystem handle enough files, maybe break into separate files
TODO make sure everything is clean
*/
package escrow

import "time"

// These path constants will be relative to the prefix provided to the
// collect function.
const (
	COL_STORE_PATH = "c.in"
	COL_OUT_PATH   = "c.out"
	COL_META_PATH  = "c.meta"
)

// These path constants will be relative to the prefix provided to the
// distribute function.
const (
	DIST_COUNTER_PATH = "d.count"
	DIST_STORE_PATH   = "d.store"
	DIST_OUT_PATH     = "d.out"
	DIST_META_PATH    = "d.meta"
)

const (
	COL_FILE_SCAN_TIMEOUT = 5 * time.Second
)

// For distribute
const (
	MAX_PAYLOAD_SIZE    = 512
	REQUEST_BUFFER_SIZE = 100
	TIMEOUT_BUFFER_SIZE = 10
)

// TODO priorities for distribute
const (
	FIRST_TIMEOUT = 5 * time.Second // Time to wa
	ACK_TIMEOUT   = 4 * time.Second
)
