package main

type FlagStore struct {
	Overwrite bool
	Diff      bool
	Recursive bool
}

func NewFlagStore() *FlagStore {
	return &FlagStore{}
}
