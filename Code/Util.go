package Code

func keyset(m map[string]bool) []string {
	keyset := make([]string, 0, len(m))
	for k := range m {
		keyset = append(keyset, k)
	}
	return keyset
}
