package tracking

import (
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/pkg/errors"

	"github.com/movidesk/go-web/logs"
)

var (
	once sync.Once
)

// Options Tracking options
type Options struct {
	SentryDsn string
}

// Init Init tracking
func Init(o Options) {
	once.Do(func() {
		log := logs.Single()
		err := sentry.Init(sentry.ClientOptions{
			Dsn: o.SentryDsn,
		})
		if err != nil {
			log.Error(errors.Wrapf(err, "Unable to init sentry"))
		}
	})
}
