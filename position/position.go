package position

import (
	"queuev2/store"
)

type Position struct {
	store store.Store
}

const positionSet = "p_set"

func NewPosition(store store.Store) *Position {
	p := &Position{
		store: store,
	}
	store.DeleteKey(positionSet)
	return p
}

func (p *Position) AddItem(item string, score int) error {
	err := p.store.AddSortedSet(positionSet, score, item)
	return err
}

func (p *Position) RemoveItem(item string) error {
	return p.store.RemoveSortedSet(positionSet, item)
}

func (p *Position) GetPosition(item string) (int, error) {
	return p.store.GetRankSortedSet(positionSet, item)
}
