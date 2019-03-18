package p2p

import "github.com/pkg/errors"

func keyset(m map[string]bool) []string {
	keyset := make([]string, 0, len(m))
	for k := range m {
		keyset = append(keyset, k)
	}
	return keyset
}

func setAsList(m map[string]bool) []string {
	list := make([]string, 0, len(m))
	for k := range m {
		if m[k] == true {
			list = append(list, k)
		}
	}
	return list
}

// helper function to find minimum of two integers
func min(a int, b int) int {
	if a <= b {
		return a
	} else {
		return b
	}
}

// helper function to find index of an element in a slice
func indexOf(element string, slice []string) (int, error) {
	for i, v := range slice {
		if element == v {
			return i, nil
		}
	}
	return -1, errors.Errorf("Element not found in list")
}
