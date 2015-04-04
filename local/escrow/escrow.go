/*
The package escrow includes the reliability layer for the SCmesh system.
*/
package escrow

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
	//PATH_PREFIX          = "/var/SlugCam/SCmesh"
	DIST_COUNTER_PATH = "count"
	DIST_STORE_PATH   = "store"
	DIST_OUT_PATH     = "out"
	DIST_META_PATH    = "meta"
)

const (
	COL_FILE_SCAN_TIMEOUT = 5 * time.Second
