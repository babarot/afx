package manager

func allTrue(list []bool) bool {
	if len(list) == 0 {
		return false
	}
	for _, item := range list {
		if !item {
			return false
		}
	}
	return true
}
