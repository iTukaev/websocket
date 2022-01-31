package provider

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

// Match - structure provider data response
type Match struct {
	Finale           bool      `json:"finale,omitempty"`
	GameId           int       `json:"game_id,omitempty"`
	GameStart        int       `json:"game_start,omitempty"`
	GameOcList       [2]GameOc `json:"game_oc_list,omitempty"`
	Opp1Icon         int       `json:"opp_1_icon,omitempty"`
	Opp1Name         string    `json:"opp_1_name,omitempty"`
	Opp2Icon         int       `json:"opp_2_icon,omitempty"`
	Opp2Name         string    `json:"opp_2_name,omitempty"`
	PeriodName       string    `json:"period_name,omitempty"`
	ScoreFull        string    `json:"score_full,omitempty"`
	ScorePeriod      string    `json:"score_period,omitempty"`
	TournamentID     int       `json:"tournament_id,omitempty"`
	TournamentNameRU string    `json:"tournament_name_ru,omitempty"`
}

// GameOc - structure command info
type GameOc struct {
	OcGroupName string  `json:"oc_group_name,omitempty"`
	OcName      string  `json:"oc_name,omitempty"`
	OcRate      float32 `json:"oc_rate,omitempty"`
	OcPointer   string  `json:"oc_pointer,omitempty"`
	OcBlock     bool    `json:"oc_block,omitempty"`
}

// Matches include matches info and type matches parameter,
// see documentation for possible values
type Matches struct {
	Parameter string  `json:"parameter"`
	Body      []Match `json:"body"`
}

func NewMatches(parameter string) *Matches {
	return &Matches{
		Parameter: parameter,
	}
}

//Equal compares two structures by body
func (m *Matches) Equal(b *Matches) bool {
	if len(m.Body) != len(b.Body) {
		return false
	}
	for i, v := range m.Body {
		if v != b.Body[i] {
			return false
		}
	}
	return true
}

//Sort selects matches CS:GO and sorts them by start time
func (m *Matches) Sort() {
	for i := len(m.Body) - 1; i >= 0; i-- {
		if !strings.Contains(m.Body[i].TournamentNameRU, "CS:GO") {
			m.Body[i] = m.Body[len(m.Body)-1]
			m.Body = m.Body[:len(m.Body)-1]
		}
	}

	sort.Slice(m.Body, func(i, j int) bool {
		return m.Body[i].GameStart < m.Body[j].GameStart
	})
}

//Bytes return byte slice of Matches
func (m *Matches) Bytes() ([]byte, error) {
	b, err := json.Marshal(m)
	if err != nil {
		return nil, fmt.Errorf("marshaling error: %w", err)
	}
	return b, nil
}
