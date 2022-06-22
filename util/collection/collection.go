package collection

func Contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}

	return false
}

func Intersect(a, b []string) []string {
	m := make(map[string]struct{}, len(b))

	for _, v := range b {
		m[v] = struct{}{}
	}

	var s []string
	for _, v := range a {
		if _, ok := m[v]; ok {
			s = append(s, v)
		}
	}

	return s
}
