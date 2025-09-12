package run

import "C"

import (
	"context"
	"encoding/json"
	"net"
	"os"
	"strconv"
	"sync"
	"time"

	"net/http"

	"github.com/openfga/openfga/pkg/logger"
	serverconfig "github.com/openfga/openfga/pkg/server/config"
)

var mtx sync.Mutex
var ctx context.Context
var cancel context.CancelFunc

//export Init
func Init(logging bool, fileSuffix int32) int32 {
	mtx.Lock()
	defer mtx.Unlock()

	// Second call to Init restarts service
	innerStop()

	ctx, cancel = context.WithCancel(context.Background())

	config := serverconfig.DefaultConfig()
	if !logging {
		config.Log.Level = "none"
	}
	config.UnixSocketFileSuffix = strconv.Itoa(int(fileSuffix))
	config.Profiler.Enabled = false
	config.Playground.Enabled = false
	config.Metrics.Enabled = false

	go func() {
		logger := logger.MustNewLogger(config.Log.Format, config.Log.Level, config.Log.TimestampFormat)
		serverCtx := &ServerContext{Logger: logger}
		if err := serverCtx.Run(ctx, config); err != nil {
			innerStop()
			panic(err)
		}
	}()

	health := checkHealth(config.UnixSocketFileSuffix)
	if health != 0 {
		innerStop()
	}
	return health
}

func Stop() {
	mtx.Lock()
	defer mtx.Unlock()
	innerStop()
}

func innerStop() {
	if cancel != nil {
		cancel()
		<-ctx.Done()
		ctx, cancel = nil, nil
	}
}

func checkHealth(suffix string) int32 {
	// Try for 4 seconds
	maxAttempts := 200
	sleepDuration := 20 * time.Millisecond
	httpSocket := os.TempDir() + "openfga-http-" + suffix + ".sock"
	for range maxAttempts {
		if checkUp(httpSocket) == 0 {
			return 0
		}
		time.Sleep(sleepDuration)
	}
	return 30
}

type HealthzResponse struct {
	Status string `json:"status"`
}

func checkUp(httpSocket string) int32 {
	// Create a client
	client := &http.Client{
		Transport: &http.Transport{
			Dial: func(proto, addr string) (net.Conn, error) {
				return net.Dial("unix", httpSocket)
			},
		},
	}

	resp, err := client.Get("http://localhost/healthz")
	if err != nil {
		return 10
	}
	defer resp.Body.Close()

	var healthzResponse HealthzResponse
	err = json.NewDecoder(resp.Body).Decode(&healthzResponse)
	if err != nil {
		return 11
	}
	if healthzResponse.Status != "SERVING" {
		return 12
	}
	return 0
}
