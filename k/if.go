package k

func If[T any](condition bool, trueVal, falseVal T) T {
	if condition {
		return trueVal
	}
	return falseVal
}

func IfLazy[T any](condition bool, trueFunc func() T, falseValue T) T {
	if condition {
		return trueFunc()
	}
	return falseValue
}
