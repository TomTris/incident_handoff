package main

import (
	"context"
	"strconv"
)

type UserStore interface {
	Create(u User) (User, error)
	GetByUsername(ctx context.Context, username string) (User, error)
}

type InMemoryUserStore struct {
	users     map[string]User // username - User
	currentID int
}

func NewInMemoryUserStore(seed []User) *InMemoryUserStore {
	m := make(map[string]User, len(seed))
	for _, u := range seed {
		m[u.Username] = u
	}
	return &InMemoryUserStore{users: m}
}

func (s *InMemoryUserStore) Create(u User) (User, error) {
	_, ok := s.users[u.Username]
	if ok == true {
		return User{}, ErrUserAlreadyExist
	}

	s.currentID++
	ID := UserPrefix + strconv.Itoa(s.currentID)
	u.ID = ID
	s.users[u.Username] = u
	return u, nil
}

func (s *InMemoryUserStore) GetByUsername(_ context.Context, username string) (User, error) {
	u, ok := s.users[username]
	if ok == false {
		return User{}, ErrUserNotFound
	}
	return u, nil
}
