package main

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func TestOnCallStoreMemory(t *testing.T) {
	store := InMemoryOnCallStore{
		OnCallEntries: make(map[string]OnCallEntry),
		currentID:     0,
	}

	t.Run("test normal OnCallEntry Creation", func(t *testing.T) {
		entryReq := OnCallEntry{
			Service:  "payment",
			Username: "tom",
			StartsAt: time.Now().Add(-1 * time.Minute),
			EndsAt:   time.Now().Add(100 * time.Minute),
		}
		entry1, err1 := store.Create(context.Background(), entryReq)
		if err1 != nil {
			t.Fatalf("expect no error, get %v", err1)
		}
		entry2, err2 := store.Create(context.Background(), entryReq)
		if err2 != nil {
			t.Fatalf("expect no error, get %v", err2)
		}
		if entry1.ID != OnCallEntryPrefix+"1" || entry2.ID != OnCallEntryPrefix+"2" {
			t.Fatal("ID not as expected")
		}
		if entry1.Service != entryReq.Service {
			t.Fatal("Not the same Service")
		}
		if entry1.Username != entryReq.Username {
			t.Fatal("Not the same Username")
		}
		if entry1.StartsAt.Equal(entryReq.StartsAt) == false {
			fmt.Println(entry1.StartsAt)
			fmt.Println(entryReq.StartsAt)
			t.Fatal("Not the same StartsAt")
		}
		if entry1.EndsAt.Equal(entryReq.EndsAt) == false {
			t.Fatal("Not the same EndsAt")
		}
	})

	t.Run("successful CurrentOncall", func(t *testing.T) {
		username, err := store.CurrentOnCall(context.Background(), "payment")
		if err != nil {
			t.Fatalf("expect no error, get error %v", err)
		}
		if username != "tom" {
			t.Fatalf("expect username %v, get %v", "tom", username)
		}
	})
	t.Run("fail CurrentOnCall", func(t *testing.T) {
		_, err := store.CurrentOnCall(context.Background(), "not-exist")
		if err != OnCallUserNotFound {
			t.Fatalf("expect error %v, get error %v", OnCallUserNotFound, err)
		}
	})
}
