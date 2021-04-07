package httpclient

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-cleanhttp"
	"golang.org/x/time/rate"
	"goldclient/models"
	"goldclient/util"
	"net/http"
	"os"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

type Client struct {
	client         http.Client
	baseUrl        string
	rateLimiter    *rate.Limiter
	rateLimiterDig *rate.Limiter
	ctx            context.Context
	hostname       string
	customClient   CustomClient
}

var client Client

const MAX_RETRY = 10

func PostCash(treasure string, w *models.World, depth int, licenseData models.LicenseData) int {
	startFull := time.Now()
	defer func() {
		w.TimeSpentInCashFull += time.Since(startFull)
	}()

	var res []int
	for i := 0; i < MAX_RETRY; i++ {
		start := time.Now()
		res = client.customClient.RawCash(client.hostname, "/cash", "\""+treasure+"\"")

		w.CashCounts[depth-1]++
		w.CashTimings[depth-1] += time.Since(start)
		w.TimeSpentInCash += time.Since(start)

		if res != nil {
			w.CashSuccessCounts[depth-1]++
			break
		}
		w.Cash500Counts[depth-1]++
		w.Cash500Timings[depth-1] += time.Since(start).Milliseconds()
		i++
	}

	w.CashedTreasureCount++
	w.CashSum[depth-1] += len(res)
	return len(res)
}

func PostDig(dig models.Dig, w *models.World, licenseData models.LicenseData) []string {
	startFull := time.Now()
	defer func() {
		w.TimeSpentInDigFull += time.Since(startFull)
	}()

	postData := fmt.Sprintf("{\"licenseID\":%d,\"posX\":%d,\"posY\":%d,\"depth\":%d}", dig.LicenseID, dig.PosX, dig.PosY, dig.Depth)
	start := time.Now()

	result := client.customClient.RawDig(client.hostname, "/dig", postData, dig.Depth)
	w.DigCounts[dig.Depth-1]++
	w.DigTimingsSum[dig.Depth-1] += time.Since(start)
	w.DigCost[dig.Depth-1] += util.GetCostByDepth(dig.Depth)
	w.TimeSpentInDig += time.Since(start)
	return result
}

func PostLicence(ch chan models.LicenseData, w *models.World, coin int) models.Licence {
	startFull := time.Now()
	defer func() {
		w.TimeSpentInLicensesFull += time.Since(startFull)
	}()
	licenseRequest := "[]"
	cost := 0
	start := time.Now()

	var response models.Licence
	for i := 0; i < 3; i++ {
		response = client.customClient.RawLicense(client.hostname, "/licenses", licenseRequest)
		if response.DigAllowed > 0 {
			break
		}
	}

	w.TimeSpentInLicenses += time.Since(start)

	for i := 0; i < response.DigAllowed; i++ {
		ch <- models.LicenseData{
			License: response.Id,
			Cost:    cost,
		}
	}
	return response
}

func SimplePostExplore(area models.Area, world *models.World) models.AreaResponse {
	startFull := time.Now()
	defer func() {
		world.TimeSpentInExploreFull += time.Since(startFull)
	}()

	postData := fmt.Sprintf("{\"posX\":%d,\"posY\":%d,\"sizeX\":%d,\"sizeY\":%d}", area.PosX, area.PosY, area.SizeX, area.SizeY)

	start := time.Now()

	areaResponse := client.customClient.RawExplore(client.hostname, "/explore", postData)
	world.TimeSpentInExplore += time.Since(start)

	return areaResponse
}

func Create(baseUrl string, w *models.World, hostname string) {
	retryableClient := retryablehttp.NewClient()
	retryableClient.Backoff = NoBackoff
	retryableClient.Logger = nil
	retryableClient.RetryMax = 10
	retryableClient.HTTPClient = cleanhttp.DefaultClient()

	defaultRoundTripper := http.DefaultTransport
	defaultTransportPointer, _ := defaultRoundTripper.(*http.Transport)
	defaultTransport := *defaultTransportPointer
	defaultTransport.MaxIdleConns = 20000
	defaultTransport.MaxIdleConnsPerHost = 20000

	client = Client{
		client: http.Client{
			Transport: &defaultTransport,
		},
		baseUrl:        baseUrl,
		rateLimiter:    w.ExploreLimiter,
		rateLimiterDig: w.DigLimiter,
		ctx:            context.Background(),
		hostname:       hostname,
		customClient:   NewCustomClient(hostname),
	}
}

func NoBackoff(min, max time.Duration, attemptNum int, resp *http.Response) time.Duration {
	if resp.StatusCode < 500 {
		fmt.Fprintf(os.Stderr, "error retry: %d \n", resp.StatusCode)
	}
	return 0
}
