package main

import (
	"context"
	"fmt"
	"image/color"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vitaminmoo/memtools/process"
	"noitrainer/noita"
)

// ringLog is a thread-safe ring buffer that implements io.Writer for log output.
type ringLog struct {
	mu    sync.Mutex
	lines []string
	max   int
}

func newRingLog(max int) *ringLog {
	return &ringLog{max: max}
}

func (r *ringLog) Write(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	// Split on newlines, append non-empty lines.
	for _, line := range strings.Split(strings.TrimRight(string(p), "\n"), "\n") {
		if line == "" {
			continue
		}
		r.lines = append(r.lines, line)
		if len(r.lines) > r.max {
			r.lines = r.lines[len(r.lines)-r.max:]
		}
	}
	return len(p), nil
}

func (r *ringLog) Lines() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]string, len(r.lines))
	copy(out, r.lines)
	return out
}

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	sectionStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1).
			MarginBottom(1)

	sectionTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(lipgloss.Color("#FF79C6"))

	labelStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#8BE9FD")).
			Width(22)

	valueStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F8F8F2"))

	hpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FF5555")).
		Bold(true)

	goldStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F1FA8C")).
			Bold(true)

	manaStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4")).
			Bold(true)

	posStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#50FA7B"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5555")).
			Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6272A4"))

	tabActiveStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#F8F8F2")).
			Background(lipgloss.Color("#7D56F4")).
			Padding(0, 1)

	tabInactiveStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#6272A4")).
				Padding(0, 1)

	detailPaneStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#7D56F4")).
			Padding(0, 1)
)

// ── Entity categorization ──────────────────────────────────────────

type entityCategory int

const (
	catPlayer entityCategory = iota
	catEnemy
	catItem
	catTorch
	catPhysics
	catProp
	catEffect
	catOther
)

func (c entityCategory) String() string {
	switch c {
	case catPlayer:
		return "Player"
	case catEnemy:
		return "Enemy"
	case catItem:
		return "Item"
	case catTorch:
		return "Torch"
	case catPhysics:
		return "Physics"
	case catProp:
		return "Prop"
	case catEffect:
		return "Effect"
	default:
		return "Other"
	}
}

func (c entityCategory) color() string {
	switch c {
	case catPlayer:
		return "#50FA7B"
	case catEnemy:
		return "#FF5555"
	case catItem:
		return "#F1FA8C"
	case catTorch:
		return "#FFB86C"
	case catPhysics:
		return "#6272A4"
	case catProp:
		return "#8BE9FD"
	case catEffect:
		return "#BD93F9"
	default:
		return "#6272A4"
	}
}

func hexToRGBA(hex string) color.RGBA {
	hex = strings.TrimPrefix(hex, "#")
	r, _ := strconv.ParseUint(hex[0:2], 16, 8)
	g, _ := strconv.ParseUint(hex[2:4], 16, 8)
	b, _ := strconv.ParseUint(hex[4:6], 16, 8)
	return color.RGBA{R: uint8(r), G: uint8(g), B: uint8(b), A: 255}
}

// overlayOption is a toggleable overlay setting.
type overlayOption int

const (
	optHideAtOrigin    overlayOption = iota
	optHideAtPlayer
	optHideInventory
	optShowEntityIDs
	optShowLabels
	optionCount // sentinel
)

func (o overlayOption) String() string {
	switch o {
	case optHideAtOrigin:
		return "Hide entities at (0, 0)"
	case optHideAtPlayer:
		return "Hide entities at player pos"
	case optHideInventory:
		return "Hide player inventory items"
	case optShowEntityIDs:
		return "Show entity IDs"
	case optShowLabels:
		return "Show text labels"
	default:
		return "?"
	}
}

// overlayCategories lists the categories available for overlay toggle (excludes Player).
var overlayCategories = []entityCategory{catEnemy, catItem, catTorch, catProp, catEffect, catOther}

func categorize(e *noita.EntitySummary) entityCategory {
	has := make(map[noita.TypeID]bool)
	for _, id := range e.ComponentIDs {
		has[id] = true
	}

	if strings.Contains(e.Name, "player") || e.Name == "arm_r" || e.Name == "cape" ||
		strings.HasPrefix(e.Name, "inventory_") || e.Name == "player_stats" {
		return catPlayer
	}
	if has[noita.TypeIDAnimalAIComponent] || strings.HasPrefix(e.Name, "$animal_") {
		return catEnemy
	}
	if has[noita.TypeIDTorchComponent] {
		return catTorch
	}
	if has[noita.TypeIDItemComponent] || has[noita.TypeIDAbilityComponent] {
		return catItem
	}
	if has[noita.TypeIDGameEffectComponent] {
		return catEffect
	}
	if has[noita.TypeIDVerletPhysicsComponent] && !has[noita.TypeIDDamageModelComponent] {
		return catProp
	}
	if has[noita.TypeIDSimplePhysicsComponent] || has[noita.TypeIDPixelSpriteComponent] {
		return catPhysics
	}
	return catOther
}

func subcategorize(e *noita.EntitySummary) string {
	has := make(map[noita.TypeID]bool)
	for _, id := range e.ComponentIDs {
		has[id] = true
	}

	cat := categorize(e)
	switch cat {
	case catPlayer:
		switch {
		case strings.Contains(e.Name, "player"):
			return "player"
		case e.Name == "arm_r" || e.Name == "cape" || e.Name == "hand_l":
			return "body"
		case strings.HasPrefix(e.Name, "inventory_"):
			return "inventory"
		default:
			return "stats"
		}
	case catEnemy:
		name := e.Name
		if strings.HasPrefix(name, "$animal_") {
			return strings.TrimPrefix(name, "$animal_")
		}
		return "unknown"
	case catItem:
		switch {
		case has[noita.TypeIDPotionComponent]:
			return "potion"
		case has[noita.TypeIDAbilityComponent] && has[noita.TypeIDManaReloaderComponent]:
			return "wand"
		case has[noita.TypeIDItemActionComponent]:
			return "spell"
		case has[noita.TypeIDAbilityComponent]:
			return "holdable"
		default:
			return "pickup"
		}
	case catTorch:
		if has[noita.TypeIDPhysicsBody2Component] {
			return "physics"
		}
		return "static"
	case catPhysics:
		if has[noita.TypeIDLuaComponent] {
			return "scripted"
		}
		return "debris"
	case catProp:
		if has[noita.TypeIDVerletWorldJointComponent] {
			return "hanging"
		}
		return "loose"
	case catEffect:
		if has[noita.TypeIDInheritTransformComponent] {
			return "attached"
		}
		return "standalone"
	default:
		if has[noita.TypeIDCollisionTriggerComponent] {
			return "trigger"
		}
		if has[noita.TypeIDVariableStorageComponent] {
			return "variable"
		}
		if has[noita.TypeIDWorldStateComponent] {
			return "world"
		}
		if has[noita.TypeIDCameraBoundComponent] {
			return "camera"
		}
		return "misc"
	}
}

func entityDisplayName(e *noita.EntitySummary) string {
	name := e.Name
	if name == "" || name == "unknown" {
		// Fall back to subcategory as display name.
		return subcategorize(e)
	}
	if strings.HasPrefix(name, "$animal_") {
		return strings.TrimPrefix(name, "$animal_")
	}
	if strings.HasPrefix(name, "DEBUG_NAME:") {
		return strings.TrimPrefix(name, "DEBUG_NAME:")
	}
	return name
}

// ── List items ─────────────────────────────────────────────────────

type entityItem struct {
	summary     *noita.EntitySummary
	category    entityCategory
	subcategory string
}

func (i entityItem) Title() string {
	return fmt.Sprintf("#%d %s", i.summary.Entity.EntityId, entityDisplayName(i.summary))
}

func (i entityItem) Description() string {
	e := i.summary
	parts := []string{
		fmt.Sprintf("%s/%s", i.category.String(), i.subcategory),
		fmt.Sprintf("(%.0f, %.0f)", e.Entity.PosX, e.Entity.PosY),
	}
	if e.HasHP {
		parts = append(parts, "HP")
	}
	return strings.Join(parts, "  ")
}

func (i entityItem) FilterValue() string {
	return fmt.Sprintf("%d %s %s %s", i.summary.Entity.EntityId, i.category.String(), i.subcategory, entityDisplayName(i.summary))
}

// ── Model ──────────────────────────────────────────────────────────

type tickMsg time.Time

type model struct {
	state         *noita.GameState
	reader        *noita.Reader
	proc          *process.Process
	tab           int
	tabs          []string
	width         int
	height        int
	err           error
	quitting      bool
	entityList    list.Model
	entityDetails *noita.EntityDetails
	listReady     bool
	categoryCounts map[entityCategory]int
	overlay       *overlayScene
	overlayCancel context.CancelFunc
	logBuf        *ringLog

	// Overlay tab state
	overlayCats    map[entityCategory]bool // which categories to render
	overlayOpts    map[overlayOption]bool  // option toggles
	overlayCursor  int                     // cursor position in overlay tab
}

// entityDelegate wraps the default delegate to apply per-item category colors.
type entityDelegate struct {
	base list.DefaultDelegate
}

func (d entityDelegate) Height() int                             { return d.base.Height() }
func (d entityDelegate) Spacing() int                            { return d.base.Spacing() }
func (d entityDelegate) Update(msg tea.Msg, m *list.Model) tea.Cmd { return d.base.Update(msg, m) }

func (d entityDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	ei, ok := item.(entityItem)
	if !ok {
		d.base.Render(w, m, index, item)
		return
	}

	if m.Width() <= 0 {
		return
	}

	title := ei.Title()
	desc := ei.Description()
	s := d.base.Styles
	catColor := lipgloss.Color(ei.category.color())

	textwidth := m.Width() - s.NormalTitle.GetPaddingLeft() - s.NormalTitle.GetPaddingRight()
	title = truncateAnsi(title, textwidth)
	desc = truncateAnsi(desc, textwidth)

	// Split title into ID portion and name portion.
	idPart, namePart := title, ""
	if sp := strings.Index(title, " "); sp >= 0 {
		idPart = title[:sp]
		namePart = title[sp:]
	}

	isSelected := index == m.Index()
	emptyFilter := m.FilterState() == list.Filtering && m.FilterValue() == ""

	if emptyFilter {
		title = s.DimmedTitle.Render(title)
		desc = s.DimmedDesc.Render(desc)
	} else if isSelected && m.FilterState() != list.Filtering {
		idStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2")).Render(idPart)
		nameStyled := lipgloss.NewStyle().Foreground(catColor).Render(namePart)
		title = s.SelectedTitle.Render(idStyled + nameStyled)
		desc = s.SelectedDesc.Render(desc)
	} else {
		idStyled := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#F8F8F2")).Render(idPart)
		nameStyled := lipgloss.NewStyle().Foreground(catColor).Render(namePart)
		title = s.NormalTitle.Render(idStyled + nameStyled)
		desc = s.NormalDesc.Render(desc)
	}

	fmt.Fprintf(w, "%s\n%s", title, desc)
}

func entityFilter(term string, targets []string) []list.Rank {
	// If the term looks like an ID search (#123 or just digits), do substring match on ID.
	idSearch := ""
	if strings.HasPrefix(term, "#") {
		idSearch = term[1:]
	} else if _, err := strconv.Atoi(term); err == nil {
		idSearch = term
	}

	if idSearch != "" {
		var results []list.Rank
		for i, t := range targets {
			// FilterValue starts with the entity ID followed by a space.
			targetID := t
			if sp := strings.IndexByte(t, ' '); sp > 0 {
				targetID = t[:sp]
			}
			if strings.Contains(targetID, idSearch) {
				results = append(results, list.Rank{Index: i})
			}
		}
		return results
	}

	return list.DefaultFilter(term, targets)
}

func newEntityList() list.Model {
	dd := list.NewDefaultDelegate()
	dd.Styles.SelectedTitle = dd.Styles.SelectedTitle.
		BorderLeftForeground(lipgloss.Color("#FF79C6"))
	dd.Styles.SelectedDesc = dd.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#6272A4")).
		BorderLeftForeground(lipgloss.Color("#FF79C6"))
	delegate := entityDelegate{base: dd}

	l := list.New(nil, delegate, 40, 20)
	l.Title = "Entities"
	l.Filter = entityFilter
	l.Styles.Title = sectionTitleStyle
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	return l
}

func initialModel(logBuf *ringLog) model {
	ctx, cancel := context.WithCancel(context.Background())
	ov := startOverlay(ctx)
	return model{
		state:          &noita.GameState{},
		tabs:           []string{"Player", "Entities", "Wands", "World", "Overlay", "Log"},
		entityList:     newEntityList(),
		categoryCounts: make(map[entityCategory]int),
		overlay:        ov,
		overlayCancel:  cancel,
		logBuf:         logBuf,
		overlayCats: map[entityCategory]bool{
			catEnemy: true,
		},
		overlayOpts: map[overlayOption]bool{
			optHideAtOrigin:  true,
			optHideAtPlayer:  true,
			optHideInventory: true,
			optShowLabels:    true,
		},
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd { return tickCmd() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.tab == 1 && m.entityList.FilterState() == list.Filtering {
			var cmd tea.Cmd
			m.entityList, cmd = m.entityList.Update(msg)
			return m, cmd
		}

		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.quitting = true
			if m.overlayCancel != nil {
				m.overlayCancel()
			}
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab"))):
			m.tab = (m.tab + 1) % len(m.tabs)
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab"))):
			m.tab = (m.tab - 1 + len(m.tabs)) % len(m.tabs)
		default:
			if m.tab == 1 {
				var cmd tea.Cmd
				m.entityList, cmd = m.entityList.Update(msg)
				return m, cmd
			}
			if m.tab == 4 { // Overlay tab
				switch msg.String() {
				case "up", "k":
					if m.overlayCursor > 0 {
						m.overlayCursor--
					}
				case "down", "j":
					total := len(overlayCategories) + int(optionCount)
					if m.overlayCursor < total-1 {
						m.overlayCursor++
					}
				case " ", "enter":
					catCount := len(overlayCategories)
					if m.overlayCursor < catCount {
						cat := overlayCategories[m.overlayCursor]
						m.overlayCats[cat] = !m.overlayCats[cat]
					} else {
						opt := overlayOption(m.overlayCursor - catCount)
						m.overlayOpts[opt] = !m.overlayOpts[opt]
					}
				}
			}
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		listHeight := msg.Height - 7
		if listHeight < 5 {
			listHeight = 5
		}
		listWidth := msg.Width * 2 / 5
		if listWidth < 30 {
			listWidth = 30
		}
		m.entityList.SetSize(listWidth, listHeight)

	case tickMsg:
		m.tryConnect()
		if m.reader != nil {
			m.state = m.reader.ReadState()
			m.updateEntityList()
			m.updateEntityDetails()
			m.updateOverlay()
		}
		return m, tickCmd()

	default:
		// Route other messages (e.g. FilterMatchesMsg) to the entity list.
		var cmd tea.Cmd
		m.entityList, cmd = m.entityList.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m *model) updateEntityList() {
	if m.state == nil || len(m.state.Entities) == 0 {
		if m.listReady {
			m.entityList.SetItems(nil)
			m.listReady = false
		}
		return
	}

	var selectedPtr uint32
	if sel, ok := m.entityList.SelectedItem().(entityItem); ok {
		selectedPtr = sel.summary.Ptr
	}

	counts := make(map[entityCategory]int)
	var entityItems []entityItem

	for _, e := range m.state.Entities {
		cat := categorize(e)
		counts[cat]++
		// Skip physics debris — too noisy
		if cat == catPhysics {
			continue
		}
		entityItems = append(entityItems, entityItem{
			summary:     e,
			category:    cat,
			subcategory: subcategorize(e),
		})
	}
	m.categoryCounts = counts

	// Sort by category, then subcategory, then entity ID
	sort.Slice(entityItems, func(i, j int) bool {
		a, b := entityItems[i], entityItems[j]
		if a.category != b.category {
			return a.category < b.category
		}
		if a.subcategory != b.subcategory {
			return a.subcategory < b.subcategory
		}
		return a.summary.Entity.EntityId < b.summary.Entity.EntityId
	})

	items := make([]list.Item, len(entityItems))
	newSelectedIdx := 0
	for i, ei := range entityItems {
		items[i] = ei
		if ei.summary.Ptr == selectedPtr {
			newSelectedIdx = i
		}
	}

	if m.entityList.FilterState() == list.Unfiltered {
		m.entityList.SetItems(items)
		if selectedPtr != 0 {
			m.entityList.Select(newSelectedIdx)
		}
	}
	m.listReady = true
}

func (m *model) updateEntityDetails() {
	if m.reader == nil {
		m.entityDetails = nil
		return
	}
	sel, ok := m.entityList.SelectedItem().(entityItem)
	if !ok {
		m.entityDetails = nil
		return
	}
	m.entityDetails = m.reader.ReadEntityDetails(sel.summary.Ptr)
}

func (m *model) updateOverlay() {
	if m.overlay == nil {
		return
	}
	// Give the overlay its own Process handle for the same PID (not thread-safe to share).
	if m.proc != nil {
		m.overlay.proc.CompareAndSwap(nil, process.New(m.proc.PID))
	}
	// Tell the overlay which entity is selected.
	sel := &overlaySelection{}
	if s, ok := m.entityList.SelectedItem().(entityItem); ok {
		sel.EntityPtr = s.summary.Ptr
	}
	m.overlay.selection.Store(sel)

	// Build filtered entity list for overlay rendering.
	var ents []overlayEntity
	if m.state != nil {
		var playerX, playerY float32
		if pe := m.state.PlayerEntity; pe != nil {
			playerX = pe.PosX
			playerY = pe.PosY
		}

		// Build set of entity pointers that are in the player's inventory tree.
		var inventoryPtrs map[uint32]bool
		if m.overlayOpts[optHideInventory] && m.state.PlayerEntity != nil {
			// Find the player entity pointer from the summary list.
			var playerPtr uint32
			playerID := m.state.PlayerEntity.EntityId
			byPtr := make(map[uint32]*noita.EntitySummary, len(m.state.Entities))
			for _, e := range m.state.Entities {
				byPtr[e.Ptr] = e
				if e.Entity.EntityId == playerID {
					playerPtr = e.Ptr
				}
			}
			if playerPtr != 0 {
				inventoryPtrs = make(map[uint32]bool)
				for _, e := range m.state.Entities {
					// Walk parent chain to see if it reaches the player.
					ptr := e.Entity.ParentEntityPtr
					for depth := 0; ptr != 0 && depth < 10; depth++ {
						if ptr == playerPtr {
							inventoryPtrs[e.Ptr] = true
							break
						}
						if p, ok := byPtr[ptr]; ok {
							ptr = p.Entity.ParentEntityPtr
						} else {
							break
						}
					}
				}
			}
		}

		for _, e := range m.state.Entities {
			cat := categorize(e)
			if !m.overlayCats[cat] {
				continue
			}
			if inventoryPtrs[e.Ptr] {
				continue
			}
			x, y := e.Entity.PosX, e.Entity.PosY

			if m.overlayOpts[optHideAtOrigin] && x == 0 && y == 0 {
				continue
			}
			if m.overlayOpts[optHideAtPlayer] && x == playerX && y == playerY {
				continue
			}

			name := ""
			if m.overlayOpts[optShowLabels] {
				name = entityDisplayName(e)
			}
			if m.overlayOpts[optShowEntityIDs] {
				id := fmt.Sprintf("#%d", e.Entity.EntityId)
				if name != "" {
					name = id + " " + name
				} else {
					name = id
				}
			}

			oe := overlayEntity{
				X:     x,
				Y:     y,
				Name:  name,
				Color: hexToRGBA(cat.color()),
			}
			if hb := e.Hitbox; hb != nil {
				oe.HasHitbox = true
				oe.AabbMinX = hb.AabbMinX
				oe.AabbMaxX = hb.AabbMaxX
				oe.AabbMinY = hb.AabbMinY
				oe.AabbMaxY = hb.AabbMaxY
				oe.HitOffsetX = hb.OffsetX
				oe.HitOffsetY = hb.OffsetY
			} else if ct := e.CollisionTrigger; ct != nil {
				oe.HasHitbox = true
				oe.AabbMinX = -ct.Width / 2
				oe.AabbMaxX = ct.Width / 2
				oe.AabbMinY = -ct.Height / 2
				oe.AabbMaxY = ct.Height / 2
			}
			ents = append(ents, oe)
		}
	}
	m.overlay.entities.Store(&ents)
}

func (m *model) tryConnect() {
	if m.proc != nil {
		return
	}
	procs, err := process.FromName("noita.exe")
	if err != nil {
		m.err = err
		m.state = &noita.GameState{Error: "Noita not found — is it running?"}
		return
	}
	m.proc = procs[0]
	m.reader = noita.NewReader(m.proc)
	m.err = nil
}

// ── View ───────────────────────────────────────────────────────────

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	title := titleStyle.Render(" NOITRAINER ")
	if m.state.Connected {
		title += dimStyle.Render("  connected")
	} else {
		title += errorStyle.Render("  disconnected")
	}
	b.WriteString(title + "\n\n")

	var tabs []string
	for i, t := range m.tabs {
		if i == m.tab {
			tabs = append(tabs, tabActiveStyle.Render(t))
		} else {
			tabs = append(tabs, tabInactiveStyle.Render(t))
		}
	}
	b.WriteString(lipgloss.JoinHorizontal(lipgloss.Top, tabs...) + "\n\n")

	if !m.state.Connected {
		b.WriteString(errorStyle.Render(m.state.Error) + "\n")
		b.WriteString(dimStyle.Render("Waiting for noita.exe...") + "\n")
		b.WriteString(dimStyle.Render("Press q to quit") + "\n")
		return b.String()
	}

	switch m.tab {
	case 0:
		b.WriteString(m.viewPlayer())
	case 1:
		b.WriteString(m.viewEntities())
	case 2:
		b.WriteString(m.viewWands())
	case 3:
		b.WriteString(m.viewWorld())
	case 4:
		b.WriteString(m.viewOverlay())
	case 5:
		b.WriteString(m.viewLog())
	}

	b.WriteString("\n" + dimStyle.Render("tab/shift+tab: switch tabs  /: filter  q: quit"))
	return b.String()
}

// ── Player tab ─────────────────────────────────────────────────────

func (m model) viewPlayer() string {
	var cols []string

	// Left column: movement + health
	{
		var sections []string
		if e := m.state.PlayerEntity; e != nil {
			rows := []string{
				row("Position", posStyle.Render(fmt.Sprintf("%.1f, %.1f", e.PosX, e.PosY))),
				row("Entity ID", fmt.Sprintf("%d", e.EntityId)),
			}
			if c := m.state.PlayerChar; c != nil {
				rows = append(rows,
					row("Velocity", fmt.Sprintf("%.1f, %.1f", c.VelocityX, c.VelocityY)),
					row("On Ground", boolStr(c.IsOnGround)),
					row("Fly Time", fmt.Sprintf("%.1f", c.FlyTime)),
				)
			}
			sections = append(sections, renderSection("Movement", rows))
		}

		if d := m.state.PlayerHP; d != nil {
			hp := d.Hp * 25
			maxHp := d.MaxHp * 25
			rows := []string{
				row("HP", hpStyle.Render(fmt.Sprintf("%.0f / %.0f", hp, maxHp))),
				row("Max HP Cap", fmt.Sprintf("%.0f", d.MaxHpCap*25)),
				row("I-Frames", fmt.Sprintf("%d", d.InvincibilityFrames)),
			}
			mults := dmgMultsNonDefault(d)
			if len(mults) > 0 {
				rows = append(rows, row("Dmg Mults", strings.Join(mults, " ")))
			}
			sections = append(sections, renderSection("Health", rows))
		}
		if len(sections) > 0 {
			cols = append(cols, strings.Join(sections, ""))
		}
	}

	// Right column: wallet + world summary
	{
		var sections []string
		if w := m.state.PlayerWallet; w != nil {
			rows := []string{
				row("Gold", goldStyle.Render(fmt.Sprintf("%d", w.Money))),
				row("Gold Spent", fmt.Sprintf("%d", w.MoneySpent)),
			}
			sections = append(sections, renderSection("Wallet", rows))
		}

		// Entity summary counts
		if len(m.categoryCounts) > 0 {
			order := []entityCategory{catEnemy, catItem, catTorch, catProp, catPhysics, catEffect, catOther}
			var rows []string
			total := 0
			for _, cat := range order {
				n := m.categoryCounts[cat]
				if n > 0 {
					total += n
					rows = append(rows, row(cat.String(),
						lipgloss.NewStyle().Foreground(lipgloss.Color(cat.color())).Render(fmt.Sprintf("%d", n))))
				}
			}
			rows = append(rows, row("Total", fmt.Sprintf("%d", total+m.categoryCounts[catPlayer])))
			sections = append(sections, renderSection("Entities", rows))
		}

		if g := m.state.Globals; g != nil {
			rows := []string{
				row("Seed", fmt.Sprintf("%d", m.state.WorldSeed)),
				row("Frame", fmt.Sprintf("%d", g.FrameCount)),
				row("Deaths", fmt.Sprintf("%d", m.state.DeathCount)),
				row("Camera", posStyle.Render(fmt.Sprintf("%.0f, %.0f", m.state.CameraX, m.state.CameraY))),
			}
			sections = append(sections, renderSection("World", rows))
		}
		if len(sections) > 0 {
			cols = append(cols, strings.Join(sections, ""))
		}
	}

	if len(cols) == 0 {
		return dimStyle.Render("No player data available")
	}
	return lipgloss.JoinHorizontal(lipgloss.Top, cols...)
}

// ── Entities tab ───────────────────────────────────────────────────

func (m model) viewEntities() string {
	listView := m.entityList.View()

	detailWidth := m.width - lipgloss.Width(listView) - 4
	if detailWidth < 30 {
		detailWidth = 30
	}
	detailHeight := m.height - 7
	if detailHeight < 5 {
		detailHeight = 5
	}

	detail := m.renderEntityDetail()
	detailPane := detailPaneStyle.
		Width(detailWidth).
		Height(detailHeight).
		Render(detail)

	return lipgloss.JoinHorizontal(lipgloss.Top, listView, " ", detailPane)
}

func (m model) renderEntityDetail() string {
	d := m.entityDetails
	if d == nil {
		return dimStyle.Render("Select an entity to view details")
	}

	var sections []string

	// Header
	{
		name := entityDisplayName(&noita.EntitySummary{Entity: d.Entity, Name: d.Name})
		rows := []string{
			row("Name", name),
			row("ID / Slot", fmt.Sprintf("%d / %d", d.Entity.EntityId, d.Entity.SlotIndex)),
			row("Position", posStyle.Render(fmt.Sprintf("%.1f, %.1f", d.Entity.PosX, d.Entity.PosY))),
		}
		sections = append(sections, renderSection("Entity", rows))
	}

	if d.HP != nil {
		hp := d.HP.Hp * 25
		maxHp := d.HP.MaxHp * 25
		rows := []string{
			row("HP", hpStyle.Render(fmt.Sprintf("%.0f / %.0f", hp, maxHp))),
		}
		if d.HP.InvincibilityFrames > 0 {
			rows = append(rows, row("I-Frames", fmt.Sprintf("%d", d.HP.InvincibilityFrames)))
		}
		mults := dmgMultsNonDefault(d.HP)
		if len(mults) > 0 {
			rows = append(rows, row("Dmg Mults", strings.Join(mults, " ")))
		}
		sections = append(sections, renderSection("Health", rows))
	}

	if d.Char != nil {
		rows := []string{
			row("Velocity", fmt.Sprintf("%.1f, %.1f", d.Char.VelocityX, d.Char.VelocityY)),
			row("On Ground", boolStr(d.Char.IsOnGround)),
		}
		sections = append(sections, renderSection("Movement", rows))
	}

	if d.Wallet != nil {
		sections = append(sections, renderSection("Wallet", []string{
			row("Gold", goldStyle.Render(fmt.Sprintf("%d", d.Wallet.Money))),
		}))
	}

	if d.Ability != nil {
		a := d.Ability
		gc := a.GunConfig
		rows := []string{
			row("Mana", manaStyle.Render(fmt.Sprintf("%.0f / %.0f", a.Mana, a.ManaMax))),
			row("Spells/Cast", fmt.Sprintf("%d", gc.ActionsPerRound)),
			row("Deck", fmt.Sprintf("%d  %s  %.2fs reload",
				gc.DeckCapacity, shuffleStr(gc.ShuffleDeckWhenEmpty), float64(gc.ReloadTime)/60.0)),
		}
		sections = append(sections, renderSection("Ability", rows))
	}

	if d.Item != nil {
		itemName := d.Item.ItemName.FormatMsvcString(m.reader.Ctx)
		if itemName != "" {
			rows := []string{row("Item Name", itemName)}
			if d.Item.UsesRemaining != -1 {
				rows = append(rows, row("Uses", fmt.Sprintf("%d", d.Item.UsesRemaining)))
			}
			sections = append(sections, renderSection("Item", rows))
		}
	}

	if d.Sprite != nil {
		imgFile := d.Sprite.ImageFile.FormatMsvcString(m.reader.Ctx)
		if imgFile != "" {
			// Shorten long paths: strip common prefix.
			imgFile = strings.TrimPrefix(imgFile, "data/")
			sections = append(sections, renderSection("Sprite", []string{
				row("Image", dimStyle.Render(imgFile)),
			}))
		}
	}

	if d.Velocity != nil {
		v := d.Velocity
		rows := []string{
			row("Gravity", fmt.Sprintf("%.0f, %.0f", v.GravityX, v.GravityY)),
			row("Mass", fmt.Sprintf("%.3f", v.Mass)),
		}
		if v.AirFriction != 0.55 { // only show if non-default
			rows = append(rows, row("Air Friction", fmt.Sprintf("%.2f", v.AirFriction)))
		}
		sections = append(sections, renderSection("Physics", rows))
	}

	if d.Light != nil && d.Light.Radius > 0 {
		sections = append(sections, renderSection("Light", []string{
			row("Radius", fmt.Sprintf("%.0f", d.Light.Radius)),
			row("Color", fmt.Sprintf("rgb(%d, %d, %d)", d.Light.R, d.Light.G, d.Light.B)),
		}))
	}

	if d.Effect != nil {
		rows := []string{
			row("Effect", fmt.Sprintf("%d", d.Effect.Effect)),
		}
		if d.Effect.Frames >= 0 {
			rows = append(rows, row("Frames Left", fmt.Sprintf("%d", d.Effect.Frames)))
		} else {
			rows = append(rows, row("Duration", "permanent"))
		}
		sections = append(sections, renderSection("Effect", rows))
	}

	if len(d.Children) > 0 {
		var rows []string
		for _, child := range d.Children {
			name := child.Name
			if name == "" {
				name = fmt.Sprintf("entity_%d", child.Entity.EntityId)
			}
			rows = append(rows, row(fmt.Sprintf("#%d", child.Entity.EntityId), name))
		}
		sections = append(sections, renderSection(fmt.Sprintf("Children (%d)", len(d.Children)), rows))
	}

	if len(sections) == 0 {
		return dimStyle.Render("No data available")
	}
	return strings.Join(sections, "")
}

// ── Wands tab ──────────────────────────────────────────────────────

func (m model) viewWands() string {
	if len(m.state.Wands) == 0 && len(m.state.Items) == 0 {
		return dimStyle.Render("No inventory items found")
	}

	const numSlots = 4
	wlabel := lipgloss.NewStyle().Foreground(lipgloss.Color("#8BE9FD")).Width(15)
	wrow := func(label, value string) string {
		return wlabel.Render(label) + valueStyle.Render(value)
	}

	var wandCols []string
	for i := 0; i < numSlots; i++ {
		if i < len(m.state.Wands) {
			item := m.state.Wands[i]
			w := item.Ability
			name := item.Name(m.reader.Ctx)
			if name == "" {
				name = fmt.Sprintf("Wand %d", i+1)
			}
			gc := w.GunConfig
			rows := []string{
				wrow("Spells/Cast", fmt.Sprintf("%d", gc.ActionsPerRound)),
				wrow("Deck Capacity", fmt.Sprintf("%d", gc.DeckCapacity)),
				wrow("Shuffle", boolStr(gc.ShuffleDeckWhenEmpty)),
				wrow("Mana", manaStyle.Render(fmt.Sprintf("%.0f / %.0f", w.Mana, w.ManaMax))),
				wrow("Mana Regen", fmt.Sprintf("%.0f/s", w.ManaChargeSpeed*60)),
				wrow("Reload", fmt.Sprintf("%.2fs (%df)", float64(gc.ReloadTime)/60.0, gc.ReloadTime)),
			}
			wandCols = append(wandCols, renderSection(name, rows))
		} else {
			wandCols = append(wandCols, renderSection(fmt.Sprintf("Wand %d", i+1), []string{dimStyle.Render("empty")}))
		}
	}
	wandCols = equalizeBoxes(wandCols)

	var itemCols []string
	for i := 0; i < numSlots; i++ {
		if i < len(m.state.Items) {
			item := m.state.Items[i]
			var rows []string
			for _, mat := range item.Contents {
				rows = append(rows, wrow(mat.Name, fmt.Sprintf("%.0f", mat.Amount)))
			}
			if len(rows) == 0 {
				rows = append(rows, dimStyle.Render("empty"))
			}
			name := item.Name(m.reader.Ctx)
			itemCols = append(itemCols, renderSection(name, rows))
		} else {
			itemCols = append(itemCols, renderSection(fmt.Sprintf("Item %d", i+1), []string{dimStyle.Render("empty")}))
		}
	}
	itemCols = equalizeBoxes(itemCols)

	return lipgloss.JoinHorizontal(lipgloss.Top, wandCols...) + "\n" +
		lipgloss.JoinHorizontal(lipgloss.Top, itemCols...)
}

// ── World tab ──────────────────────────────────────────────────────

func (m model) viewWorld() string {
	var sections []string

	{
		rows := []string{
			row("World Seed", fmt.Sprintf("%d", m.state.WorldSeed)),
			row("Death Count", fmt.Sprintf("%d", m.state.DeathCount)),
			row("Orbs Total", fmt.Sprintf("%d", m.state.NumOrbsTotal)),
		}
		if g := m.state.Globals; g != nil {
			rows = append(rows,
				row("Frame", fmt.Sprintf("%d", g.FrameCount)),
				row("Game Time", fmt.Sprintf("%.1f", g.GameTime)),
				row("Camera", posStyle.Render(fmt.Sprintf("%.1f, %.1f", m.state.CameraX, m.state.CameraY))),
				row("View Size", fmt.Sprintf("%.0f x %.0f", m.state.ViewW, m.state.ViewH)),
			)
		}
		sections = append(sections, renderSection("World", rows))
	}

	if ws := m.state.WorldState; ws != nil {
		rows := []string{
			row("Gods Afraid", fmt.Sprintf("%d", ws.GodsAfraid)),
			row("Gods Impressed", fmt.Sprintf("%d", ws.GodsImpressed)),
			row("Gods Enraged", fmt.Sprintf("%d", ws.GodsEnraged)),
		}
		sections = append(sections, renderSection("Gods", rows))
	}

	return strings.Join(sections, "")
}

// ── Overlay tab ────────────────────────────────────────────────────

func (m model) viewOverlay() string {
	checkOn := lipgloss.NewStyle().Foreground(lipgloss.Color("#50FA7B")).Render("[x]")
	checkOff := dimStyle.Render("[ ]")
	cursor := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF79C6")).Render(">")
	noCursor := "  "

	var rows []string
	idx := 0

	// Categories section
	rows = append(rows, sectionTitleStyle.Render("Categories")+" "+dimStyle.Render("(space to toggle)"))
	for _, cat := range overlayCategories {
		check := checkOff
		if m.overlayCats[cat] {
			check = checkOn
		}
		prefix := noCursor
		if m.overlayCursor == idx {
			prefix = cursor
		}
		catStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(cat.color()))
		n := m.categoryCounts[cat]
		count := dimStyle.Render(fmt.Sprintf("(%d)", n))
		rows = append(rows, fmt.Sprintf(" %s %s %s %s", prefix, check, catStyle.Render(cat.String()), count))
		idx++
	}

	rows = append(rows, "")

	// Options section
	rows = append(rows, sectionTitleStyle.Render("Options"))
	for opt := overlayOption(0); opt < optionCount; opt++ {
		check := checkOff
		if m.overlayOpts[opt] {
			check = checkOn
		}
		prefix := noCursor
		if m.overlayCursor == idx {
			prefix = cursor
		}
		rows = append(rows, fmt.Sprintf(" %s %s %s", prefix, check, opt.String()))
		idx++
	}

	// Count of entities being rendered
	var renderCount int
	if ents := m.overlay.entities.Load(); ents != nil {
		renderCount = len(*ents)
	}
	rows = append(rows, "")
	rows = append(rows, dimStyle.Render(fmt.Sprintf("Rendering %d entities on overlay", renderCount)))

	content := strings.Join(rows, "\n")
	return sectionStyle.Render(content)
}

// ── Log tab ───────────────────────────────────────────────────────

func (m model) viewLog() string {
	lines := m.logBuf.Lines()
	if len(lines) == 0 {
		return dimStyle.Render("(no log messages)")
	}
	var b strings.Builder
	for _, l := range lines {
		b.WriteString(dimStyle.Render(l) + "\n")
	}
	return b.String()
}

// ── Helpers ────────────────────────────────────────────────────────

func truncateAnsi(s string, maxWidth int) string {
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	runes := []rune(s)
	for len(runes) > 0 && lipgloss.Width(string(runes)) > maxWidth-1 {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "…"
}

func row(label, value string) string {
	return labelStyle.Render(label) + valueStyle.Render(value)
}

func renderSection(title string, rows []string) string {
	content := sectionTitleStyle.Render(title) + "\n" + strings.Join(rows, "\n")
	return sectionStyle.Render(content) + "\n"
}

func equalizeBoxes(boxes []string) []string {
	maxW, maxH := 0, 0
	for _, b := range boxes {
		if w := lipgloss.Width(b); w > maxW {
			maxW = w
		}
		if h := lipgloss.Height(b); h > maxH {
			maxH = h
		}
	}
	for i, b := range boxes {
		boxes[i] = lipgloss.Place(maxW, maxH, lipgloss.Left, lipgloss.Top, b)
	}
	return boxes
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func shuffleStr(b bool) string {
	if b {
		return "shuffle"
	}
	return "no-shuffle"
}

func dmgMultsNonDefault(d *noita.DamageModelComponent) []string {
	mults := []struct {
		name string
		val  float32
	}{
		{"Melee", d.DmgMultMelee}, {"Proj", d.DmgMultProjectile},
		{"Expl", d.DmgMultExplosion}, {"Elec", d.DmgMultElectricity},
		{"Fire", d.DmgMultFire}, {"Drill", d.DmgMultDrill},
		{"Slice", d.DmgMultSlice}, {"Ice", d.DmgMultIce},
		{"Heal", d.DmgMultHealing}, {"Phys", d.DmgMultPhysicsHit},
		{"Rad", d.DmgMultRadioactive}, {"Poison", d.DmgMultPoison},
		{"Holy", d.DmgMultHoly}, {"Curse", d.DmgMultCurse},
	}
	var out []string
	for _, m := range mults {
		if m.val != 1.0 {
			out = append(out, fmt.Sprintf("%s:%.2f", m.name, m.val))
		}
	}
	return out
}

func main() {
	logBuf := newRingLog(20)
	log.SetOutput(logBuf)
	log.SetFlags(log.Ltime)

	p := tea.NewProgram(initialModel(logBuf), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
