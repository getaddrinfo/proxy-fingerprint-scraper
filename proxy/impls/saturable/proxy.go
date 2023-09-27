package saturable

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy"
	"go.uber.org/zap"
)

const GenerationPerTwoMinuteLimit = 3
const FailedRequestImpliesDownLimit = 2

type SaturableProxy struct {
	sync.Mutex

	ip        string
	client    *http.Client
	userAgent string

	manager *SaturableProxyManager

	timesUsed     uint8
	timesUsedChan chan struct{}

	timesFailed     uint8
	timesFailedChan chan struct{}
}

func (p *SaturableProxy) IP() string {
	return p.ip
}

func (p *SaturableProxy) Client() *http.Client {
	return p.client
}

func (p *SaturableProxy) Do(r *http.Request) (*http.Response, error) {
	r.Header.Set("User-Agent", p.userAgent)

	res, err := p.Client().Do(r)

	if err != nil {
		p.incrementFail()
	}

	return res, err
}

func (p *SaturableProxy) incrementUsage() {
	p.Lock()
	defer p.Unlock()

	if p.timesUsed >= GenerationPerTwoMinuteLimit {
		panic("invariant: proxy should not be used if it is saturated")
	}

	p.timesUsed = p.timesUsed + 1

	// the first time it has been used, notify timesUsedChan to start a timer
	// and eventually reset it
	if p.timesUsed == 1 {
		p.timesUsedChan <- struct{}{}
	}

	if p.timesUsed == GenerationPerTwoMinuteLimit {
		p.manager.taintSaturated(p)
	}
}

func (p *SaturableProxy) incrementFail() {
	if p.timesFailedChan == nil {
		panic("proxy is not running times failed loop")
	}

	p.Lock()
	defer p.Unlock()

	p.timesFailed = p.timesFailed + 1

	if p.timesFailed == FailedRequestImpliesDownLimit {
		zap.L().Warn("proxy may be unhealthy, ignoring for now", zap.String("proxy", p.ip))
		p.timesFailedChan <- struct{}{}
		p.manager.taintUnhealthy(p)
	}
}

// TODO: is it acceptable to leak the cancel here? we know the parent context will eventually be cancelled...
func (p *SaturableProxy) startChannels(ctx context.Context) {
	usedCtx, _ := context.WithCancel(ctx)
	go p.runTimesUsedResetter(usedCtx)

	failedCtx, _ := context.WithCancel(ctx)
	go p.runTimesFailedResetter(failedCtx)
}

func (p *SaturableProxy) runTimesUsedResetter(ctx context.Context) {
Iter:
	for {
		select {
		case <-ctx.Done():
			break Iter

		case <-p.timesUsedChan:
			<-time.NewTimer(time.Minute*3 + time.Minute*20).C
			p.Lock()
			p.timesUsed = 0
			p.manager.freeSaturated(p)
			p.Unlock()
		}
	}
}

func (p *SaturableProxy) runTimesFailedResetter(ctx context.Context) {
Iter:
	for {
		select {
		case <-ctx.Done():
			break Iter

		case <-p.timesFailedChan:
			<-time.NewTimer(time.Minute * 10).C
			p.Lock()
			p.timesFailed = 0
			p.manager.freeUnhealthy(p)
			p.Unlock()
		}
	}
}

// assert proxy.Proxy interface is implemented by SaturableProxy
var _ proxy.Proxy = (*SaturableProxy)(nil)
