package libzookeeper

import (
	"errors"
	"log"
	"time"
)

var (
	ErrUnhandledFieldType = errors.New("zk: unhandled field type")
	ErrPtrExpected        = errors.New("zk: encode/decode expect a non-nil pointer to struct")
	ErrShortBuffer        = errors.New("zk: buffer too small")
)

type defaultLogger struct{}

func (defaultLogger) Printf(format string, a ...interface{}) {
	log.Printf(format, a...)
}

type ACL struct {
	Perms  int32
	Scheme string
	ID     string
}

type Stat struct {
	Czxid          int64 // The zxid of the change that caused this znode to be created.
	Mzxid          int64 // The zxid of the change that last modified this znode.
	Ctime          int64 // The time in milliseconds from epoch when this znode was created.
	Mtime          int64 // The time in milliseconds from epoch when this znode was last modified.
	Version        int32 // The number of changes to the data of this znode.
	Cversion       int32 // The number of changes to the children of this znode.
	Aversion       int32 // The number of changes to the ACL of this znode.
	EphemeralOwner int64 // The session id of the owner of this znode if the znode is an ephemeral node. If it is not an ephemeral node, it will be zero.
	DataLength     int32 // The length of the data field of this znode.
	NumChildren    int32 // The number of children of this znode.
	Pzxid          int64 // last modified children
}

// ServerClient is the information for a single Zookeeper client and its session.
// This is used to parse/extract the output fo the `cons` command.
type ServerClient struct {
	Queued        int64
	Received      int64
	Sent          int64
	SessionID     int64
	Lcxid         int64
	Lzxid         int64
	Timeout       int32
	LastLatency   int32
	MinLatency    int32
	AvgLatency    int32
	MaxLatency    int32
	Established   time.Time
	LastResponse  time.Time
	Addr          string
	LastOperation string // maybe?
	Error         error
}

// ServerClients is a struct for the FLWCons() function. It's used to provide
// the list of Clients.
//
// This is needed because FLWCons() takes multiple servers.
type ServerClients struct {
	Clients []*ServerClient
	Error   error
}

// ServerStats is the information pulled from the Zookeeper `stat` command.
type ServerStats struct {
	Sent        int64
	Received    int64
	NodeCount   int64
	MinLatency  int64
	AvgLatency  int64
	MaxLatency  int64
	Connections int64
	Outstanding int64
	Epoch       int32
	Counter     int32
	BuildTime   time.Time
	Mode        Mode
	Version     string
	Error       error
}

type requestHeader struct {
	Xid    int32
	Opcode int32
}

type responseHeader struct {
	Xid  int32
	Zxid int64
	Err  ErrCode
}

type multiHeader struct {
	Type int32
	Done bool
	Err  ErrCode
}

type auth struct {
	Type   int32
	Scheme string
	Auth   []byte
}

// Generic request structs

type pathRequest struct {
	Path string
}

type PathVersionRequest struct {
	Path    string
	Version int32
}

type pathWatchRequest struct {
	Path  string
	Watch bool
}

type pathResponse struct {
	Path string
}

type statResponse struct {
	Stat Stat
}

//

type CheckVersionRequest PathVersionRequest
type closeRequest struct{}
type closeResponse struct{}

type connectRequest struct {
	ProtocolVersion int32
	LastZxidSeen    int64
	TimeOut         int32
	SessionID       int64
	Passwd          []byte
}

type connectResponse struct {
	ProtocolVersion int32
	TimeOut         int32
	SessionID       int64
	Passwd          []byte
}

type CreateRequest struct {
	Path  string
	Data  []byte
	Acl   []ACL
	Flags int32
}

type createResponse pathResponse
type DeleteRequest PathVersionRequest
type deleteResponse struct{}

type errorResponse struct {
	Err int32
}

type existsRequest pathWatchRequest
type existsResponse statResponse
type getAclRequest pathRequest

type getAclResponse struct {
	Acl  []ACL
	Stat Stat
}

type getChildrenRequest pathRequest

type getChildrenResponse struct {
	Children []string
}

type getChildren2Request pathWatchRequest

type getChildren2Response struct {
	Children []string
	Stat     Stat
}

type getDataRequest pathWatchRequest

type getDataResponse struct {
	Data []byte
	Stat Stat
}

type getMaxChildrenRequest pathRequest

type getMaxChildrenResponse struct {
	Max int32
}

type getSaslRequest struct {
	Token []byte
}

type pingRequest struct{}
type pingResponse struct{}

type setAclRequest struct {
	Path    string
	Acl     []ACL
	Version int32
}

type setAclResponse statResponse

type SetDataRequest struct {
	Path    string
	Data    []byte
	Version int32
}

type setDataResponse statResponse

type setMaxChildren struct {
	Path string
	Max  int32
}

type setSaslRequest struct {
	Token string
}

type setSaslResponse struct {
	Token string
}

type setWatchesRequest struct {
	RelativeZxid int64
	DataWatches  []string
	ExistWatches []string
	ChildWatches []string
}

type setWatchesResponse struct{}

type syncRequest pathRequest
type syncResponse pathResponse

type setAuthRequest auth
type setAuthResponse struct{}

type multiRequestOp struct {
	Header multiHeader
	Op     interface{}
}
type multiRequest struct {
	Ops        []multiRequestOp
	DoneHeader multiHeader
}
type multiResponseOp struct {
	Header multiHeader
	String string
	Stat   *Stat
	Err    ErrCode
}
type multiResponse struct {
	Ops        []multiResponseOp
	DoneHeader multiHeader
}

type watchType int
type watchPathType struct {
	path  string
	wType watchType
}
