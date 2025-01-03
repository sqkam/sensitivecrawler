package sensitivematcher

type SensitiveMatcher interface {
	Match(b []byte, name string)
}
