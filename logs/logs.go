package logs

import (
	"fmt"
	"io/ioutil"
	"sync"
	"time"

	"github.com/olivere/elastic/v7"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"gopkg.in/sohlich/elogrus.v7"
)

var (
	once     sync.Once
	instance *logrus.Logger
)

// AppOptions Application options
type AppOptions struct {
	_       struct{}
	AppHost string
}

// ElkOptions Elastic options
type ElkOptions struct {
	_                 struct{}
	ElkDsn            string
	ElkIndex          string
	ElkSync           bool
	ElkSniff          bool
	ElkHealth         bool
	ElkHealthInterval time.Duration
	ElkHealthTimeout  time.Duration
}

// OptionsFn Options function type
type OptionsFn func(*logrus.Logger, *Options)

// SetDiscardOutput Defines if should discard output
func SetDiscardOutput() OptionsFn {
	return func(l *logrus.Logger, o *Options) {
		l.Out = ioutil.Discard
	}
}

// SetElkSync Defines if should sync elk
func SetElkSync(b bool) OptionsFn {
	return func(l *logrus.Logger, o *Options) {
		o.ElkSync = b
	}
}

// SetTestingMode Testing mode discards the outputs and doesn't sync elk
func SetTestingMode() OptionsFn {
	return func(l *logrus.Logger, o *Options) {
		SetDiscardOutput()(l, o)
		SetElkSync(false)(l, o)
	}
}

// Options Log options
type Options struct {
	AppOptions
	ElkOptions
}

// DefaultAppOptions Default app options
var DefaultAppOptions = AppOptions{
	AppHost: "localhost",
}

// DefaultElkOptions Default elastic options
var DefaultElkOptions = ElkOptions{
	ElkDsn:            "http://localhost:9200",
	ElkIndex:          "org.module.app.%s",
	ElkSync:           true,
	ElkSniff:          false,
	ElkHealth:         true,
	ElkHealthInterval: time.Second * 30,
	ElkHealthTimeout:  time.Second * 3,
}

// DefaultOptions Default options
var DefaultOptions = Options{
	DefaultAppOptions,
	DefaultElkOptions,
}

// New Init new log
func New(fns ...OptionsFn) *logrus.Logger {
	o := DefaultOptions
	log := logrus.New()
	for _, fn := range fns {
		fn(log, &o)
	}
	if o.ElkSync {
		registerElk(log, o)
	}
	return log
}

func registerElk(log *logrus.Logger, o Options) {
	client, err := elastic.NewClient(
		elastic.SetURL(o.ElkDsn),
		elastic.SetHealthcheck(o.ElkHealth),
		elastic.SetHealthcheckInterval(o.ElkHealthInterval),
		elastic.SetHealthcheckTimeout(o.ElkHealthTimeout),
		elastic.SetSniff(o.ElkSniff), // disable cluster discovery
	)
	if err != nil {
		log.Error(errors.Wrapf(err, "Unable to create new elasticsearch client on dsn %s", o.ElkDsn))
	}

	hook, err := elogrus.NewBulkProcessorElasticHookWithFunc(client, o.AppHost, logrus.DebugLevel, func() string {
		n := time.Now().UTC()
		y := n.Year()
		m := n.Month()
		d := n.Day()
		s := fmt.Sprintf("%d.%02d.%02d", y, m, d)
		return fmt.Sprintf(o.ElkIndex, s)
	})
	if err != nil {
		log.Error(errors.Wrapf(err, "Unable to create new async elastic hook"))
	}
	if hook != nil {
		log.Hooks.Add(hook)
	}
}

// Init Init new instance
func Init(fns ...OptionsFn) {
	once.Do(func() {
		instance = New(fns...)
	})
}

// Single Use a silgle log instance
func Single() *logrus.Logger {
	return instance
}
