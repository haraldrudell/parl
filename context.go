/*
© 2018–present Harald Rudell <harald.rudell@gmail.com> (https://haraldrudell.github.io/haraldrudell/)
ISC License
*/

package parl

import (
	"context"
	"sync"
)

// Context is a context.Context with a Value() implementation
type Context struct {
	context.Context
	Map sync.Map
}

var _ context.Context = &Context{}

// NewContext provides a context.Context with a Value() implementation
func NewContext(ctx ...context.Context) (c *Context) {
	var ct context.Context
	if len(ctx) > 0 {
		ct = ctx[0]
	}
	if ct == nil {
		ct = context.Background()
	}
	return &Context{Context: ct}
}

// StoreInContext stores a value in context. Thread-safe
func StoreInContext(ctx context.Context, key string, value interface{}) (ok bool) {
	var c *Context
	if c, ok = ctx.(*Context); ok {
		c.Store(key, value)
	}
	return
}

// DelFromContext removes a value from context. Thread-safe
func DelFromContext(ctx context.Context, key string) (ok bool) {
	var c *Context
	if c, ok = ctx.(*Context); ok {
		c.Delete(key)
	}
	return
}

// Value retrieves data from context. Thread-safe
func (c *Context) Value(key interface{}) (result interface{}) {
	result, _ = c.Value2(key)
	return
}

// Value2 retrieves data and was–present indicator from context. Thread-safe
func (c *Context) Value2(key interface{}) (result interface{}, ok bool) {
	return c.Map.Load(key)
}

// StoreInContext stores a value in context. Thread-safe
func (c *Context) Store(key string, value interface{}) {
	c.Map.Store(key, value)
}

// StoreInContext stores a value in context. Thread-safe
func (c *Context) Delete(key string) {
	c.Map.Delete(key)
}
