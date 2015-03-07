package constants

import "time"

// Constants for the delimiter sequences for the parser. IMPORTANT, PACK_SEQ and
// COMM_SEQ need to be the same length and must both begin with the letter 'C'.
// If this requirement needs to change then prefilter.checkEscape() will have to
// be modified.
const (
	PACK_SEQ = "CBU\r\n"
	COMM_SEQ = "CMD\r\n"
	EXIT_SEQ = "EXIT\r\n"
)

// TODO magic number
// ESCAPE_SEQ_TIMEOUT is used by the prefilter. When a possible command sequence
// is encountered the scanner will wait for this amount of time before
// determining that the value is not a command sequence but part of a packet or
// it is junk depending on the mode.
const DELIM_SEQ_TIMEOUT = 50 * time.Millisecond
