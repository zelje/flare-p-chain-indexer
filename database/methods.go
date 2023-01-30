package database

import "time"

func (s *State) Update(nextIndex, lastIndex uint64) {
	s.NextDBIndex = nextIndex
	s.LastChainIndex = lastIndex
	s.Updated = time.Now()
}
