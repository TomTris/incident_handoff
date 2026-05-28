package main

import "context"

type UserStore interface {
	GetByUsername(ctx context.Context, username string) (User, error)
}

type InMemoryUserStore struct {
	users map[string]User
}

func NewInMemoryUserStore(seed []User) *InMemoryUserStore {
	m := make(map[string]User, len(seed))
	for _, u := range seed {
		m[u.Username] = u
	}
	return &InMemoryUserStore{users: m}
}

func (s *InMemoryUserStore) GetByUsername(_ context.Context, username string) (User, error) {
	u, ok := s.users[username]
	if ok == false {
		return User{}, ErrUserNotFound
	}
	return u, nil
}
