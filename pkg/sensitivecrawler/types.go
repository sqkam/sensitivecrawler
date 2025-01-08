package sensitivecrawler

import "context"

type Service interface {
	Run(ctx context.Context)
	AddTask(site string, options ...TaskOption)
	RunOneTask(ctx context.Context)
}
