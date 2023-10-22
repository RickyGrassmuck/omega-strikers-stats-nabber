package tracker

import (
	"time"
)

func NewMatch() *MatchState {
	match := &MatchState{
		Phase:      GamePhase("PreGame"),
		Results:    MatchResults{},
		CurrentSet: 0,
	}
	match.Results.Sets = make(map[int]SetResult)
	return match
}

func (m *MatchState) MatchStarted() {
	m.Start = time.Now()
}

func (m *MatchState) MatchEnded() {
	m.End = time.Now()
}

func (m *MatchState) MatchDuration() time.Duration {
	return m.End.Sub(m.Start)
}

func (m *MatchState) UpdateMatchWinner(winningTeam string) {
	m.Results.Winner = winningTeam
}

func (m *MatchState) StartNewSet() {
	m.CurrentSet++
	m.Results.Sets[m.CurrentSet] = SetResult{}
}

func (m *MatchState) UpdateSetScore(scoringTeam string) {
	set := m.Results.Sets[m.CurrentSet]
	if scoringTeam == "TeamOne" {
		set.TeamOneScore++
	} else if scoringTeam == "TeamTwo" {
		set.TeamTwoScore++
	}
	m.Results.Sets[m.CurrentSet] = set
}

func (m *MatchState) SetWinner(winningTeam string) {
	set := m.Results.Sets[m.CurrentSet]
	set.Winner = winningTeam
	m.Results.Sets[m.CurrentSet] = set
}
