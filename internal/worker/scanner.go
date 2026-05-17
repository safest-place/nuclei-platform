package worker

import (
	"context"

	nuclei "github.com/projectdiscovery/nuclei/v3/lib"
	"github.com/projectdiscovery/nuclei/v3/pkg/output"
	"github.com/rs/zerolog"
)

type Scanner struct {
	engine *nuclei.ThreadSafeNucleiEngine
	log    *zerolog.Logger
}

func NewScanner(ctx context.Context, log *zerolog.Logger, opts ...nuclei.NucleiSDKOptions) (*Scanner, error) {
	engine, err := nuclei.NewThreadSafeNucleiEngineCtx(ctx, opts...)
	if err != nil {
		return nil, err
	}
	log.Info().Msg("loading nuclei templates...")
	if err := engine.GlobalLoadAllTemplates(); err != nil {
		engine.Close()
		return nil, err
	}
	log.Info().Msg("nuclei templates loaded")
	return &Scanner{engine: engine, log: log}, nil
}

func (s *Scanner) SetResultCallback(callback func(event *output.ResultEvent)) {
	s.engine.GlobalResultCallback(callback)
}

func (s *Scanner) Execute(ctx context.Context, targets []string, opts ...nuclei.NucleiSDKOptions) error {
	return s.engine.ExecuteNucleiWithOptsCtx(ctx, targets, opts...)
}

func (s *Scanner) Close() {
	s.engine.Close()
}
