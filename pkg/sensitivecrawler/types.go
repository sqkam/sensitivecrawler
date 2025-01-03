package sensitivecrawler

import "context"

type Service interface {
	Run(ctx context.Context)
}
