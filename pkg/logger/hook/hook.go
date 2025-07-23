package hook

import (
	"github.com/LyricTian/queue"
	"go.uber.org/zap/zapcore"
)

var defaultOptions = options{
	maxQueues:  512,
	maxWorkers: 1,
	levels: []zapcore.Level{
		zapcore.PanicLevel,
		zapcore.FatalLevel,
		zapcore.ErrorLevel,
		zapcore.WarnLevel,
		zapcore.InfoLevel,
		zapcore.DebugLevel,
	},
}

// ExecCloser write the zapcore entry to the store and close the store
type ExecCloser interface {
	Exec(entry *zapcore.Entry) error
	Close() error
}

// FilterHandle a filter handler
type FilterHandle func(*zapcore.Entry) *zapcore.Entry

type options struct {
	maxQueues  int
	maxWorkers int
	extra      map[string]interface{}
	filter     FilterHandle
	levels     []zapcore.Level
}

// SetMaxQueues set the number of buffers
func SetMaxQueues(maxQueues int) Option {
	return func(o *options) {
		o.maxQueues = maxQueues
	}
}

// SetMaxWorkers set the number of worker threads
func SetMaxWorkers(maxWorkers int) Option {
	return func(o *options) {
		o.maxWorkers = maxWorkers
	}
}

// SetExtra set extended parameters
func SetExtra(extra map[string]interface{}) Option {
	return func(o *options) {
		o.extra = extra
	}
}

// SetFilter set the entry filter
func SetFilter(filter FilterHandle) Option {
	return func(o *options) {
		o.filter = filter
	}
}

// SetLevels set the available log level
func SetLevels(levels ...zapcore.Level) Option {
	return func(o *options) {
		if len(levels) == 0 {
			return
		}
		o.levels = levels
	}
}

// Option a hook parameter options
type Option func(*options)

// New creates a hook to be added to an instance of logger
func New(exec ExecCloser, opt ...Option) *Hook {
	opts := defaultOptions
	for _, o := range opt {
		o(&opts)
	}

	q := queue.NewQueue(opts.maxQueues, opts.maxWorkers)
	q.Run()

	return &Hook{
		opts: opts,
		q:    q,
		e:    exec,
	}
}

// Hook to send logs to a mongo database
type Hook struct {
	opts options
	q    *queue.Queue
	e    ExecCloser
}

// Levels returns the available logging levels
func (h *Hook) Levels() []zapcore.Level {
	return h.opts.levels
}

// Fire is called when a log event is fired
func (h *Hook) Fire(entry *zapcore.Entry) error {

	// h.q.Push(queue.NewJob(entry.Dup(), func(v interface{}) {
	// 	h.exec(v.(*zapcore.Entry))
	// }))
	return nil
}

func (h *Hook) exec(entry *zapcore.Entry) {
	// for k, v := range h.opts.extra {
	// 	if _, ok := entry.Data[k]; !ok {
	// 		entry.Data[k] = v
	// 	}
	// }

	// if filter := h.opts.filter; filter != nil {
	// 	entry = filter(entry)
	// }

	// err := h.e.Exec(entry)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "[zapcore-hook] execution error: %s", err.Error())
	// }
}

// Flush waits for the log queue to be empty
func (h *Hook) Flush() {
	h.q.Terminate()
	h.e.Close()
}
