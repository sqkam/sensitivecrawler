package sensitivematcher

type SensitiveMatcher interface {
	Match(b []byte) (matchStrings []string)
}
