package drivers

import (
	"context"
	"io"
	"time"

	"github.com/MontFerret/ferret/pkg/runtime/core"
)

const DefaultTimeout = time.Second * 30

type (
	ctxKey struct{}

	ctxValue struct {
		opts    *options
		drivers map[string]Driver
	}

	OpenPageParams struct {
		URL         string
		UserAgent   string
		KeepCookies bool
		Cookies     []HTTPCookie
		Header      HTTPHeader
	}

	Driver interface {
		io.Closer
		Name() string
		Open(ctx context.Context, params OpenPageParams) (HTMLPage, error)
	}
)

func WithContext(ctx context.Context, drv Driver, opts ...Option) context.Context {
	ctx, value := resolveValue(ctx)

	value.drivers[drv.Name()] = drv

	for _, opt := range opts {
		opt(drv, value.opts)
	}

	// set first registered driver as a default one
	if value.opts.defaultDriver == "" {
		value.opts.defaultDriver = drv.Name()
	}

	return ctx
}

func FromContext(ctx context.Context, name string) (Driver, error) {
	_, value := resolveValue(ctx)

	if name == "" {
		name = value.opts.defaultDriver
	}

	drv, exists := value.drivers[name]

	if !exists {
		return nil, core.Error(core.ErrNotFound, name)
	}

	return drv, nil
}

func resolveValue(ctx context.Context) (context.Context, *ctxValue) {
	key := ctxKey{}
	v := ctx.Value(key)
	value, ok := v.(*ctxValue)

	if !ok {
		value = &ctxValue{
			opts:    &options{},
			drivers: make(map[string]Driver),
		}

		return context.WithValue(ctx, key, value), value
	}

	return ctx, value
}
