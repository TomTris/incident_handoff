package main

func derefOrDefault(s *string, de string) string {
	if s == nil {
		return de
	}
	return *s
}
