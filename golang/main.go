package main

import (
	"fmt"
	"goldclient/httpclient"
	"goldclient/logic"
	"goldclient/models"
	"math/rand"
	"os"
	"runtime"
	"sync"
	"time"
)

func main() {
	fmt.Fprintf(os.Stderr, "Base. with depth to skip calculator. 1775 init explore. Skip calculator normal \n")

	world := &models.World{}
	world.InitWorld()

	seed := time.Now().UnixNano()
	rand.Seed(seed)

	address := os.Getenv("ADDRESS")
	if address != "localhost" && address != "127.0.0.1" {
		time.Sleep(10)
	}
	httpclient.Create("http://"+address+":8000", world, address)

	start := time.Now()

	ticker := time.NewTicker(600 * time.Second)
	world.StartTime = time.Now()

	go logic.SetupLicences(world)

	licenseTicker := time.NewTicker(7 * time.Millisecond)
	go func() {
		for {
			select {
			case <-licenseTicker.C:
				go logic.SendLicenseIfNeeded(world)
			}
		}
	}()

	for i := 0; i < world.ExploresThreads; i++ {
		go logic.Explore(world, i, logic.EXPLORE_SIZE)
		//time.Sleep(200 * time.Millisecond)
	}

	go func() {
		for {
			select {
			case <-ticker.C:
				fmt.Fprintf(os.Stderr, "single cell explore: %d \n", world.SingleCellExploreCount)
				fmt.Fprintf(os.Stderr, "single cell explore success: %d \n", world.SingleCellExploreSuccess)
				fmt.Fprintf(os.Stderr, "Time spend in explore: %d \n", world.TimeSpentInExplore.Milliseconds())
				fmt.Fprintf(os.Stderr, "Time spend in explore full: %d \n", world.TimeSpentInExploreFull.Milliseconds())
				fmt.Fprintf(os.Stderr, "Time spend in dig: %d \n", world.TimeSpentInDig.Milliseconds())
				fmt.Fprintf(os.Stderr, "Time spend in dig full: %d \n", world.TimeSpentInDigFull.Milliseconds())
				fmt.Fprintf(os.Stderr, "Time spend in licenses: %d \n", world.TimeSpentInLicenses.Milliseconds())
				fmt.Fprintf(os.Stderr, "Time spend in licenses full: %d \n", world.TimeSpentInLicensesFull.Milliseconds())
				fmt.Fprintf(os.Stderr, "Time spend in cash: %d \n", world.TimeSpentInCash.Milliseconds())
				fmt.Fprintf(os.Stderr, "Time spend in cash full: %d \n", world.TimeSpentInCashFull.Milliseconds())

				fmt.Fprintf(os.Stderr, "Digged treasure count: %d \n", world.DiggedTreasureCount)
				fmt.Fprintf(os.Stderr, "Cashed treasure count: %d \n", world.CashedTreasureCount)

				fmt.Fprintf(os.Stderr, "Number of goroutines: %d \n", runtime.NumGoroutine())

				totalDig := int64(0)
				totalDigCost := float32(0)
				for i := 0; i < 10; i++ {
					fmt.Fprintf(os.Stderr, "Digs in depth (depth, count, average micro, cost): %d %d %f %f \n", i+1, world.DigCounts[i], float64(world.DigTimingsSum[i].Microseconds())/float64(world.DigCounts[i]), float64(world.DigCost[i])/float64(world.DigCounts[i]))
					fmt.Fprintf(os.Stderr, "Digs total full (depth, count, total, average micro): %d %d %d %f \n", i+1, world.DigCountsFull[i], world.DigTimingsFull[i], float64(world.DigTimingsFull[i].Microseconds())/float64(world.DigCountsFull[i]))
					totalDig += world.DigCounts[i]
					totalDigCost += world.DigCost[i]
				}

				totalExplore := 0
				totalExploreCost := 0
				//fmt.Fprintf(os.Stderr, "EXPLORE: \n")
				for key, element := range world.ExploresStat {
					//fmt.Fprintf(os.Stderr, "%d %f;", key, float64(element.TotalTime)/float64(element.TotalExplored))
					fmt.Fprintf(os.Stderr, "Explore for (area, count, totalFound, averageFound, average micro, cost): %d %d %d %f %f %f \n", key, element.TotalExplored, element.TotalFound, float64(element.TotalFound)/float64(element.TotalExplored), float64(element.TotalTime.Microseconds())/float64(element.TotalExplored), float64(element.TotalCost)/float64(element.TotalExplored))
					totalExplore += element.TotalExplored
					totalExploreCost += element.TotalCost
				}

				totalCash := 0
				successCash := 0
				for i := 0; i < 10; i++ {
					fmt.Fprintf(os.Stderr, "Cash in level (total, average cash, average elapsed): %d %f %f \n", world.CashSuccessCounts[i], float64(world.CashSum[i])/float64(world.CashSuccessCounts[i]), float64(world.CashTimings[i].Milliseconds())/float64(world.CashSuccessCounts[i]))
					totalCash += world.CashCounts[i]
					successCash += world.CashSuccessCounts[i]
				}

				fmt.Fprintf(os.Stderr, "Cost in 0.1s. Dig = %f, explore = %f, cash = %f, total = %f \n", world.TotalCost[0][0], world.TotalCost[0][1], world.TotalCost[0][2], world.TotalCost[0][0] + world.TotalCost[0][1] + world.TotalCost[0][2])
				fmt.Fprintf(os.Stderr, "Cost in 0.2s. Dig = %f, explore = %f, cash = %f, total = %f \n", world.TotalCost[1][0], world.TotalCost[1][1], world.TotalCost[1][2], world.TotalCost[1][0] + world.TotalCost[1][1] + world.TotalCost[1][2])
				fmt.Fprintf(os.Stderr, "Cost in 0.5s. Dig = %f, explore = %f, cash = %f, total = %f \n", world.TotalCost[2][0], world.TotalCost[2][1], world.TotalCost[2][2], world.TotalCost[2][0] + world.TotalCost[2][1] + world.TotalCost[2][2])
				fmt.Fprintf(os.Stderr, "Cost in 1s. Dig = %f, explore = %f, cash = %f, total = %f \n", world.TotalCost[3][0], world.TotalCost[3][1], world.TotalCost[3][2], world.TotalCost[3][0] + world.TotalCost[3][1] + world.TotalCost[3][2])
				fmt.Fprintf(os.Stderr, "Cost in 1 - 2s. Dig = %f, explore = %f, cash = %f, total = %f \n", world.TotalCost[4][0], world.TotalCost[4][1], world.TotalCost[4][2], world.TotalCost[4][0] + world.TotalCost[4][1] + world.TotalCost[4][2])

				fmt.Fprintf(os.Stderr, "Total (dig, explore, cash, lic): %d %d %d %d \n", totalDig, totalExplore, totalCash, world.TotalLicenseCount)
				fmt.Fprintf(os.Stderr, "RPS (dig, explore, cash, lic, all): %d %d %d %d %d \n", totalDig/600, totalExplore/600, totalCash/600, world.TotalLicenseCount/600, totalDig/int64(600)+int64(totalExplore/600)+int64(totalCash/600)+world.TotalLicenseCount/600)
				fmt.Fprintf(os.Stderr, "average cost per sec(dig, explore, cash, all): %f %f %f %f \n", totalDigCost/600.0, float32(totalExploreCost)/600.0, float32(successCash)*20/600.0, totalDigCost/600+float32(totalExploreCost)/600+float32(successCash)*20/600)
				fmt.Fprintf(os.Stderr, "Average (dig, explore, cash): %f %f %f \n", float64(world.TimeSpentInDig.Milliseconds())/float64(totalDig), float64(world.TimeSpentInExplore.Milliseconds())/float64(totalExplore), float64(world.TimeSpentInCash.Milliseconds())/float64(successCash))
				fmt.Fprintf(os.Stderr, "Total licenses count (total, paid, superpaid, free): %d %d %d %d \n", world.TotalLicenseCount, world.TotalLicenseCountPaid, world.TotalLicenseSuperPaid, world.TotalLicenseCountFree)
				fmt.Fprintf(os.Stderr, "Licenses from (cash, dig): %d %d \n", world.LicensesFromCash, world.LicensesFromDig)
				fmt.Fprintf(os.Stderr, "configuration (explorer threads, post cash timer, licethreshold): %d %d %d \n", world.ExploresThreads, world.PostCashTimerMS, world.LicensesThreshold)
				fmt.Fprintf(os.Stderr, "configuration (explore, dig limiters): %f %f \n", world.ExploreLimiter.Limit(), world.DigLimiter.Limit())
			}
		}
	}()

	var wg = &sync.WaitGroup{}
	wg.Add(1)
	wg.Wait()
	t := time.Now()
	elapsed := t.Sub(start)
	fmt.Println(start)
	fmt.Println(t)
	fmt.Println("elapsed: " + elapsed.String())
	fmt.Println(elapsed)
}
