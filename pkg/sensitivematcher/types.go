package sensitivematcher

import "context"

type SensitiveMatcher interface {
	Match(ctx context.Context, b []byte) (matchStrings []string)
}
