package utility

func CalculateLevel(exp int) int {
	//Calculate user's level based on exp
	var min, max, base, level int = 0, 0, 100, 0
	for i := 0; ; i++ {
		min = base * (i*i - i)
		max = base * (i*i + i)
		if min <= exp && exp < max {
			level = i
			break
		}
	}

	return level
}
