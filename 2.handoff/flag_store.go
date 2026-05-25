package main

import (
	"hash/fnv"
	"sync"
)

type FlagStore struct {
	m     sync.RWMutex
	Flags map[string]FeatureFlag
}

func CreateFlagStore() FlagStore {
	return FlagStore{
		Flags: make(map[string]FeatureFlag),
	}
}

func (flagStore *FlagStore) Create(f FeatureFlag) error {
	flagStore.m.Lock()
	defer flagStore.m.Unlock()
	_, ok := flagStore.Flags[f.Name]
	if ok == true {
		return ErrFlagAlreadyExist
	}
	flagStore.Flags[f.Name] = f
	return nil
}

func (flagStore *FlagStore) Update(u FeatureFlagUpdate) error {
	flagStore.m.Lock()
	defer flagStore.m.Unlock()
	flag, ok := flagStore.Flags[u.Name]
	if ok == false {
		return ErrFlagNotfound
	}
	if u.Enabled != nil {
		flag.Enabled = *u.Enabled
	}
	if u.Rollout != nil {
		flag.Rollout = *u.Rollout
	}
	flagStore.Flags[u.Name] = flag
	return nil
}

func (flagStore *FlagStore) AllFlags() ([]FeatureFlag, error) {
	flagStore.m.Lock()
	defer flagStore.m.Unlock()

	flags := make([]FeatureFlag, 0, len(flagStore.Flags))
	for _, f := range flagStore.Flags {
		flags = append(flags, f)
	}
	return flags, nil
}

func (flagStore *FlagStore) Evaluate(flagName string, userID string) (*FlagEvaluateAnswer, error) {
	h1 := fnv.New32a()
	h1.Write([]byte(flagName + ":rollout" + userID))
	hashRollout := h1.Sum32()

	h2 := fnv.New32a()
	h2.Write([]byte(flagName + ":variants" + userID))
	hashVariants := h2.Sum32()

	flagStore.m.RLock()
	defer flagStore.m.RUnlock()

	flag, ok := flagStore.Flags[flagName]
	if ok == false {
		return nil, ErrFlagNotfound
	}

	var answer FlagEvaluateAnswer
	bucket := hashRollout % 100
	if userID == "" || flag.Enabled == false || int(bucket) >= flag.Rollout {
		answer = FlagEvaluateAnswer{
			Name:      flagName,
			UserID:    userID,
			Enabled:   flag.Enabled,
			InRollout: false,
			Variant:   nil,
		}
	} else {
		variant := hashVariants % uint32(len(flag.Variants))
		answer = FlagEvaluateAnswer{
			Name:      flagName,
			UserID:    userID,
			Enabled:   true,
			InRollout: true,
			Variant:   &flag.Variants[variant],
		}
	}
	return &answer, nil
}
