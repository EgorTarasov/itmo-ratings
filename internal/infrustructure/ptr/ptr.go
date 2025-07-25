package ptr

func Get[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

func To[T any](v T) *T {
	return &v
}

func ToOrNil[T any](v T, cond bool) *T {
	if cond {
		return &v
	}
	return nil
}
