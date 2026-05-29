package main

import (
	"context"
	"maps"
	"strconv"
	"time"
)

type OnCallStore interface {
	Create(ctx context.Context, entry OnCallEntry) (OnCallEntry, error)
	CurrentOnCall(ctx context.Context, service string) (string, error)
}

type InMemoryOnCallStore struct {
	OnCallEntries map[string]OnCallEntry
	currentID     int
}

func NewInMemoryOnCallStore(entries map[string]OnCallEntry) *InMemoryOnCallStore {
	s := InMemoryOnCallStore{
		OnCallEntries: make(map[string]OnCallEntry),
		currentID:     0,
	}
	maps.Copy(s.OnCallEntries, entries)
	return &s
}

func (store *InMemoryOnCallStore) Create(ctx context.Context, entry OnCallEntry) (OnCallEntry, error) {
	store.currentID++
	ID := OnCallEntryPrefix + strconv.Itoa(store.currentID)
	entry.ID = ID
	store.OnCallEntries[entry.ID] = entry
	return entry, nil
}

func (store *InMemoryOnCallStore) CurrentOnCall(ctx context.Context, service string) (string, error) {
	now := time.Now()
	for _, each := range store.OnCallEntries {
		if each.Service == service && each.StartsAt.Before(now) && each.EndsAt.After(now) {
			return each.Username, nil
		}
	}
	return "", OnCallUserNotFound
}
