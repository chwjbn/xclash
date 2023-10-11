package context

import (
	"github.com/chwjbn/xclash/component/auth"
	"net"

	C "github.com/chwjbn/xclash/constant"

	"github.com/gofrs/uuid"
)

type ConnContext struct {
	id       uuid.UUID
	metadata *C.Metadata
	conn     net.Conn
	authUser  *auth.AuthUser
}

func NewConnContext(conn net.Conn, metadata *C.Metadata) *ConnContext {
	id, _ := uuid.NewV4()
	return &ConnContext{
		id:       id,
		metadata: metadata,
		conn:     conn,
	}
}

func NewConnContextWithAuth(conn net.Conn, metadata *C.Metadata,authUser auth.AuthUser) *ConnContext {
	id, _ := uuid.NewV4()
	return &ConnContext{
		id:       id,
		metadata: metadata,
		conn:     conn,
		authUser: &authUser,
	}
}

// ID implement C.ConnContext ID
func (c *ConnContext) ID() uuid.UUID {
	return c.id
}

// Metadata implement C.ConnContext Metadata
func (c *ConnContext) Metadata() *C.Metadata {
	return c.metadata
}

// Conn implement C.ConnContext Conn
func (c *ConnContext) Conn() net.Conn {
	return c.conn
}

func (c *ConnContext)AuthUser() *auth.AuthUser  {
	return c.authUser
}

func (c *ConnContext)SetAuthUser(authUser *auth.AuthUser) {
	c.authUser=authUser
}