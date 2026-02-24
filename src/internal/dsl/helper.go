package dsl

import (
	"strconv"
	"strings"
)

// getFieldValue returns a value from a flat context map by key.
// If the key is missing, it returns (nil, nil) so comparisons evaluate to false.
func getFieldValue(ctx map[string]interface{}, field string) (interface{}, error) {
	if ctx == nil {
		return nil, nil
	}
	if v, ok := ctx[field]; ok {
		return v, nil
	}
	return nil, nil
}

func compareFloats(a float64, op string, b float64) bool {
	switch op {
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	case "=":
		return a == b
	case "!=":
		return a != b
	}
	return false
}

func compareStrings(a string, op string, b string) bool {
	switch op {
	case "=":
		return a == b
	case "!=":
		return a != b
	// > or < on strings is usually not required by basic business logic,
	// but valid in SQL. Implemented here for completeness.
	case ">":
		return a > b
	case ">=":
		return a >= b
	case "<":
		return a < b
	case "<=":
		return a <= b
	}
	return false
}

// compareVersions compares dotted numeric versions like "1.5.7" or "2.3".
// returns (cmp, true) where cmp < 0 if a < b, 0 if equal, > 0 if a > b.
// if either value is not a dotted numeric version, it returns (0, false).
func compareVersions(a, b string) (int, bool) {
	pa, okA := parseVersion(a)
	pb, okB := parseVersion(b)
	if !okA || !okB {
		return 0, false
	}

	n := len(pa)
	if len(pb) > n {
		n = len(pb)
	}

	for i := 0; i < n; i++ {
		va := 0
		vb := 0
		if i < len(pa) {
			va = pa[i]
		}
		if i < len(pb) {
			vb = pb[i]
		}
		if va < vb {
			return -1, true
		}
		if va > vb {
			return 1, true
		}
	}
	return 0, true
}

func parseVersion(s string) ([]int, bool) {
	if s == "" {
		return nil, false
	}

	parts := strings.Split(s, ".")

	out := make([]int, len(parts))
	for i, p := range parts {
		if p == "" {
			return nil, false
		}

		n, err := strconv.Atoi(p)
		if err != nil || n < 0 {
			return nil, false
		}

		out[i] = n
	}

	return out, true
}
