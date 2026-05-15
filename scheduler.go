package main

import (
	"time"

	"noitrainer/noita"
)

// Tab indices, parallel to initialModel().tabs.
const (
	tabPlayer = iota
	tabEntities
	tabWands
	tabWorld
	tabOverlay
	tabLog
)

// domainScheduler decides which game-state domains to refresh on each tick.
// It combines two ingredients: the wayland overlay's always-on needs and any
// extras required by the active TUI tab, gated by per-domain refresh intervals
// so static or rarely-changing data isn't reread every 100ms.
type domainScheduler struct {
	interval map[noita.Domain]time.Duration
}

// newDomainScheduler returns a scheduler with sensible cadences. An interval
// of 0 means "run every time the domain is requested".
func newDomainScheduler() *domainScheduler {
	return &domainScheduler{
		interval: map[noita.Domain]time.Duration{
			noita.DomainStatics:          5 * time.Second,
			noita.DomainGlobalsAndCamera: 0,
			noita.DomainWorldState:       1 * time.Second,
			noita.DomainPlayerCore:       0,
			noita.DomainPlayerInventory:  250 * time.Millisecond,
			noita.DomainPlayerEffects:    200 * time.Millisecond,
			noita.DomainEntities:         0,
		},
	}
}

// overlayBase is the set of domains the wayland overlay always needs to render:
// camera transform, player position, and the entity list. Updated every tick
// while the overlay is alive, regardless of which TUI tab is active.
func (s *domainScheduler) overlayBase() []noita.Domain {
	return []noita.Domain{
		noita.DomainGlobalsAndCamera,
		noita.DomainPlayerCore,
		noita.DomainEntities,
	}
}

// tabExtras returns domains added by the active TUI tab beyond overlayBase.
func (s *domainScheduler) tabExtras(tab int) []noita.Domain {
	switch tab {
	case tabPlayer:
		return []noita.Domain{noita.DomainStatics, noita.DomainPlayerEffects}
	case tabEntities:
		return []noita.Domain{noita.DomainStatics}
	case tabWands:
		return []noita.Domain{noita.DomainStatics, noita.DomainPlayerInventory}
	case tabWorld:
		return []noita.Domain{noita.DomainStatics, noita.DomainWorldState}
	case tabOverlay, tabLog:
		return nil
	}
	return nil
}

// pick returns the subset of required domains whose last-read timestamp is at
// least their refresh interval old (or that have never been read).
func (s *domainScheduler) pick(tab int, state *noita.GameState, now time.Time) []noita.Domain {
	required := s.overlayBase()
	required = append(required, s.tabExtras(tab)...)

	seen := make(map[noita.Domain]bool, len(required))
	out := make([]noita.Domain, 0, len(required))
	for _, d := range required {
		if seen[d] {
			continue
		}
		seen[d] = true
		if state == nil || state.Domains == nil {
			out = append(out, d)
			continue
		}
		last, ok := state.Domains[d]
		if !ok {
			out = append(out, d)
			continue
		}
		if now.Sub(last) >= s.interval[d] {
			out = append(out, d)
		}
	}
	return out
}
