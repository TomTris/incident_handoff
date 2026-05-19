package main

// type InstrumentedStore struct {
// 	s Store
// }

// func (s *InstrumentedStore) MetricInit() {
// 	incidents, err := s.s.ListIncidents(context.Background(), IncidentFilter{})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	for _, incident := range incidents {
// 		incidentTotal.WithLabelValues(incident.Status).Inc()
// 		totalEntries.Add(float64(len(incident.Entries)))
// 	}
// }

// func (s *InstrumentedStore) CreateIncident(ctx context.Context, inc Incident) (Incident, error) {
// 	timer := prometheus.NewTimer(dbQueryDurationSeconds.WithLabelValues("create_incident"))
// 	inc, err := s.s.CreateIncident(ctx, inc)
// 	timer.ObserveDuration()
// 	if err == nil {
// 		incidentTotal.WithLabelValues(inc.Status).Inc()
// 	}
// 	return inc, err
// }

// func (s *InstrumentedStore) GetIncident(ctx context.Context, id string) (Incident, error) {
// 	timer := prometheus.NewTimer(dbQueryDurationSeconds.WithLabelValues("get_incident"))
// 	defer timer.ObserveDuration()
// 	return s.s.GetIncident(ctx, id)
// }

// func (s *InstrumentedStore) AddEntry(ctx context.Context, incidentID string, entry TimelineEntry) (TimelineEntry, error) {
// 	timer := prometheus.NewTimer(dbQueryDurationSeconds.WithLabelValues("add_entry"))
// 	entry, err := s.s.AddEntry(ctx, incidentID, entry)
// 	timer.ObserveDuration()

// 	if err == nil {
// 		totalEntries.Inc()
// 	}
// 	return entry, err
// }

// func (s *InstrumentedStore) ListIncidents(ctx context.Context, filter IncidentFilter) ([]Incident, error) {
// 	timer := prometheus.NewTimer(dbQueryDurationSeconds.WithLabelValues("list_incident"))
// 	defer timer.ObserveDuration()
// 	return s.s.ListIncidents(ctx, filter)
// }

// func (s *InstrumentedStore) UpdateIncident(ctx context.Context, id string, update IncidentUpdate) (Incident, error) {
// 	timer := prometheus.NewTimer(dbQueryDurationSeconds.WithLabelValues("update_incident"))
// 	incBefore, err := s.s.UpdateIncident(ctx, id, update)
// 	timer.ObserveDuration()
// 	if err == nil && update.Status != nil {
// 		incidentTotal.WithLabelValues(*update.Status).Inc()
// 		incidentTotal.WithLabelValues(incBefore.Status).Dec()
// 	}
// 	return incBefore, err
// }
