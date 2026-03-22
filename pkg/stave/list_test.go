package stave

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yaklabco/stave/internal/parse"
)

func TestRenderTargetList_ColumnsAlignAcrossGroups(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	t.Setenv("COLUMNS", "120")

	info := &parse.PkgInfo{
		PkgName: "main",
		Funcs: []*parse.Function{
			// Local targets with short names.
			{Name: "Build", Synopsis: "Compile the project"},
			{Name: "Test", Synopsis: "Run the test suite"},
			// Namespace targets with longer names.
			{Name: "GenerateProto", Receiver: "Backend", Synopsis: "Generate protobuf stubs"},
			{Name: "RunMigrations", Receiver: "Backend", Synopsis: "Apply database migrations"},
		},
	}

	buf := &bytes.Buffer{}
	err := renderTargetList(buf, info, nil)
	require.NoError(t, err)

	output := buf.String()

	// Find column positions of NAME, USAGE, and SYNOPSIS in every header line.
	// Header lines are identified by containing all three column labels.
	lines := strings.Split(output, "\n")
	nameColumns := make([]int, 0, len(lines))
	usageColumns := make([]int, 0, len(lines))
	synopsisColumns := make([]int, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, "NAME") || !strings.Contains(line, "USAGE") || !strings.Contains(line, "SYNOPSIS") {
			continue
		}
		nameColumns = append(nameColumns, strings.Index(line, "NAME"))
		usageColumns = append(usageColumns, strings.Index(line, "USAGE"))
		synopsisColumns = append(synopsisColumns, strings.Index(line, "SYNOPSIS"))
	}

	require.GreaterOrEqual(t, len(nameColumns), 2, "expected at least two header rows (Local + Namespaces)")
	for group := 1; group < len(nameColumns); group++ {
		assert.Equal(t, nameColumns[0], nameColumns[group],
			"NAME column should be at the same position in all groups (group 0 vs group %d)", group)
		assert.Equal(t, usageColumns[0], usageColumns[group],
			"USAGE column should be at the same position in all groups (group 0 vs group %d)", group)
		assert.Equal(t, synopsisColumns[0], synopsisColumns[group],
			"SYNOPSIS column should be at the same position in all groups (group 0 vs group %d)", group)
	}
}

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
	err := renderTargetList(buf, info, nil)
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
