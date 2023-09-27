package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/getaddrinfo/proxy-fingerprint-scraper/api"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/common"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/database"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/fingerprints"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy/impls/saturable"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy/ip"
	"github.com/getaddrinfo/proxy-fingerprint-scraper/proxy/ua"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var db *database.Database = nil

var Debug = flag.Bool("debug", false, "enables debug mode")
var NumWorkers = flag.Int("workers", 1, "number of concurrent workers the app should use")
var FeatureFetchNewFingerprints = flag.Bool("fingerprints", false, "fetch new fingerprints")
var Port = flag.Int("port", 48832, "what port to listen on")

var UserAgentSource = flag.String("ua", "file", "what source to load user agents from (currently only 'file')")
var ProxySoure = flag.String("ip", "file", "what source to load proxy ip:port from (currently only 'file')")

var proxyManager proxy.Manager

// makes sure that the program keeps running
// until it receives an exit sig (one of SIGINT, SIGTERM, SIGHUP for unix)
// os.Interrupt will use platform relevant signals
func preserve(cancel context.CancelFunc) {
	exitSignal := make(chan os.Signal, 1)

	wg := new(sync.WaitGroup)
	wg.Add(1)
	signal.Notify(exitSignal, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

	go func() {
		sig := <-exitSignal

		zap.L().Named("preserver").Info("exiting", zap.String("signal", sig.String()))
		cancel() // cancel the context
		wg.Done()
	}()

	wg.Wait()
	os.Exit(0)
}

func init() {
	flag.Parse()

	config := zap.NewDevelopmentConfig()
	config.DisableStacktrace = true
	config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	config.EncoderConfig.EncodeCaller = nil
	config.EncoderConfig.TimeKey = "ts"

	if *Debug {
		config.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
		config.EncoderConfig.EncodeName = func(name string, prim zapcore.PrimitiveArrayEncoder) {
			prim.AppendString("[" + name + "]")
		}
	} else {
		config.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}

	logger, _ := config.Build()
	zap.ReplaceGlobals(logger)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	if *NumWorkers < 1 && *FeatureFetchNewFingerprints {
		zap.L().Fatal("number of workers must be at least 1")
		os.Exit(1)
	}

	if *Port < 0 || *Port > 65535 {
		zap.L().Fatal("port must be in 0..65535")
		os.Exit(1)
	}

	if *Debug {
		zap.L().Warn("debug mode: do not use this in production.")
	}

	dbUrl, present := os.LookupEnv("DATABASE_URL")

	if !present {
		zap.L().Fatal("DATABASE_URL env variable must be supplied")
		os.Exit(1)
	}

	zap.L().Info("starting")

	localDb, err := database.NewDatabase(ctx, dbUrl)

	if err != nil {
		zap.S().Errorf("could not open db connection: %s", err.Error())
	}

	db = localDb

	svr := api.NewServer(db, proxyManager, *Port)
	svr.InitRoutes()
	go svr.Run()

	if *FeatureFetchNewFingerprints {
		StartFingerprintFetcher(ctx)
	}

	preserve(cancel)
}

// TODO: Move to a map[string]Factory for uaSource and ipSource
// likely taking a function to config
func StartFingerprintFetcher(ctx context.Context) {
	zap.S().Named("fingerprint").Info("feature enabled")

	var ipSource ip.Source

	if !ip.IsValidSource(*ProxySoure) {
		zap.S().Fatalf("invalid proxy source, permitted: %s", strings.Join(ip.ValidSources(), ", "))
		os.Exit(1)
	}

	if *ProxySoure == "file" {
		ipSource = ip.NewFileSystemSource("proxies.txt")
	}

	var uaSource ua.Source

	if !ua.IsValidSource(*UserAgentSource) {
		zap.S().Fatalf("invalid user agent source, permitted: %s", strings.Join(ua.ValidSources(), ", "))
		os.Exit(1)
	}

	if *UserAgentSource == "file" {
		uaSource = ua.NewFileSystemSource("ua.txt")
	}

	proxies := saturable.NewSaturableProxyManager(
		ctx,
		uaSource,
		ipSource,
	)

	zap.S().Named("fingerprint").Infof("loaded %d proxies", len(proxies.IPs()))

	proxyManager = proxies
	results := make(common.FingerprintResultChannel)

	go db.ListenForNewFingerprints(results)

	zap.S().Infof("starting %d workers", *NumWorkers)
	for i := 0; i < *NumWorkers; i++ {
		worker := fingerprints.NewWorker(fingerprints.NewWorkerOptions{
			Id:           i,
			ProxyManager: proxies,
			Context:      ctx,
			Database:     db,
			Results:      results,
		})

		go worker.Run()
	}
}
