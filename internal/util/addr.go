package util

func BoolAddr(b bool) *bool {
	boolVar := b
	return &boolVar
}
