package fingerprints

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/common"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/database"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy"
	"go.uber.org/zap"
)

type Worker struct {
	ProxyHandler proxy.Manager
	WorkerID     int
	Log          *zap.Logger
	ctx          context.Context
	Results      common.FingerprintResultChannel
}

type NewWorkerOptions struct {
	Id           int
	ProxyManager proxy.Manager
	Context      context.Context
	Database     *database.Database
	Results      common.FingerprintResultChannel
}

func NewWorker(options NewWorkerOptions) Worker {
	ctx, _ := context.WithCancel(options.Context)

	return Worker{
		ProxyHandler: options.ProxyManager,
		WorkerID:     options.Id,
		Log:          zap.L().Named(fmt.Sprintf("fingerprint.worker(id=%d)", options.Id)),
		ctx:          ctx,
		Results:      options.Results,
	}
}

func (w Worker) Run() {
	ticker := time.NewTicker(time.Second * 20)

Iter:
	for {
		select {
		case <-w.ctx.Done():
			w.Log.Info("exiting")
			ticker.Stop()
			break Iter
		case <-ticker.C:
			w.Log.Debug("handling")
			w.Handle()
			w.Log.Debug("handled")
		}
	}
}

func (w Worker) Handle() {
	prox, err := w.ProxyHandler.Get()

	if err != nil && errors.Is(err, proxy.ErrNoneMatch) {
		w.Log.Debug("all proxies unavailable")
		return
	}

	if err != nil {
		w.Log.Sugar().Error(err.Error())
		return
	}

	ip := prox.IP()
	traceId := common.GenerateID("trace")
	logger := w.Log.With(zap.String("trace", traceId), zap.String("proxy", ip))

	req, err := http.NewRequest("GET", "https://discord.com/api/v8/experiments", nil)

	if err != nil {
		logger.Warn("could not create http request", zap.String("error", err.Error()))
		return
	}

	logger.Info("GET /experiments", zap.String("proxy", ip))
	resp, err := prox.Do(req)

	if err != nil {
		logger.Sugar().Errorf("proxy error: %s", err.Error())
		return
	}

	logger.Info("Response received")
	data, err := common.DecodeFingerprintBody(resp)

	if err != nil {
		logger.Sugar().Errorf("json decode error: %s", err.Error())
		return
	}

	if data.Fingerprint == nil {
		logger.Sugar().Errorf("missing fingerprint: %s", string(data.Raw))
		return
	}

	w.Results <- common.FingerprintResult{
		Fingerprint: *data.Fingerprint,
		ProxyIP:     ip,
	}
}
