package stave

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/internal/parse"
)

func TestRenderTargetList_Watch(t *testing.T) {
	info := &parse.PkgInfo{
		PkgName: "main",
		Funcs: []*parse.Function{
			{
				Name:     "WatchTarget",
				IsWatch:  true,
				Synopsis: "A watch target",
			},
			{
				Name:     "NormalTarget",
				IsWatch:  false,
				Synopsis: "A normal target",
			},
		},
	}

	buf := &bytes.Buffer{}
	err := renderTargetList(buf, "stave", info, nil)
	require.NoError(t, err)

	output := buf.String()
	// Check if [W] is present for WatchTarget
	assert.Contains(t, output, "watchTarget [W]")
	// Check if NormalTarget does NOT have [W]
	assert.Contains(t, output, "normalTarget")
	assert.NotContains(t, output, "normalTarget [W]")

	// Check if legend is present
	assert.Contains(t, output, "[W] = watch target")
}
