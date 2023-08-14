package prices

// list of symbols we support

func GetUniverseAssets() []string {
	out := []string{}
	// TODO - get all historic symbols?
	// only issue is that it sorta breaks
	// the generalized product. oh well,
	// let's just do for open positions
	oneoffs := []string{
		"COIN",
		"HOOD",
		"AAPL",
	}

	out = append(
		out,
		GetSpySymbols()...,
	)
	out = append(
		out,
		oneoffs...,
	)

	return removeDuplicates(out)
}

func GetSpySymbols() []string {
	return []string{}
}

func removeDuplicates(input []string) []string {
	seen := make(map[string]bool)
	result := []string{}

	for _, item := range input {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}

	return result
}
