package post_finder

func AppendUint64SliceKeepOrder(slice []uint64, u uint64) []uint64 {
	var x bool
	for i, item := range slice {
		if item >= u {
			slice = append(append(slice[:i], u), slice[i:]...)
			x = true
			break
		}
	}
	if !x {
		slice = append(slice, u)
	}
	return slice
}
