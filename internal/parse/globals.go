//nolint:gochecknoglobals // These are all intended as constants (and are private).
package parse

const (
	stringType  = "string"
	intType     = "int"
	float64Type = "float64"
	boolType    = "bool"
	timeType    = "time.Duration"
)

var argTypes = map[string]string{
	stringType:         stringType,
	intType:            intType,
	float64Type:        float64Type,
	boolType:           boolType,
	"&{time Duration}": timeType,
}
