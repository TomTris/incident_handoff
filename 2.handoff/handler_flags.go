package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"sync"
)

// For Minimalist and just for understanding purpose sofar,
// I will keep this flag feature in 1 file
// And not create second flag_store.go

type FlagStore struct {
	m    sync.RWMutex
	Flag map[string]*FeatureFlag
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

	flag, ok := flagStore.Flag[flagName]
	if ok == false {
		return nil, ErrFlagNotfound
	}

	var answer FlagEvaluateAnswer
	bucket := hashRollout % 100
	fmt.Println(bucket)
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

func CreateFlagStore() FlagStore {
	return FlagStore{
		Flag: make(map[string]*FeatureFlag),
	}
}

func (incHandler *IncidentHandler) CreateFlag(w http.ResponseWriter, r *http.Request) {
	f := FeatureFlag{}

	err := json.NewDecoder(r.Body).Decode(&f)
	requestID := r.Context().Value(requestIDKey).(string)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   ErrBadRequest.Error(),
			RequestID: requestID,
		})
		return
	}
	err = f.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: requestID,
		})
		return
	}
	// flag Store
	flagName := f.Name
	incHandler.FlagStore.m.Lock()
	incHandler.FlagStore.Flag[flagName] = &f
	defer incHandler.FlagStore.m.Unlock()

	writeJSON(w, http.StatusCreated, requestID, f)
}

func (incHandler *IncidentHandler) UpdateFlag(w http.ResponseWriter, r *http.Request) {
	u := FeatureFlagUpdate{}

	requestID := r.Context().Value(requestIDKey).(string)
	err := json.NewDecoder(r.Body).Decode(&u)
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   ErrBadRequest.Error(),
			RequestID: requestID,
		})
		return
	}

	u.Name = r.PathValue("name")
	err = u.Validate()
	if err != nil {
		writeError(w, http.StatusBadRequest, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   err.Error(),
			RequestID: requestID,
		})
		return
	}

	// flag Store
	flagName := u.Name
	incHandler.FlagStore.m.Lock()
	defer incHandler.FlagStore.m.Unlock()
	flag, ok := incHandler.FlagStore.Flag[flagName]
	if ok == false {
		writeError(w, http.StatusNotFound, ErrorMessageJSON{
			ErrorCode: FLAG_NOT_FOUND,
			Message:   ErrFlagNotfound.Error(),
			RequestID: requestID,
		})
		return
	}
	if u.Enabled != nil {
		flag.Enabled = *u.Enabled
	}
	if u.Rollout != nil {
		flag.Rollout = *u.Rollout
	}
	w.WriteHeader(http.StatusNoContent)
}

func (incHandler *IncidentHandler) ListAllFlag(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(requestIDKey).(string)

	// flag Store
	incHandler.FlagStore.m.RLock()
	defer incHandler.FlagStore.m.RUnlock()
	writeJSON(w, http.StatusOK, requestID, incHandler.FlagStore.Flag)
}

type FlagEvaluateAnswer struct {
	Name      string  `json:"name"`
	UserID    string  `json:"user_id"`
	Enabled   bool    `json:"enabled"`
	InRollout bool    `json:"in_rollout"`
	Variant   *string `json:"variants"`
}

func (incHandler *IncidentHandler) Evaluate(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(requestIDKey).(string)
	flagName := r.PathValue("name")
	userID := r.URL.Query().Get("user_id")

	// Validate
	if userID == "" {
		writeError(w, http.StatusNotFound, ErrorMessageJSON{
			ErrorCode: BAD_REQUEST,
			Message:   ErrBadRequest.Error(),
			RequestID: requestID,
		})
		return
	}

	// evaluate and answer
	flagAnswer, err := incHandler.FlagStore.Evaluate(flagName, userID)
	if err != nil {
		if errors.Is(err, ErrFlagNotfound) {
			writeError(w, http.StatusNotFound, ErrorMessageJSON{
				ErrorCode: FLAG_NOT_FOUND,
				Message:   ErrBadRequest.Error(),
				RequestID: requestID,
			})
		} else {
			writeError(w, http.StatusInternalServerError, ErrorMessageJSON{
				ErrorCode: INTERNAL_SERVER_ERROR,
				Message:   err.Error(),
				RequestID: requestID,
			})
		}
		return
	}

	writeJSON(w, http.StatusOK, requestID, flagAnswer)
}
