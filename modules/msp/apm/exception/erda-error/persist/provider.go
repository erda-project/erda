package persist

import (
	"context"
	"fmt"
	"github.com/erda-project/erda-infra/base/logs"
	"github.com/erda-project/erda-infra/base/servicehub"
	"github.com/erda-project/erda-infra/providers/kafka"
	"github.com/erda-project/erda/modules/core/monitor/storekit"
	"github.com/erda-project/erda/modules/msp/apm/exception/erda-error/storage"
	"time"
)

type (
	config struct {
		Input         kafka.BatchReaderConfig `file:"input"`
		Parallelism      int                  `file:"parallelism" default:"1"`
		BufferSize       int                  `file:"buffer_size" default:"1024"`
		ReadTimeout      time.Duration        `file:"read_timeout" default:"5s"`
		IDKeys           []string             `file:"id_keys"`
		PrintInvalidError bool                 `file:"print_invalid_error" default:"false"`
	}
	provider struct {
		Cfg           *config
		Log           logs.Logger
		Kafka         kafka.Interface `autowired:"kafka"`
		StorageWriter storage.Storage `autowired:"error-storage-writer"`

		storage storage.Storage
		stats   Statistics
		validator Validator
		metadata  MetadataProcessor
	}
)

func (p *provider) Init(ctx servicehub.Context) (err error) {

	p.validator = newValidator(p.Cfg)
	if runner, ok := p.validator.(servicehub.ProviderRunnerWithContext); ok {
		ctx.AddTask(runner.Run, servicehub.WithTaskName("error validator"))
	}

	p.metadata = newMetadataProcessor(p.Cfg)
	if runner, ok := p.metadata.(servicehub.ProviderRunnerWithContext); ok {
		ctx.AddTask(runner.Run, servicehub.WithTaskName("error metadata processor"))
	}

	p.stats = sharedStatistics

	// add consumer task
	for i := 0; i < p.Cfg.Parallelism; i++ {
		//spot
		ctx.AddTask(func(ctx context.Context) error {
			r, err := p.Kafka.NewBatchReader(&p.Cfg.Input, kafka.WithReaderDecoder(p.decodeError))
			if err != nil {
				return err
			}
			defer r.Close()

			w, err := p.StorageWriter.NewWriter(ctx)
			if err != nil {
				return err
			}
			defer w.Close()
			return storekit.BatchConsume(ctx, r, w, &storekit.BatchConsumeOptions{
				BufferSize:          p.Cfg.BufferSize,
				ReadTimeout:         p.Cfg.ReadTimeout,
				ReadErrorHandler:    p.handleReadError,
				WriteErrorHandler:   p.handleWriteError,
				ConfirmErrorHandler: p.confirmErrorHandler,
				Statistics:          p.stats,
			})
		}, servicehub.WithTaskName(fmt.Sprintf("spot-error-consumer(%d)", i)))
	}
	return nil
}

func init() {
	servicehub.Register("error-persist", &servicehub.Spec{
		ConfigFunc: func() interface{} { return &config{} },
		Creator: func() servicehub.Provider {
			return &provider{}
		},
	})
}
