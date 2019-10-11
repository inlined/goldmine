package goldmine

/*

// positionMemo shows what paths we've seen at a given spot. The first ndxVisited
// elements have already been considered in another path and we no longer attempt
// to keep them ordered.
type positionMemo struct {
	paths      []*Path
	ndxVisited int
}

func MakeMemoStack(maxMoves int) []map[Vertex]*positionMemo {
	memo := make([]map[Vertex]*positionMemo, maxMoves)
	for ndx := range memo {
		memo[ndx] = make(map[Vertex]*positionMemo, maxMoves)
	}
	return memo
}

func (m *positionMemo) insert(s Strategy, contender *Path) {
	// If no strategy can differentiate these, don't keep duplicates.
	for _, prior := range m.paths {
		if prior.rawValue == contender.rawValue && prior.pickaxes == contender.pickaxes {
			return
		}
	}

	val := s.value(contender)
	for ndx, prior := range m.paths[m.ndxVisited:] {
		if val > s.value(prior) {
			m.paths = append(m.paths, nil)
			copy(m.paths[m.ndxVisited+ndx+1:], m.paths[m.ndxVisited+ndx:])
			m.paths[m.ndxVisited+ndx] = contender
			return
		}
	}
	m.paths = append(m.paths, contender)
}

func collect(memos map[Vertex]*positionMemo, keepFromEach int) []*Path {
	total := make([]*Path, 0, len(memos)*keepFromEach) // Reuse?
	for _, memo := range memos {
		toTake := keepFromEach
		if toTake > len(memo.paths)-memo.ndxVisited {
			toTake = len(memo.paths) - memo.ndxVisited
		}
		if toTake == 0 {
			continue
		}
		total = append(total, memo.paths[memo.ndxVisited:toTake]...)
		memo.ndxVisited += toTake
	}
	return total
}
*/
