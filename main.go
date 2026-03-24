package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"sort"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/vitaminmoo/memtools/process"
	"noitrainer/noita"
)

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

func categorize(e *noita.EntitySummary) entityCategory {
	has := make(map[int]bool)
	for _, id := range e.ComponentIDs {
		has[id] = true
	}

	if strings.Contains(e.Name, "player") || e.Name == "arm_r" || e.Name == "cape" ||
		strings.HasPrefix(e.Name, "inventory_") || e.Name == "player_stats" {
		return catPlayer
	}
	if has[noita.TypeAnimalAIComponent] || strings.HasPrefix(e.Name, "$animal_") {
		return catEnemy
	}
	if has[noita.TypeTorchComponent] {
		return catTorch
	}
	if has[noita.TypeItemComponent] || has[noita.TypeAbilityComponent] {
		return catItem
	}
	if has[noita.TypeGameEffectComponent] {
		return catEffect
	}
	if has[noita.TypeVerletPhysicsComponent] && !has[noita.TypeDamageModelComponent] {
		return catProp
	}
	if has[noita.TypeSimplePhysicsComponent] || has[noita.TypePixelSpriteComponent] {
		return catPhysics
	}
	return catOther
}

func subcategorize(e *noita.EntitySummary) string {
	has := make(map[int]bool)
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
		case has[noita.TypePotionComponent]:
			return "potion"
		case has[noita.TypeAbilityComponent] && has[noita.TypeManaReloaderComponent]:
			return "wand"
		case has[noita.TypeItemActionComponent]:
			return "spell"
		case has[noita.TypeAbilityComponent]:
			return "holdable"
		default:
			return "pickup"
		}
	case catTorch:
		if has[noita.TypePhysicsBody2Component] {
			return "physics"
		}
		return "static"
	case catPhysics:
		if has[noita.TypeLuaComponent] {
			return "scripted"
		}
		return "debris"
	case catProp:
		if has[noita.TypeVerletWorldJointComponent] {
			return "hanging"
		}
		return "loose"
	case catEffect:
		if has[noita.TypeInheritTransformComponent] {
			return "attached"
		}
		return "standalone"
	default:
		if has[noita.TypeCollisionTriggerComponent] {
			return "trigger"
		}
		if has[noita.TypeVariableStorageComponent] {
			return "variable"
		}
		if has[noita.TypeWorldStateComponent] {
			return "world"
		}
		if has[noita.TypeCameraBoundComponent] {
			return "camera"
		}
		return "misc"
	}
}

func entityDisplayName(e *noita.EntitySummary) string {
	name := e.Name
	if name == "" || name == "unknown" {
		return fmt.Sprintf("#%d", e.Entity.EntityId)
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
	catStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(i.category.color()))
	subStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(i.category.color())).Faint(true)
	return fmt.Sprintf("%s/%s %s",
		catStyle.Render(i.category.String()),
		subStyle.Render(i.subcategory),
		entityDisplayName(i.summary))
}

func (i entityItem) Description() string {
	e := i.summary
	parts := []string{fmt.Sprintf("(%.0f, %.0f)", e.Entity.PosX, e.Entity.PosY)}
	if e.HasHP {
		parts = append(parts, "HP")
	}
	return strings.Join(parts, "  ")
}

func (i entityItem) FilterValue() string {
	return i.category.String() + " " + i.subcategory + " " + entityDisplayName(i.summary)
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
}

func newEntityList() list.Model {
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = delegate.Styles.SelectedTitle.
		Foreground(lipgloss.Color("#FF79C6")).
		BorderLeftForeground(lipgloss.Color("#FF79C6"))
	delegate.Styles.SelectedDesc = delegate.Styles.SelectedDesc.
		Foreground(lipgloss.Color("#6272A4")).
		BorderLeftForeground(lipgloss.Color("#FF79C6"))

	l := list.New(nil, delegate, 40, 20)
	l.Title = "Entities"
	l.Styles.Title = sectionTitleStyle
	l.SetShowStatusBar(true)
	l.SetShowHelp(false)
	l.DisableQuitKeybindings()
	return l
}

func initialModel() model {
	return model{
		state:          &noita.GameState{},
		tabs:           []string{"Player", "Entities", "Wands", "World"},
		entityList:     newEntityList(),
		categoryCounts: make(map[entityCategory]int),
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
		}
		return m, tickCmd()
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

	if m.entityList.FilterState() != list.Filtering {
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

// ── Helpers ────────────────────────────────────────────────────────

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
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
