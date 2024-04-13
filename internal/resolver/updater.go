package resolver

import (
	"context"
	"net"
)

type Updater interface {
	Update(context.Context, net.IP, string, string) error
}
