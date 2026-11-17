package main

import "charm.land/lipgloss/v2"

var (
	primary = lipgloss.Color("#67E8F9")
	accent  = lipgloss.Color("#A78BFA")
	muted   = lipgloss.Color("#94A3B8")
	success = lipgloss.Color("#4ADE80")
	warning = lipgloss.Color("#FBBF24")
	white   = lipgloss.Color("#F8FAFC")
 danger = lipgloss.Color("#F87171")
 background = lipgloss.Color("#0F172A")

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			MarginBottom(1)

	subtitleStyle = lipgloss.NewStyle().
			Foreground(muted)

	selectedStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primary).
			PaddingLeft(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(white).
			PaddingLeft(3)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#334155")).
 Background(background).
 Foreground(white).
			Padding(1, 2)

	labelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(accent)

	valueStyle = lipgloss.NewStyle().
			Foreground(white).
			PaddingLeft(2)

	helpStyle = lipgloss.NewStyle().
			Foreground(muted).
			MarginTop(1)

	errorStyle = lipgloss.NewStyle().
			Foreground(danger)
 warningStyle = lipgloss.NewStyle().Foreground(warning)

	successStyle = lipgloss.NewStyle().
			Foreground(success)
)
