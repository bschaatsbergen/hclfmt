package main

type FlagStore struct {
	Overwrite bool
	Diff      bool
}

func NewFlagStore() *FlagStore {
	return &FlagStore{}
}
