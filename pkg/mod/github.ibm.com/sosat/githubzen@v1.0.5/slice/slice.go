package slice

// Find returns the first occurrence of the given string in the slice
// and if the string was found.
func Find(a []string, x string) (int, bool) {
	for i, n := range a {
		if x == n {
			return i, true
		}
	}
	return len(a), false
}

// Contains tells whether a contains x.
// func Contains(a []string, x string) bool {
// 	for _, n := range a {
// 		if x == n {
// 			return true
// 		}
// 	}
// 	return false
// }
