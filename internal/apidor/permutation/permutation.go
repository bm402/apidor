package permutation

import (
	"strconv"
)

var cache map[string][]string

// GetAllCombinationsOfHighAndLowPrivilege is a permutations function that returns all the permutations
// of high and low privilege for n variables
func GetAllCombinationsOfHighAndLowPrivilege(n int) []string {
	cacheKey := "full-" + strconv.Itoa(n)
	if cachedPermutation, ok := getPermutationsFromCache(cacheKey); ok {
		return cachedPermutation
	}

	permutations := []string{}
	curPermutation := ""
	fullPermutationsBuilder(&permutations, &curPermutation, n, 0)

	if len(permutations) > 1 {
		permutations = permutations[:len(permutations)-1]
	}
	setPermutationsInCache(cacheKey, permutations)

	return permutations
}

func fullPermutationsBuilder(permutations *[]string, curPermutation *string, maxLevel int, curLevel int) {
	if curLevel == maxLevel {
		*permutations = append(*permutations, *curPermutation)
		return
	}

	*curPermutation += "h"
	fullPermutationsBuilder(permutations, curPermutation, maxLevel, curLevel+1)
	*curPermutation = (*curPermutation)[:len(*curPermutation)-1]

	*curPermutation += "l"
	fullPermutationsBuilder(permutations, curPermutation, maxLevel, curLevel+1)
	*curPermutation = (*curPermutation)[:len(*curPermutation)-1]
}

// GetCombinationsOfOppositePrivilege is a permutations function that returns permutations
// of opposite privilege for n variables
func GetCombinationsOfOppositePrivilege(n int) []string {
	cacheKey := "opposite-" + strconv.Itoa(n)
	if cachedPermutation, ok := getPermutationsFromCache(cacheKey); ok {
		return cachedPermutation
	}

	permutations := []string{}
	curPermutation := ""
	oppositePermutationsBuilder(&permutations, &curPermutation, n, 0)
	setPermutationsInCache(cacheKey, permutations)

	return permutations
}

func oppositePermutationsBuilder(permutations *[]string, curPermutation *string, maxLevel int, curLevel int) {
	if curLevel == maxLevel {
		*permutations = append(*permutations, *curPermutation)
		return
	}

	if curLevel%2 == 0 || (*curPermutation)[curLevel-1] == 'l' {
		*curPermutation += "h"
		oppositePermutationsBuilder(permutations, curPermutation, maxLevel, curLevel+1)
		*curPermutation = (*curPermutation)[:len(*curPermutation)-1]
	}

	if curLevel%2 == 0 || (*curPermutation)[curLevel-1] == 'h' {
		*curPermutation += "l"
		oppositePermutationsBuilder(permutations, curPermutation, maxLevel, curLevel+1)
		*curPermutation = (*curPermutation)[:len(*curPermutation)-1]
	}
}

func getPermutationsFromCache(key string) ([]string, bool) {
	if cache == nil {
		cache = make(map[string][]string)
		return []string{}, false
	}
	permutation, status := cache[key]
	return permutation, status
}

func setPermutationsInCache(key string, permutations []string) {
	cache[key] = permutations
}
