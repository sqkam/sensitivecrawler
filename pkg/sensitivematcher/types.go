package sensitivematcher

type SensitiveMatcher interface {
	Match(b []byte, name string) (matchStr string, ok bool)
}
