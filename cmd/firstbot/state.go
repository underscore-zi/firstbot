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

func (s *State) ClaimedBy() (string, int, int) {
	return s.raw.ClaimedBy, s.raw.TotalClaims[s.raw.ClaimedBy], s.raw.Streak
}
