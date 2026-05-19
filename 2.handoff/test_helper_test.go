package main

// type testENV struct {
// 	router http.Handler
// 	store  *MongoStore
// 	token  string //JWT
// }

// func setupTestENV() {
// 	t.Helper()
// 	os.Setenv("HANDOFF_CONNECT_STRING", "mongodb://127.0.0.1:27017/?directConnection=true")
// 	config := loadConfig()
// 	client, mongoStore := NewStore(config)
// 	incHandler := IncidentHandler{Store: mongoStore, Registry: NewRegistry()}
// 	router := getRouter(&incHandler, client, prometheus.NewRegistry())
// 	go incHandler.Registry.run()
// 	instrumented, ok := mongoStore.(*InstrumentedStore)
// 	ms := instrumented.s.(*MongoStore)
// 	ms.DropAll(context.Background())
// 	token, _ := createSignedToken(User{ID: "u1", Username: "anh", Role: "engineer"})
// 	t.Cleanup(func() {
// 		close(incHandler.Registry.done)
// 	})

// 	return &testEnv{router: router, store: ms, token: token}
// }
