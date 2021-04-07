package debug

import (
	"fmt"
	"goldclient/util"
	"net/http"
	"os"
	"sort"
	"time"
)

type StatEntry struct {
	StartedAt time.Duration
	Duration  time.Duration
}

type ExploreStatEntry struct {
	StatEntry
	Area int
}

type ResponseStat struct {
	TotalCount             int
	MinResponseTimeMs      int64
	MinRequestFromStart    time.Duration
	MinRequestRequest      http.Request
	RequestsCountBeforeMin int
	TotalElapsedMs         int64
}

type ResponseStatExplore struct {
	ResponseStat
	Entries []ExploreStatEntry
}

type DigStatEntry struct {
	StatEntry
	Depth int
}

type ResponseStatDig struct {
	ResponseStat
	Entries []DigStatEntry
}

type ResponseStatCash struct {
	ResponseStat
	MaxCash            int
	MaxCashLicenseCost int
	MaxCashDepth       int
	Treasures          []string
}

func (r *ResponseStat) ProcessNew(elapsed time.Duration, req http.Request, timeFromStart time.Duration) {
	r.TotalCount++
	if elapsed.Microseconds() < r.MinResponseTimeMs || r.MinResponseTimeMs == 0 {
		r.MinResponseTimeMs = elapsed.Microseconds()
		r.MinRequestFromStart = timeFromStart
		r.MinRequestRequest = req
		r.RequestsCountBeforeMin = r.TotalCount - 1
	}
	r.TotalElapsedMs += elapsed.Microseconds()
}

func (r *ResponseStatExplore) ProcessNew(elapsed time.Duration, req http.Request, timeFromStart time.Duration, area int) {
	r.ResponseStat.ProcessNew(elapsed, req, timeFromStart)
	r.Entries = append(r.Entries, ExploreStatEntry{
		StatEntry: StatEntry{
			Duration:  elapsed,
			StartedAt: timeFromStart,
		},
		Area: area,
	})
}

func (r *ResponseStatDig) ProcessNew(elapsed time.Duration, req http.Request, timeFromStart time.Duration, depth int) {
	r.ResponseStat.ProcessNew(elapsed, req, timeFromStart)
	r.Entries = append(r.Entries, DigStatEntry{
		StatEntry: StatEntry{
			Duration:  elapsed,
			StartedAt: timeFromStart,
		},
		Depth: depth,
	})
}

func (r *ResponseStatCash) ProcessNew(elapsed time.Duration, req http.Request, timeFromStart time.Duration, cash, licenseCost, depth int, treasure string) {
	r.ResponseStat.ProcessNew(elapsed, req, timeFromStart)
	if cash > r.MaxCash {
		r.MaxCash = cash
		r.MaxCashLicenseCost = licenseCost
		r.MaxCashDepth = depth
	}
	r.Treasures = append(r.Treasures, treasure)
}

func (r *ResponseStat) Print() {
	fmt.Fprintf(os.Stderr, "timings (average, min): %f %d \n", float64(r.TotalElapsedMs)/float64(r.TotalCount)/1000, r.MinResponseTimeMs)
	fmt.Fprintf(os.Stderr, "min data (req from start, count before min): %f %d \n", r.MinRequestFromStart.Seconds(), r.RequestsCountBeforeMin)
}

func (r *ResponseStatDig) Print() {
	r.ResponseStat.Print()

	sortedByDuration := make([]DigStatEntry, len(r.Entries))
	copy(sortedByDuration, r.Entries)
	sort.Slice(sortedByDuration, func(i, j int) bool {
		return sortedByDuration[i].Duration.Microseconds() < sortedByDuration[j].Duration.Microseconds()
	})

	for i := 0; i < 10; i++ {
		sumInRange := int64(0)
		countInRange := 0
		startDate := sortedByDuration[i].StartedAt - 1*time.Second
		endDate := sortedByDuration[i].StartedAt + 1*time.Second

		lastSec := sortedByDuration[i].StartedAt - 1*time.Second
		sumInLastSecond := int64(0)
		countInLastSecond := 0
		costInLastSecond := 0
		for j := 0; j < len(r.Entries); j++ {
			if r.Entries[j].StartedAt > startDate && r.Entries[j].StartedAt < endDate {
				sumInRange += r.Entries[j].Duration.Microseconds()
				countInRange++
			}
			if r.Entries[j].StartedAt > lastSec && r.Entries[j].StartedAt < sortedByDuration[i].StartedAt {
				sumInLastSecond += r.Entries[j].Duration.Microseconds()
				countInLastSecond++
				costInLastSecond += util.GetCostByArea(r.Entries[j].Depth)
			}
		}

		fmt.Fprintf(os.Stderr, "min (index, min, avg, startedSec) %d %d %f %f \n", i, sortedByDuration[i].Duration.Microseconds(), float64(sumInRange)/float64(countInRange), sortedByDuration[i].StartedAt.Seconds())
		fmt.Fprintf(os.Stderr, "min in last second (index, avg, count, cost) %d %f %d %d \n", i, float64(sumInLastSecond)/float64(countInLastSecond), countInLastSecond, costInLastSecond)
	}
}

func (r *ResponseStatExplore) Print() {
	r.ResponseStat.Print()
	sortedByDuration := make([]ExploreStatEntry, len(r.Entries))
	copy(sortedByDuration, r.Entries)
	sort.Slice(sortedByDuration, func(i, j int) bool {
		return sortedByDuration[i].Duration.Microseconds() < sortedByDuration[j].Duration.Microseconds()
	})

	for i := 0; i < 10; i++ {
		sumInRange := int64(0)
		countInRange := 0
		startDate := sortedByDuration[i].StartedAt - 1*time.Second
		endDate := sortedByDuration[i].StartedAt + 1*time.Second

		lastSec := sortedByDuration[i].StartedAt - 1*time.Second
		sumInLastSecond := int64(0)
		countInLastSecond := 0
		costInLastSecond := 0
		for j := 0; j < len(r.Entries); j++ {
			//if r.Entries[j].Area >= 4 {
			//	continue
			//}

			if r.Entries[j].StartedAt > startDate && r.Entries[j].StartedAt < endDate {
				sumInRange += r.Entries[j].Duration.Microseconds()
				countInRange++
			}
			if r.Entries[j].StartedAt > lastSec && r.Entries[j].StartedAt < sortedByDuration[i].StartedAt {
				sumInLastSecond += r.Entries[j].Duration.Microseconds()
				countInLastSecond++
				costInLastSecond += util.GetCostByArea(r.Entries[j].Area)
			}
		}

		fmt.Fprintf(os.Stderr, "min (index, min, avg, startedSec) %d %d %f %f \n", i, sortedByDuration[i].Duration.Microseconds(), float64(sumInRange)/float64(countInRange), sortedByDuration[i].StartedAt.Seconds())
		fmt.Fprintf(os.Stderr, "min in last second (index, avg, count, cost) %d %f %d %d \n", i, float64(sumInLastSecond)/float64(countInLastSecond), countInLastSecond, costInLastSecond)
	}
}

func (r *ResponseStatCash) Print() {
	r.ResponseStat.Print()
	fmt.Fprintf(os.Stderr, "Max cash (max, license cost, depth): %d %d %d \n", r.MaxCash, r.MaxCashLicenseCost, r.MaxCashDepth)
	fmt.Fprintf(os.Stderr, "Treasure (len, len uniq): %d %d \n", len(r.Treasures), len(uniqueNonEmptyElementsOf(r.Treasures)))

	for i := 0; i < 10; i++ {
		fmt.Fprintf(os.Stderr, "%s\n", r.Treasures[i])
	}
}

func uniqueNonEmptyElementsOf(s []string) []string {
	unique := make(map[string]bool, len(s))
	us := make([]string, len(unique))
	for _, elem := range s {
		if len(elem) != 0 {
			if !unique[elem] {
				us = append(us, elem)
				unique[elem] = true
			}
		}
	}

	return us

}
