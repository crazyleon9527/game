package mock

import (
	"github.com/google/wire"
)

// FreeSet 注入Free
var FreeSet = wire.NewSet(wire.Struct(new(Free), "*"))

type Free struct {
}
