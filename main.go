package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/key"
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
)

type tickMsg time.Time

type model struct {
	state    *noita.GameState
	reader   *noita.Reader
	proc     *process.Process
	tab      int
	tabs     []string
	width    int
	height   int
	err      error
	quitting bool
}

func initialModel() model {
	return model{
		state: &noita.GameState{},
		tabs:  []string{"Player", "Wands", "World", "ECS"},
	}
}

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

func (m model) Init() tea.Cmd {
	return tickCmd()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, key.NewBinding(key.WithKeys("q", "ctrl+c"))):
			m.quitting = true
			return m, tea.Quit
		case key.Matches(msg, key.NewBinding(key.WithKeys("tab", "right", "l"))):
			m.tab = (m.tab + 1) % len(m.tabs)
		case key.Matches(msg, key.NewBinding(key.WithKeys("shift+tab", "left", "h"))):
			m.tab = (m.tab - 1 + len(m.tabs)) % len(m.tabs)
		}

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height

	case tickMsg:
		m.tryConnect()
		if m.reader != nil {
			m.state = m.reader.ReadState()
		}
		return m, tickCmd()
	}

	return m, nil
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

func (m model) View() string {
	if m.quitting {
		return ""
	}

	var b strings.Builder

	// Title bar
	title := titleStyle.Render(" NOITRAINER ")
	if m.state.Connected {
		title += dimStyle.Render("  connected")
	} else {
		title += errorStyle.Render("  disconnected")
	}
	b.WriteString(title + "\n\n")

	// Tabs
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

	// Content based on active tab
	switch m.tab {
	case 0:
		b.WriteString(m.viewPlayer())
	case 1:
		b.WriteString(m.viewWands())
	case 2:
		b.WriteString(m.viewWorld())
	case 3:
		b.WriteString(m.viewECS())
	}

	b.WriteString("\n" + dimStyle.Render("tab/shift+tab: switch tabs  q: quit"))

	return b.String()
}

func (m model) viewPlayer() string {
	var sections []string

	// Position & Movement
	{
		var rows []string
		if e := m.state.PlayerEntity; e != nil {
			rows = append(rows, row("Position", posStyle.Render(fmt.Sprintf("%.1f, %.1f", e.PosX, e.PosY))))
			rows = append(rows, row("Entity ID", fmt.Sprintf("%d", e.EntityId)))
			name := noita.MsvcStringValue(&e.Name, m.reader.Ctx)
			if name != "" {
				rows = append(rows, row("Name", name))
			}
		}
		if c := m.state.PlayerChar; c != nil {
			rows = append(rows, row("Velocity", fmt.Sprintf("%.1f, %.1f", c.VelocityX, c.VelocityY)))
			rows = append(rows, row("On Ground", boolStr(c.IsOnGround)))
			rows = append(rows, row("Gravity", fmt.Sprintf("%.2f", c.Gravity)))
			rows = append(rows, row("Fly Time", fmt.Sprintf("%.1f", c.FlyTime)))
		}
		if len(rows) > 0 {
			sections = append(sections, renderSection("Movement", rows))
		}
	}

	// Health
	{
		var rows []string
		if d := m.state.PlayerHP; d != nil {
			hp := d.Hp * 25
			maxHp := d.MaxHp * 25
			rows = append(rows, row("HP", hpStyle.Render(fmt.Sprintf("%.0f / %.0f", hp, maxHp))))
			rows = append(rows, row("Max HP Cap", fmt.Sprintf("%.0f", d.MaxHpCap*25)))
			rows = append(rows, row("I-Frames", fmt.Sprintf("%d", d.InvincibilityFrames)))

			// Damage multipliers that differ from 1.0
			mults := []struct {
				name string
				val  float32
			}{
				{"Melee", d.DmgMultMelee}, {"Projectile", d.DmgMultProjectile},
				{"Explosion", d.DmgMultExplosion}, {"Electricity", d.DmgMultElectricity},
				{"Fire", d.DmgMultFire}, {"Drill", d.DmgMultDrill},
				{"Slice", d.DmgMultSlice}, {"Ice", d.DmgMultIce},
				{"Healing", d.DmgMultHealing}, {"Physics", d.DmgMultPhysicsHit},
				{"Radioactive", d.DmgMultRadioactive}, {"Poison", d.DmgMultPoison},
				{"Holy", d.DmgMultHoly}, {"Curse", d.DmgMultCurse},
				{"Overeating", d.DmgMultOvereating}, {"Material", d.DmgMultMaterial},
			}
			var nonDefault []string
			for _, mult := range mults {
				if mult.val != 1.0 {
					nonDefault = append(nonDefault, fmt.Sprintf("%s:%.2f", mult.name, mult.val))
				}
			}
			if len(nonDefault) > 0 {
				rows = append(rows, row("Dmg Mults", strings.Join(nonDefault, " ")))
			}
		}
		if len(rows) > 0 {
			sections = append(sections, renderSection("Health", rows))
		}
	}

	// Wallet & Inventory
	{
		var rows []string
		if w := m.state.PlayerWallet; w != nil {
			rows = append(rows, row("Gold", goldStyle.Render(fmt.Sprintf("%d", w.Money))))
			rows = append(rows, row("Gold Spent", fmt.Sprintf("%d", w.MoneySpent)))
		}
		if inv := m.state.PlayerInv; inv != nil {
			rows = append(rows, row("Active Item ID", fmt.Sprintf("%d", inv.ActiveItem)))
			rows = append(rows, row("Wand Slots", fmt.Sprintf("%d", inv.QuickInventorySlots)))
		}
		if len(rows) > 0 {
			sections = append(sections, renderSection("Inventory", rows))
		}
	}

	if len(sections) == 0 {
		return dimStyle.Render("No player data available")
	}
	return strings.Join(sections, "")
}

func (m model) viewWands() string {
	if len(m.state.Wands) == 0 && len(m.state.Items) == 0 {
		return dimStyle.Render("No inventory items found")
	}

	var sections []string

	// Wands
	for i, item := range m.state.Wands {
		w := item.Ability
		var rows []string
		name := item.Name(m.reader.Ctx)
		if name == "" {
			name = fmt.Sprintf("Wand %d", i+1)
		}

		gc := w.GunConfig
		rows = append(rows, row("Spells/Cast", fmt.Sprintf("%d", gc.ActionsPerRound)))
		rows = append(rows, row("Deck Capacity", fmt.Sprintf("%d", gc.DeckCapacity)))
		rows = append(rows, row("Shuffle", boolStr(gc.ShuffleDeckWhenEmpty)))
		rows = append(rows, row("Mana", manaStyle.Render(fmt.Sprintf("%.0f / %.0f", w.Mana, w.ManaMax))))
		rows = append(rows, row("Mana Regen", fmt.Sprintf("%.0f/s", w.ManaChargeSpeed*60)))
		rows = append(rows, row("Reload Time", fmt.Sprintf("%.2fs (%d frames)", float64(gc.ReloadTime)/60.0, gc.ReloadTime)))

		if w.ReloadFramesLeft > 0 {
			rows = append(rows, row("Reloading", fmt.Sprintf("%d frames left", w.ReloadFramesLeft)))
		}

		sections = append(sections, renderSection(name, rows))
	}

	// Items (potions, powder pouches, etc.)
	for _, item := range m.state.Items {
		var rows []string
		for _, mat := range item.Contents {
			rows = append(rows, row(mat.Name, fmt.Sprintf("%.0f cells", mat.Amount)))
		}
		if len(rows) == 0 {
			rows = append(rows, dimStyle.Render("empty"))
		}
		name := item.Name(m.reader.Ctx)
		sections = append(sections, renderSection(name, rows))
	}

	return strings.Join(sections, "")
}

func (m model) viewWorld() string {
	var sections []string

	// Global info
	{
		var rows []string
		rows = append(rows, row("World Seed", fmt.Sprintf("%d", m.state.WorldSeed)))
		rows = append(rows, row("Death Count", fmt.Sprintf("%d", m.state.DeathCount)))
		rows = append(rows, row("Orbs Total", fmt.Sprintf("%d", m.state.NumOrbsTotal)))

		if g := m.state.Globals; g != nil {
			rows = append(rows, row("Frame", fmt.Sprintf("%d", g.FrameCount)))
			rows = append(rows, row("Game Time", fmt.Sprintf("%.1f", g.GameTime)))
			rows = append(rows, row("Camera", posStyle.Render(fmt.Sprintf("%.1f, %.1f", m.state.CameraX, m.state.CameraY))))
			rows = append(rows, row("View Size", fmt.Sprintf("%.0f x %.0f", m.state.ViewW, m.state.ViewH)))
		}
		sections = append(sections, renderSection("World", rows))
	}

	// Gods
	if ws := m.state.WorldState; ws != nil {
		var rows []string
		rows = append(rows, row("Gods Afraid", fmt.Sprintf("%d", ws.GodsAfraid)))
		rows = append(rows, row("Gods Impressed", fmt.Sprintf("%d", ws.GodsImpressed)))
		rows = append(rows, row("Gods Enraged", fmt.Sprintf("%d", ws.GodsEnraged)))
		rows = append(rows, row("Gods Afraid Dmg", fmt.Sprintf("%d", ws.GodsAfraidDamage)))
		rows = append(rows, row("Biome Crypt", fmt.Sprintf("%d", ws.BiomeCryptCount)))
		sections = append(sections, renderSection("Gods", rows))
	}

	return strings.Join(sections, "")
}

func (m model) viewECS() string {
	var sections []string

	if g := m.state.Globals; g != nil {
		var rows []string
		rows = append(rows, row("WorldManager", fmt.Sprintf("0x%08X", g.PWorldManager)))
		rows = append(rows, row("ChunkSystem", fmt.Sprintf("0x%08X", g.PChunkSystem)))
		rows = append(rows, row("CellGrid", fmt.Sprintf("0x%08X", g.PCellGrid)))
		rows = append(rows, row("CellFactory", fmt.Sprintf("0x%08X", g.PCellFactory)))
		rows = append(rows, row("PhysicsWorld", fmt.Sprintf("0x%08X", g.PPhysicsWorld)))
		rows = append(rows, row("AudioManager", fmt.Sprintf("0x%08X", g.PAudioManager)))
		sections = append(sections, renderSection("Manager Pointers", rows))
	}

	if e := m.state.PlayerEntity; e != nil {
		var rows []string
		rows = append(rows, row("Slot Index", fmt.Sprintf("%d", e.SlotIndex)))
		rows = append(rows, row("Pending Kill", fmt.Sprintf("%d", e.PendingKill)))
		rows = append(rows, row("Flags", fmt.Sprintf("0x%08X", e.Flags10)))
		rows = append(rows, row("Children Ptr", fmt.Sprintf("0x%08X", e.ChildrenPtr)))
		rows = append(rows, row("Parent Ptr", fmt.Sprintf("0x%08X", e.ParentEntityPtr)))
		rows = append(rows, row("Scale", fmt.Sprintf("%.2f, %.2f", e.ScaleX, e.ScaleY)))
		rows = append(rows, row("Rotation", fmt.Sprintf("cos=%.3f sin=%.3f", e.RotCos, e.RotSin)))
		sections = append(sections, renderSection("Player Entity", rows))
	}

	if len(sections) == 0 {
		return dimStyle.Render("No ECS data available")
	}
	return strings.Join(sections, "")
}

func row(label, value string) string {
	return labelStyle.Render(label) + valueStyle.Render(value)
}

func renderSection(title string, rows []string) string {
	content := sectionTitleStyle.Render(title) + "\n" + strings.Join(rows, "\n")
	return sectionStyle.Render(content) + "\n"
}

func boolStr(b bool) string {
	if b {
		return "yes"
	}
	return "no"
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
