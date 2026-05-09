package main

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
