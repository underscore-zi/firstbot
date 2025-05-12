package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sync"
)

type StateFile struct {
	IsLive      bool           `json:"is_live"`
	IsClaimed   bool           `json:"is_claimed"`
	ClaimedBy   string         `json:"claimed_by"`
	TotalClaims map[string]int `json:"total_claims"`
	Streak      int            `json:"streak"`
	LateClaims  []string       `json:"late_claims"`
}

type State struct {
	Lock     sync.Mutex `json:"-"`
	Filename string     `json:"-"`
	raw      StateFile
}

func (s *State) Save() error {
	jsonData, err := json.Marshal(s.raw)
	if err != nil {
		fmt.Printf("error marshalling state: %v\n", err)
		return err
	}

	err = os.WriteFile(s.Filename, jsonData, 0644)
	if err != nil {
		fmt.Printf("error writing state: %v\n", err)
		return err
	}

	return nil
}

func (s *State) Load() error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	jsonData, err := os.ReadFile(s.Filename)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	err = json.Unmarshal(jsonData, &s.raw)
	if err != nil {
		return err
	}

	return nil
}

func (s *State) IsLive() bool {
	return s.raw.IsLive
}

func (s *State) SetOnline() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	defer func() { _ = s.Save() }()

	if !s.raw.IsLive {
		s.raw.IsLive = true
		s.raw.IsClaimed = false
		s.raw.LateClaims = nil
	}
}

func (s *State) SetOffline() {
	s.Lock.Lock()
	defer s.Lock.Unlock()
	defer func() { _ = s.Save() }()

	s.raw.IsLive = false
}

var ErrAlreadyClaimed = errors.New("already claimed")
var ErrNotLive = errors.New("streamer is not live")
var ErrIncorrectOrdinal = errors.New("streamer is not live")

func (s *State) TryClaim(username string) error {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	if !s.raw.IsLive {
		return ErrNotLive
	}
	if s.raw.IsClaimed {
		return ErrAlreadyClaimed
	}

	s.raw.IsClaimed = true

	// Check if this is adding onto the existing streak
	if username == s.raw.ClaimedBy {
		s.raw.Streak++
	} else {
		s.raw.Streak = 1
	}

	// Update the claimed By Username, needs to happen AFTER streak check
	s.raw.ClaimedBy = username

	// Update the total claims
	if s.raw.TotalClaims == nil {
		s.raw.TotalClaims = make(map[string]int)
	}
	s.raw.TotalClaims[username]++

	_ = s.Save()
	return nil
}

func (s *State) TryLateClaim(username, ordinal string) (error, int) {
	s.Lock.Lock()
	defer s.Lock.Unlock()

	// Can't do a late claim if first isn't claimed
	if !s.raw.IsClaimed {
		return ErrNotLive, 0
	}

	expected := s.ExpectedOrdinal(len(s.raw.LateClaims) + 2)
	if ordinal != expected {
		return ErrIncorrectOrdinal, 0
	}

	// Check if LateClaims contains the username already
	if username == s.raw.ClaimedBy {
		return ErrAlreadyClaimed, 1
	}
	for i, claim := range s.raw.LateClaims {
		if claim == username {
			return ErrAlreadyClaimed, i + 2
		}
	}

	s.raw.LateClaims = append(s.raw.LateClaims, username)
	_ = s.Save()

	return nil, len(s.raw.LateClaims) + 2
}

func (s *State) ExpectedOrdinal(position int) string {
	if position <= 0 {
		return ""
	}

	// Predefined ordinals for common numbers
	ordinals := map[int]string{
		1:  "first",
		2:  "second",
		3:  "third",
		4:  "fourth",
		5:  "fifth",
		6:  "sixth",
		7:  "seventh",
		8:  "eighth",
		9:  "ninth",
		10: "tenth",
		11: "eleventh",
		12: "twelfth",
		13: "thirteenth",
		14: "fourteenth",
		15: "fifteenth",
		16: "sixteenth",
		17: "seventeenth",
		18: "eighteenth",
		19: "nineteenth",
	}

	if word, exists := ordinals[position]; exists {
		return word
	}

	// Handle numbers beyond 20
	tensExact := []string{"", "", "twentieth", "thirtieth", "fortieth", "fiftieth", "sixtieth", "seventieth", "eightieth", "ninetieth"}
	tens := []string{"", "", "twenty", "thirty", "forty", "fifty", "sixty", "seventy", "eighty", "ninety"}

	if position < 100 {
		tensPart := position / 10
		onesPart := position % 10
		if onesPart == 0 {
			return tensExact[tensPart]
		}
		return tens[tensPart] + "-" + ordinals[onesPart]
	}

	return fmt.Sprintf("%dth", position)
}

func (s *State) ClaimedBy() (string, int, int) {
	return s.raw.ClaimedBy, s.raw.TotalClaims[s.raw.ClaimedBy], s.raw.Streak
}

func (s *State) AllClaims() []string {
	if s.raw.ClaimedBy == "" {
		return nil
	}
	return append([]string{s.raw.ClaimedBy}, s.raw.LateClaims...)
}
