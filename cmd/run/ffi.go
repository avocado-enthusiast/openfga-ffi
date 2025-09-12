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
var initialized bool

//export Init
func Init(logging bool, fileSuffix int32) int32 {
	mtx.Lock()
	defer mtx.Unlock()

	if !initialized {
		// This entire block is basically what the run() function does, just simplified
		config := serverconfig.DefaultConfig()
		config.UnixSocketFileSuffix = strconv.Itoa(int(fileSuffix))

		if !logging {
			config.Log.Level = "none"
		}
		config.Playground.Enabled = false

		logger := logger.MustNewLogger(config.Log.Format, config.Log.Level, config.Log.TimestampFormat)
		serverCtx := &ServerContext{Logger: logger}

		go func() {
			if err := serverCtx.Run(context.Background(), config); err != nil {
				panic(err)
			}
		}()

		health := checkHealth()

		// Not sure what to do, since in theory things could be in partially initialized state
		initialized = true

		return health
	}

	return 0
}

func checkHealth() int32 {
	// Try for 4 seconds
	maxAttempts := 200
	sleepDuration := 20 * time.Millisecond

	for range maxAttempts {
		if checkUp() == 0 {
			return 0
		}
		time.Sleep(sleepDuration)
	}
	return 30
}

type HealthzResponse struct {
	Status string `json:"status"`
}

func checkUp() int32 {
	httpSocket := os.TempDir() + "openfga-http.sock"

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
