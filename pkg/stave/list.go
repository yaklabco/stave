package stave

import (
	"cmp"
	"fmt"
	"io"
	"maps"
	"os"
	"slices"
	"strconv"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/term"
	"github.com/muesli/reflow/wordwrap"
	"github.com/yaklabco/stave/internal/parse"
	"github.com/yaklabco/stave/pkg/st"
	"github.com/yaklabco/stave/pkg/ui"
)

type targetGroupKind int

const (
	targetGroupLocal targetGroupKind = iota
	targetGroupNamespace
	targetGroupImport
)

const (
	termWidthFloor    = 20
	fallbackTermWidth = 80
)

type targetKey struct {
	importPath string
	receiver   string
	name       string
}

type targetItem struct {
	key targetKey

	displayName string
	usage       string
	synopsis    string
	aliases     []string
	isDefault   bool
	isWatch     bool

	groupKind targetGroupKind
	groupName string // receiver name, import label, or empty for local
	groupMeta string // import path (when groupKind == import)
}

// renderTargetList renders the output of `stave -l`.
//
// It is implemented in the Stave binary (not in the generated mainfile) so it can
// use Charmbracelet styling without requiring additional dependencies in user projects.
func renderTargetList(out io.Writer, binaryName string, info *parse.PkgInfo, filters []string) error {
	items := buildTargetItems(binaryName, info)
	items = applyTargetFilters(items, filters)

	anyWatch := false
	for _, it := range items {
		if it.isWatch {
			anyWatch = true
			break
		}
	}

	cs := ui.GetFangScheme()
	colorEnabled := enableColorForList()
	const indent = "  "

	// Styles
	titleStyle := lipgloss.NewStyle().Bold(colorEnabled)
	sectionStyle := lipgloss.NewStyle().Bold(colorEnabled)
	subsectionStyle := lipgloss.NewStyle().Bold(colorEnabled)
	tableHeaderStyle := lipgloss.NewStyle().Bold(colorEnabled)
	defaultNameStyle := lipgloss.NewStyle().Bold(colorEnabled)
	targetStyle := st.TargetStyle()

	if colorEnabled {
		titleStyle = titleStyle.Foreground(cs.QuotedString)
		sectionStyle = sectionStyle.Foreground(cs.Program)
		subsectionStyle = subsectionStyle.Foreground(cs.Base)
		tableHeaderStyle = tableHeaderStyle.Foreground(cs.Base).Faint(true)
		defaultNameStyle = defaultNameStyle.Foreground(cs.Flag).Bold(true)
	}

	watchStyle := lipgloss.NewStyle()
	if colorEnabled {
		watchStyle = watchStyle.Foreground(cs.QuotedString).Reverse(true).Bold(true)
	}

	renderName := func(name string, isDefault, isWatch bool) string {
		if !colorEnabled {
			if isWatch {
				return name + " [W]"
			}
			return name
		}

		var renderedName string
		if isDefault {
			// Default target is highlighted with a distinct color so it is visually discoverable.
			renderedName = defaultNameStyle.Render(name)
		} else {
			// Non-default targets use the existing env-driven target color semantics.
			renderedName = targetStyle.Render(name)
		}

		if isWatch {
			renderedName += " " + watchStyle.Render("[W]")
		}
		return renderedName
	}

	// Header
	desc := strings.TrimSpace(info.Description)
	if desc != "" {
		width := detectTermWidth(out)
		usable := max(termWidthFloor, width) // ensure a sane floor
		wrapped := wordwrap.String(desc, usable)
		_, _ = fmt.Fprintln(out, wrapped)
		_, _ = fmt.Fprintln(out)
	}

	_, _ = fmt.Fprintln(out, titleStyle.Render("Targets:"))

	sections := groupTargets(items)
	writeSection := func(title string, groups []targetGroup) {
		if len(groups) == 0 {
			return
		}
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, sectionStyle.Render(title))
		for _, g := range groups {
			writeTable(out, tableHeaderStyle, subsectionStyle, g, renderName, indent)
		}
	}

	writeSection("Local", sections.local)
	writeSection("Namespaces", sections.namespaces)
	writeSection("Imports", sections.imports)

	if anyWatch {
		_, _ = fmt.Fprintln(out)
		_, _ = fmt.Fprintln(out, watchStyle.Render("[W]")+" = watch target")
	}

	return nil
}

func buildTargetItems(binaryName string, info *parse.PkgInfo) []targetItem {
	aliasByKey := make(map[targetKey][]string)
	for alias, fn := range info.Aliases {
		if fn == nil {
			continue
		}
		aliasKey := targetKey{importPath: fn.ImportPath, receiver: fn.Receiver, name: fn.Name}
		aliasByKey[aliasKey] = append(aliasByKey[aliasKey], lowerFirstTargetName(alias))
	}
	for aliasKey := range aliasByKey {
		slices.Sort(aliasByKey[aliasKey])
	}

	defaultKey := targetKey{}
	if info.DefaultFunc != nil {
		defaultKey = targetKey{importPath: info.DefaultFunc.ImportPath, receiver: info.DefaultFunc.Receiver, name: info.DefaultFunc.Name}
	}

	// Pre-allocate items with estimated capacity.
	itemCount := len(info.Funcs)
	for _, imp := range info.Imports {
		if imp != nil {
			itemCount += len(imp.Info.Funcs)
		}
	}
	items := make([]targetItem, 0, itemCount)

	// Local funcs
	for _, fn := range info.Funcs {
		if fn == nil {
			continue
		}
		funcKey := targetKey{importPath: fn.ImportPath, receiver: fn.Receiver, name: fn.Name}
		display := lowerFirstTargetName(fn.TargetName())
		items = append(items, targetItem{
			key:         funcKey,
			displayName: display,
			usage:       usageFor(binaryName, display, fn.Args),
			synopsis:    fn.Synopsis,
			aliases:     aliasByKey[funcKey],
			isDefault:   funcKey == defaultKey && fn.Name != "",
			isWatch:     fn.IsWatch,
			groupKind:   localGroupKind(fn),
			groupName:   localGroupName(fn),
		})
	}

	// Imports
	for _, imp := range info.Imports {
		if imp == nil {
			continue
		}
		label := imp.Name
		if imp.Alias != "" {
			label = imp.Alias
		}
		for _, fn := range imp.Info.Funcs {
			if fn == nil {
				continue
			}
			funcKey := targetKey{importPath: fn.ImportPath, receiver: fn.Receiver, name: fn.Name}
			display := lowerFirstTargetName(fn.TargetName())
			items = append(items, targetItem{
				key:         funcKey,
				displayName: display,
				usage:       usageFor(binaryName, display, fn.Args),
				synopsis:    fn.Synopsis,
				aliases:     aliasByKey[funcKey],
				isDefault:   funcKey == defaultKey && fn.Name != "",
				isWatch:     fn.IsWatch,
				groupKind:   targetGroupImport,
				groupName:   label,
				groupMeta:   imp.Path,
			})
		}
	}

	return items
}

func localGroupKind(fn *parse.Function) targetGroupKind {
	if fn.Receiver != "" {
		return targetGroupNamespace
	}
	return targetGroupLocal
}

func localGroupName(fn *parse.Function) string {
	if fn.Receiver == "" {
		return ""
	}
	return lowerFirstTargetName(fn.Receiver)
}

func usageFor(binaryName, display string, args []parse.Arg) string {
	var sb strings.Builder
	sb.WriteString(binaryName)
	sb.WriteString(" ")
	sb.WriteString(display)
	for _, a := range args {
		if strings.TrimSpace(a.Name) == "" {
			continue
		}
		sb.WriteString(" <")
		sb.WriteString(a.Name)
		sb.WriteString(">")
	}
	return sb.String()
}

func applyTargetFilters(items []targetItem, filters []string) []targetItem {
	if len(filters) == 0 {
		return items
	}

	needles := make([]string, 0, len(filters))
	for _, f := range filters {
		f = strings.TrimSpace(f)
		if f != "" {
			needles = append(needles, strings.ToLower(f))
		}
	}
	if len(needles) == 0 {
		return items
	}

	matchAll := func(haystack string) bool {
		haystack = strings.ToLower(haystack)
		for _, n := range needles {
			if !strings.Contains(haystack, n) {
				return false
			}
		}
		return true
	}

	out := make([]targetItem, 0, len(items))
	for _, it := range items {
		aliases := strings.Join(it.aliases, ", ")
		if matchAll(strings.Join([]string{it.displayName, it.usage, it.synopsis, aliases, it.groupName, it.groupMeta}, " ")) {
			out = append(out, it)
		}
	}
	return out
}

type targetGroup struct {
	header string
	meta   string // optional extra info, e.g. import path
	items  []targetItem
}

type targetSections struct {
	local      []targetGroup
	namespaces []targetGroup
	imports    []targetGroup
}

// compareTargetItems returns a comparison function for sorting targetItems by display name.
func compareTargetItems(a, b targetItem) int {
	return cmp.Compare(strings.ToLower(a.displayName), strings.ToLower(b.displayName))
}

// buildGroups converts a map of label->items into sorted groups with optional metadata.
func buildGroups(byLabel map[string][]targetItem, metaByLabel map[string]string) []targetGroup {
	labels := slices.Collect(maps.Keys(byLabel))
	slices.Sort(labels)

	groups := make([]targetGroup, 0, len(labels))
	for _, label := range labels {
		items := byLabel[label]
		slices.SortFunc(items, compareTargetItems)
		groups = append(groups, targetGroup{
			header: label,
			meta:   metaByLabel[label],
			items:  items,
		})
	}
	return groups
}

func groupTargets(items []targetItem) targetSections {
	var locals []targetItem
	nsByName := make(map[string][]targetItem)
	impByLabel := make(map[string][]targetItem)
	impMetaByLabel := make(map[string]string)

	for _, it := range items {
		switch it.groupKind {
		case targetGroupLocal:
			locals = append(locals, it)
		case targetGroupNamespace:
			nsByName[it.groupName] = append(nsByName[it.groupName], it)
		case targetGroupImport:
			impByLabel[it.groupName] = append(impByLabel[it.groupName], it)
			if it.groupMeta != "" {
				impMetaByLabel[it.groupName] = it.groupMeta
			}
		}
	}

	slices.SortFunc(locals, compareTargetItems)

	var localGroups []targetGroup
	if len(locals) > 0 {
		localGroups = append(localGroups, targetGroup{header: "", items: locals})
	}

	return targetSections{
		local:      localGroups,
		namespaces: buildGroups(nsByName, nil),
		imports:    buildGroups(impByLabel, impMetaByLabel),
	}
}

func writeTable(
	out io.Writer,
	headerStyle, subsectionStyle lipgloss.Style,
	group targetGroup,
	renderName func(name string, isDefault, isWatch bool) string,
	indent string,
) {
	if len(group.items) == 0 {
		return
	}

	if group.header != "" {
		_, _ = fmt.Fprintln(out)
		subtitle := group.header
		if group.meta != "" {
			subtitle = fmt.Sprintf("%s (%s)", group.header, group.meta)
		}
		_, _ = fmt.Fprintln(out, subsectionStyle.Render(subtitle))
	}

	type row struct {
		name      string
		usage     string
		synopsis  string
		isDefault bool
		isWatch   bool
	}

	rows := make([]row, 0, len(group.items)+1)
	rows = append(rows, row{
		name:     "NAME",
		usage:    "USAGE",
		synopsis: "SYNOPSIS",
	})

	for _, it := range group.items {
		syn := strings.TrimSpace(it.synopsis)
		if syn == "" {
			syn = "-"
		}
		name := it.displayName
		if len(it.aliases) > 0 {
			name = fmt.Sprintf("%s (%s)", name, strings.Join(it.aliases, ", "))
		}
		rows = append(rows, row{
			name:      name,
			usage:     it.usage,
			synopsis:  syn,
			isDefault: it.isDefault,
			isWatch:   it.isWatch,
		})
	}

	// Column widths (ANSI-aware via lipgloss.Width).
	maxName, maxUsage, maxSyn := 0, 0, 0
	for _, theRow := range rows {
		name := theRow.name
		if theRow.isWatch {
			name += " [W]"
		}
		maxName = max(maxName, lipgloss.Width(name))
		maxUsage = max(maxUsage, lipgloss.Width(theRow.usage))
		maxSyn = max(maxSyn, lipgloss.Width(theRow.synopsis))
	}

	pad := func(text string, width int) string {
		if width <= 0 {
			return text
		}
		textWidth := lipgloss.Width(text)
		if textWidth >= width {
			return text
		}
		return text + strings.Repeat(" ", width-textWidth)
	}

	// Print header.
	h := rows[0]
	headerLine := strings.Join([]string{
		pad(h.name, maxName),
		pad(h.usage, maxUsage),
		h.synopsis,
	}, "  ")
	_, _ = fmt.Fprintln(out, indent+headerStyle.Render(headerLine))

	// Compute terminal width and synopsis column width for wrapping.
	termWidth := detectTermWidth(out)
	const gap = 2
	leftOffset := lipgloss.Width(indent) + maxName + gap + maxUsage + gap
	synWidth := termWidth - leftOffset
	if synWidth < termWidthFloor {
		synWidth = termWidthFloor
	}

	spaceLeft := strings.Repeat(" ", leftOffset)

	// Print rows with word-wrapped synopsis using a hanging indent.
	for _, theRow := range rows[1:] {
		name := renderName(theRow.name, theRow.isDefault, theRow.isWatch)

		wrappedSyn := wordwrap.String(theRow.synopsis, synWidth)
		// Align continuation lines under the start of the synopsis column.
		wrappedSyn = strings.ReplaceAll(wrappedSyn, "\n", "\n"+spaceLeft)

		line := strings.Join([]string{
			pad(name, maxName),
			pad(theRow.usage, maxUsage),
			wrappedSyn,
		}, strings.Repeat(" ", gap))
		_, _ = fmt.Fprintln(out, indent+line)
	}
}

func enableColorForList() bool {
	// Use auto-detection (like stave --version), not opt-in env var.
	// Respects NO_COLOR and TERM blacklist via st.ColorEnabled().
	return st.ColorEnabled()
}

// detectTermWidth returns the terminal width to use for wrapping.
// It prefers the actual stdout size, falls back to $COLUMNS, then 80.
func detectTermWidth(_ io.Writer) int {
	if w, _, err := term.GetSize(os.Stdout.Fd()); err == nil && w > 0 {
		return w
	}
	if cols := os.Getenv("COLUMNS"); cols != "" {
		if v, err := strconv.Atoi(cols); err == nil && v > 0 {
			return v
		}
	}

	return fallbackTermWidth
}

func lowerFirstTargetName(s string) string {
	parts := strings.Split(s, ":")
	for i := range parts {
		parts[i] = lowerFirstWord(parts[i])
	}
	return strings.Join(parts, ":")
}
