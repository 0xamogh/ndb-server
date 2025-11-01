// internal/nbd/constants.go
package nbd

// Handshake & reply magics (big-endian on wire)
const (
	NBDMAGIC                   uint64 = 0x4e42444d41474943 // "NBDMAGIC"
	IHAVEOPT                   uint64 = 0x49484156454f5054 // "IHAVEOPT"
	NBD_REP_MAGIC              uint64 = 0x3e889045565a9
	NBD_SIMPLE_REPLY_MAGIC     uint32 = 0x67446698
	NBD_REQUEST_MAGIC          uint32 = 0x25609513
	NBD_STRUCTURED_REPLY_MAGIC uint32 = 0x668e33ef
)

// Handshake flags (server -> client)
const (
	NBD_FLAG_FIXED_NEWSTYLE = 1 << 0
	NBD_FLAG_NO_ZEROES      = 1 << 1
)

// Client flags (client -> server)
const (
	NBD_FLAG_C_FIXED_NEWSTYLE = 1 << 0
	NBD_FLAG_C_NO_ZEROES      = 1 << 1
)

// Transmission flags (server -> client)
const (
	NBD_FLAG_HAS_FLAGS         = 1 << 0
	NBD_FLAG_READ_ONLY         = 1 << 1
	NBD_FLAG_SEND_FLUSH        = 1 << 2
	NBD_FLAG_SEND_FUA          = 1 << 3
	NBD_FLAG_ROTATIONAL        = 1 << 4
	NBD_FLAG_SEND_TRIM         = 1 << 5
	NBD_FLAG_SEND_WRITE_ZEROES = 1 << 6
	NBD_FLAG_SEND_DF           = 1 << 7
	NBD_FLAG_CAN_MULTI_CONN    = 1 << 8
)

// Options (client -> server during handshake)
const (
	NBD_OPT_EXPORT_NAME       = 1
	NBD_OPT_ABORT             = 2
	NBD_OPT_LIST              = 3
	NBD_OPT_STARTTLS          = 5
	NBD_OPT_INFO              = 6
	NBD_OPT_GO                = 7
	NBD_OPT_STRUCTURED_REPLY  = 8
	NBD_OPT_LIST_META_CONTEXT = 9
	NBD_OPT_SET_META_CONTEXT  = 10
)

// Option replies (server -> client)
const (
	NBD_REP_ACK                 = 1
	NBD_REP_SERVER              = 2
	NBD_REP_INFO                = 3
	NBD_REP_META_CONTEXT        = 4
	NBD_REP_ERR_UNSUP           = (1 << 31) + 1
	NBD_REP_ERR_POLICY          = (1 << 31) + 2
	NBD_REP_ERR_INVALID         = (1 << 31) + 3
	NBD_REP_ERR_PLATFORM        = (1 << 31) + 4
	NBD_REP_ERR_TLS_REQD        = (1 << 31) + 5
	NBD_REP_ERR_UNKNOWN         = (1 << 31) + 6
	NBD_REP_ERR_SHUTDOWN        = (1 << 31) + 7
	NBD_REP_ERR_BLOCK_SIZE_REQD = (1 << 31) + 8
)

// Transmission request types
const (
	NBD_CMD_READ         = 0
	NBD_CMD_WRITE        = 1
	NBD_CMD_DISC         = 2
	NBD_CMD_FLUSH        = 3
	NBD_CMD_TRIM         = 4
	NBD_CMD_CACHE        = 5
	NBD_CMD_WRITE_ZEROES = 6
	NBD_CMD_BLOCK_STATUS = 7
)

// Error codes (simple reply `error` field)
const (
	NBD_EPERM     = 1
	NBD_EIO       = 5
	NBD_ENOMEM    = 12
	NBD_EINVAL    = 22
	NBD_ENOSPC    = 28
	NBD_EOVERFLOW = 75
	NBD_ENOTSUP   = 95
	NBD_ESHUTDOWN = 108
)
