package models

import (
	"github.com/opentracing/opentracing-go"
	"github.com/uber/jaeger-client-go"
	"github.com/wangjia184/sortedset"
	"golang.org/x/time/rate"
	"goldclient/debug"
	"sync"
	"time"
)

type CounterMapSafe struct {
	sync.Mutex
	Internal map[int]int
}

type ExploreInfo struct {
	TotalExplored int
	TotalTime     time.Duration
	TotalCost     int
	TotalFound    int
}

type ExploreRecord struct {
	Area   int
	Width  int
	Height int
	Timing int64
}

func (counterMap *CounterMapSafe) Increment(key int) {
	counterMap.Lock()
	if _, ok := counterMap.Internal[key]; ok {
		counterMap.Internal[key]++
	} else {
		counterMap.Internal[key] = 1
	}
	counterMap.Unlock()
}

type LicenseData struct {
	License int
	Cost    int
}

type World struct {
	ProcessedYArray []int

	FirstLevelAreas *sortedset.SortedSet

	LicenceMutex          sync.RWMutex
	LicensesMutex         []sync.RWMutex
	Licences              []Licence
	TotalLicenseCount     int64
	TotalLicenseCountPaid int64
	TotalLicenseSuperPaid int64
	TotalLicenseCountFree int64
	LicensesChannel       chan LicenseData
	LicenseRequestsCount  int32
	LicensesFromDig       int
	LicensesFromCash      int
	LicensesDigByCost     map[int]int

	SingleCellExploreCount   int
	SingleCellExploreSuccess int

	Coins []int

	TimeSpentInExplore      time.Duration
	TimeSpentInDig          time.Duration
	TimeSpentInLicenses     time.Duration
	TimeSpentInExploreFull  time.Duration
	TimeSpentInDigFull      time.Duration
	TimeSpentInLicensesFull time.Duration
	TimeSpentInCash         time.Duration
	TimeSpentInCashFull     time.Duration

	ExploresStat     map[int]*ExploreInfo
	ExploresStatFull map[int][]ExploreRecord

	DiggedTreasureCount int
	CashedTreasureCount int64

	DigCounts      []int64
	DigTimingsSum  []time.Duration
	DigCountsFull  []int64
	DigTimingsFull []time.Duration
	Dig400Counts   []int
	Dig400Timings  []int64
	DigCost        []float32

	// second index: 0 = dig, 1 = explore, 2 = cash
	TotalCost [][]float32

	CashSuccessCounts       []int
	CashCounts              []int
	CashSum                 []int
	Cash500Counts           []int
	Cash500Timings          []int64
	CashTimings             []time.Duration
	CashByLicensePriceCount map[int]int
	CashByLicensePriceSum   map[int]int

	CoinErrors    int
	LicenseErrors int

	ExploresThreads   int
	DigThreads        int
	PostCashTimerMS   int
	LicensesThreshold int32

	FirstBreakDown     *sync.WaitGroup
	FirstBreakDownUsed bool

	ResponseCodes CounterMapSafe

	CashPriorityQueue PriorityQueue

	ExploreResponseStat debug.ResponseStatExplore
	CashResponseStat    debug.ResponseStatCash
	LicenseResponseStat debug.ResponseStat
	DigResponseStat     debug.ResponseStatDig

	StartTime time.Time

	ExploreLimiter *rate.Limiter
	DigLimiter     *rate.Limiter

	Tracer opentracing.Tracer
	Spans  []*jaeger.Span
}

func (w *World) InitWorld() {
	w.TotalCost = make([][]float32, 10)
	for i := 0; i < 10; i++ {
		w.TotalCost[i] = make([]float32, 10)
	}

	w.ProcessedYArray = make([]int, 3500)
	w.FirstLevelAreas = sortedset.New()

	w.Licences = make([]Licence, 0)
	w.LicensesMutex = make([]sync.RWMutex, 10)
	w.LicensesChannel = make(chan LicenseData, 1000)
	w.LicensesDigByCost = make(map[int]int, 0)

	w.Coins = make([]int, 0, 10000)

	w.DigCounts = make([]int64, 10)
	w.DigTimingsSum = make([]time.Duration, 10)
	w.DigCountsFull = make([]int64, 10)
	w.DigTimingsFull = make([]time.Duration, 10)
	w.Dig400Timings = make([]int64, 10)
	w.Dig400Counts = make([]int, 10)
	w.DigCost = make([]float32, 10)

	w.CashCounts = make([]int, 10)
	w.CashSuccessCounts = make([]int, 10)
	w.CashSum = make([]int, 10)
	w.Cash500Counts = make([]int, 10)
	w.Cash500Timings = make([]int64, 10)
	w.CashTimings = make([]time.Duration, 10)
	w.CashByLicensePriceCount = make(map[int]int, 0)
	w.CashByLicensePriceSum = make(map[int]int, 0)

	w.ExploresStat = make(map[int]*ExploreInfo)
	w.ExploresStatFull = make(map[int][]ExploreRecord)

	w.ResponseCodes.Internal = make(map[int]int)

	w.CashPriorityQueue = make(PriorityQueue, 0)

	w.FirstBreakDownUsed = false
	w.FirstBreakDown = &sync.WaitGroup{}

	w.LicensesThreshold = 30
	w.ExploresThreads = 80
	w.DigThreads = 10
	w.PostCashTimerMS = 27

	w.ExploreResponseStat = debug.ResponseStatExplore{
		ResponseStat: debug.ResponseStat{},
		Entries:      make([]debug.ExploreStatEntry, 0),
	}
	w.CashResponseStat = debug.ResponseStatCash{
		ResponseStat:       debug.ResponseStat{},
		MaxCash:            0,
		MaxCashLicenseCost: 0,
		MaxCashDepth:       0,
		Treasures:          make([]string, 0),
	}
	w.DigResponseStat = debug.ResponseStatDig{
		ResponseStat: debug.ResponseStat{},
		Entries:      make([]debug.DigStatEntry, 0),
	}

	//w.ExploreLimiter = rate.NewLimiter(500, 5)
	//w.DigLimiter = rate.NewLimiter(400, 10)
	w.ExploreLimiter = rate.NewLimiter(20000, 200)
	w.DigLimiter = rate.NewLimiter(4000000000, 10)
}
