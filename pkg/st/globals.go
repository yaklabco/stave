//nolint:gochecknoglobals // These are all intended as constants (and are private).
package st

import (
	"context"
	"reflect"
	"time"
)

// Here we define the types that are supported as arguments/returns.
var (
	ctxType   = reflect.TypeOf(func(context.Context) {}).In(0)
	errType   = reflect.TypeOf(func() error { return nil }).Out(0)
	emptyType = reflect.TypeOf(struct{}{})

	intType    = reflect.TypeOf(int(0))
	stringType = reflect.TypeOf(string(""))
	boolType   = reflect.TypeOf(bool(false))
	durType    = reflect.TypeOf(time.Second)

	// don't put ctx in here, this is for non-context types.
	argTypes = map[reflect.Type]bool{
		intType:    true,
		boolType:   true,
		stringType: true,
		durType:    true,
	}
)
