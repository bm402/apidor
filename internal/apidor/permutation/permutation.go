package permutation

var cache map[int][]string

// GetAllCombinationsOfHighAndLowPrivilege is a permutations function that returns all the permutations
// of high and low privilege for n variables
func GetAllCombinationsOfHighAndLowPrivilege(n int) []string {
	if cachedPermutation, ok := getPermutationsFromCache(n); ok {
		return cachedPermutation
	}

	permutations := []string{}
	curPermutation := ""
	permutationsBuilder(&permutations, &curPermutation, n, 0)
	setPermutationsInCache(n, permutations)

	return permutations
}

func permutationsBuilder(permutations *[]string, curPermutation *string, maxLevel int, curLevel int) {
	if curLevel == maxLevel {
		*permutations = append(*permutations, *curPermutation)
		return
	}

	*curPermutation += "h"
	permutationsBuilder(permutations, curPermutation, maxLevel, curLevel+1)
	*curPermutation = (*curPermutation)[:len(*curPermutation)-1]

	*curPermutation += "l"
	permutationsBuilder(permutations, curPermutation, maxLevel, curLevel+1)
	*curPermutation = (*curPermutation)[:len(*curPermutation)-1]
}

func getPermutationsFromCache(n int) ([]string, bool) {
	if cache == nil {
		cache = make(map[int][]string)
		return []string{}, false
	}
	permutation, status := cache[n]
	return permutation, status
}

func setPermutationsInCache(n int, permutations []string) {
	cache[n] = permutations
}
