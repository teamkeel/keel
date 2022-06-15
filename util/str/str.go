package str

func PadRight(str string, padAmount int) string {
	for len(str) < padAmount {
		str += " "
	}

	return str
}
