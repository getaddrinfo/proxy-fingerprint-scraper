package saturable

import (
	"context"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"sync"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy/ip"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy/ua"
	"go.uber.org/zap"
)

type SaturableProxyManager struct {
	sync.Mutex

	agentProvider ua.Source
	ipProvider    ip.Source

	Proxies []*SaturableProxy

	saturated map[string]bool
	unhealthy map[string]bool

	ctx context.Context
}

func NewSaturableProxyManager(
	ctx context.Context,
	userAgentProvider ua.Source,
	ipProvider ip.Source,
) proxy.Manager {
	mgr := &SaturableProxyManager{
		Proxies:   []*SaturableProxy{},
		saturated: map[string]bool{},
		unhealthy: map[string]bool{},
		ctx:       ctx,

		agentProvider: userAgentProvider,
		ipProvider:    ipProvider,
	}

	if err := mgr.Init(); err != nil {
		zap.S().Fatal("cannot init manager: %s", err)
	}

	return mgr
}

func (m *SaturableProxyManager) Init() error {
	return m.loadProxies()
}

func (m *SaturableProxyManager) Get() (proxy.Proxy, error) {
	m.Lock()

	var usable []*SaturableProxy

	// nesting here isn't very nice, but oh well
	for _, proxy := range m.Proxies {
		if _, ok := m.saturated[proxy.IP()]; !ok {
			if _, ok := m.unhealthy[proxy.IP()]; !ok {
				usable = append(usable, proxy)
			}
		}
	}

	m.Unlock()

	// if none meet the criteria, all are saturated or unhealthy
	if len(usable) == 0 {
		return nil, proxy.ErrNoneMatch
	}

	// assume if it's being returned, then it is going to be used
	p := usable[rand.Intn(len(usable))]
	p.incrementUsage()

	return p, nil
}

func (m *SaturableProxyManager) loadProxies() error {
	ips, err := m.ipProvider.Load()

	if err != nil {
		return err
	}

	for _, ip := range ips {
		if err := m.Add(ip); err != nil {
			zap.L().Error(
				"failed to add proxy",
				zap.Error(err),
				zap.String("ip", ip),
			)
		}
	}

	return nil
}

func (m *SaturableProxyManager) IPs() []string {
	ips := []string{}

	m.Lock()
	defer m.Unlock()

	for _, proxy := range m.Proxies {
		ips = append(ips, proxy.IP())
	}

	return ips
}

func (m *SaturableProxyManager) Add(ip string) error {
	m.Lock()
	defer m.Unlock()

	parsed, err := url.Parse(fmt.Sprintf("http://%s", ip))

	if err != nil {
		return err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyURL(parsed),
		},
	}

	userAgent, err := m.agentProvider.Random()

	if err != nil {
		return err
	}

	proxy := &SaturableProxy{
		client:          client,
		ip:              ip,
		manager:         m,
		userAgent:       userAgent,
		timesFailedChan: make(chan struct{}),
		timesUsedChan:   make(chan struct{}),
	}

	go proxy.startChannels(m.ctx)

	m.Proxies = append(m.Proxies, proxy)

	zap.L().Debug("added proxy", zap.String("ip", parsed.Host), zap.String("port", parsed.Port()))

	return nil
}

func (m *SaturableProxyManager) taintSaturated(proxy *SaturableProxy) {
	m.Lock()
	m.saturated[proxy.IP()] = true
	m.Unlock()
}

func (m *SaturableProxyManager) freeSaturated(proxy *SaturableProxy) {
	m.Lock()
	delete(m.saturated, proxy.IP())
	m.Unlock()
}

func (m *SaturableProxyManager) taintUnhealthy(proxy *SaturableProxy) {
	m.Lock()
	m.unhealthy[proxy.IP()] = true
	m.Unlock()
}

func (m *SaturableProxyManager) freeUnhealthy(proxy *SaturableProxy) {
	m.Lock()
	delete(m.unhealthy, proxy.IP())
	m.Unlock()
}

var _ proxy.Manager = (*SaturableProxyManager)(nil)
