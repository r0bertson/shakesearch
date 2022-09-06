package utils

// Unique returns all unique elements in a group of slices
func Unique(intSlices ...[]int) []int {
	uniqueMap := map[int]bool{}

	for _, intSlice := range intSlices {
		for _, element := range intSlice {
			uniqueMap[element] = true
		}
	}

	result := make([]int, 0, len(uniqueMap))

	for key := range uniqueMap {
		result = append(result, key)
	}

	return result
}
