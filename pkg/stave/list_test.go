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

	info := &parse.PkgInfo{
		PkgName: "main",
		Funcs: []*parse.Function{
			// Short local target.
			{Name: "Build", Synopsis: "Compile the project"},
			// Namespace with a long target name + args → wider USAGE column.
			{
				Name:     "VeryLongTargetName",
				Receiver: "LongNamespace",
				Synopsis: "Does something elaborate",
				Args:     []parse.Arg{{Name: "file"}, {Name: "output"}},
			},
		},
		Imports: []*parse.Import{
			{
				Name: "ext",
				Path: "example.com/ext",
				Info: parse.PkgInfo{
					PkgName: "ext",
					Funcs: []*parse.Function{
						{Name: "Run", Synopsis: "Run the external tool"},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := renderTargetList(&buf, info, nil)
	require.NoError(t, err)

	output := buf.String()
	lines := strings.Split(output, "\n")

	// Collect horizontal positions of "USAGE" and "SYNOPSIS" headers.
	var usagePositions, synPositions []int
	for _, line := range lines {
		if idx := strings.Index(line, "USAGE"); idx >= 0 {
			usagePositions = append(usagePositions, idx)
		}
		if idx := strings.Index(line, "SYNOPSIS"); idx >= 0 {
			synPositions = append(synPositions, idx)
		}
	}

	require.GreaterOrEqual(t, len(usagePositions), 3, "expected USAGE header in at least 3 groups, got %d", len(usagePositions))
	for i := 1; i < len(usagePositions); i++ {
		assert.Equal(t, usagePositions[0], usagePositions[i],
			"USAGE column misaligned: group 0 at %d, group %d at %d", usagePositions[0], i, usagePositions[i])
	}

	require.GreaterOrEqual(t, len(synPositions), 3, "expected SYNOPSIS header in at least 3 groups, got %d", len(synPositions))
	for i := 1; i < len(synPositions); i++ {
		assert.Equal(t, synPositions[0], synPositions[i],
			"SYNOPSIS column misaligned: group 0 at %d, group %d at %d", synPositions[0], i, synPositions[i])
	}

	// Verify data rows also align: every non-blank, non-header content line should
	// have its synopsis text starting at the same column as the SYNOPSIS header.
	synCol := synPositions[0]
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		// Skip header lines, section titles, and short lines that don't reach the synopsis column.
		if strings.Contains(line, "SYNOPSIS") || strings.Contains(line, "USAGE") {
			continue
		}
		if len(line) <= synCol {
			continue
		}
		// Data rows have synopsis text; check that the character at synCol is non-space
		// (i.e. the synopsis starts there, not shifted left or right). Skip continuation
		// lines (indented wrapped text) which start with spaces at synCol.
		if strings.HasPrefix(strings.TrimLeft(line, " "), "stave ") ||
			strings.HasPrefix(strings.TrimLeft(line, " "), "build") ||
			strings.HasPrefix(strings.TrimLeft(line, " "), "ext:") ||
			strings.HasPrefix(strings.TrimLeft(line, " "), "longNamespace:") {
			// This is a data row — the synopsis column should have non-space content.
			assert.NotEqual(t, ' ', rune(line[synCol]),
				"data row synopsis not aligned at column %d: %q", synCol, line)
		}
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
