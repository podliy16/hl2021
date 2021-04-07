package models

import (
	"container/heap"
)

// An Item is something we manage in a priority queue.
type Item struct {
	Value    string // The value of the item; arbitrary.
	Depth    int
	Priority int // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	Index       int // The index of the item in the heap.
	LicenseData LicenseData
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].Priority > pq[j].Priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].Index = i
	pq[j].Index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.Index = n
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	old[n-1] = nil  // avoid memory leak
	item.Index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value string, priority int) {
	item.Value = value
	item.Priority = priority
	heap.Fix(pq, item.Index)
}

//THREAD SAFE
type heapPopChanMsg struct {
	h      heap.Interface
	result chan interface{}
}

// heapPushChanMsg - the message structure for a push chan
type heapPushChanMsg struct {
	h heap.Interface
	x interface{}
}

var (
	quitChan chan bool
	// heapPushChan - push channel for pushing to a heap
	heapPushChan = make(chan heapPushChanMsg)
	// heapPopChan - pop channel for popping from a heap
	heapPopChan = make(chan heapPopChanMsg)
)

// HeapPush - safely push item to a heap interface
func HeapPush(h heap.Interface, x interface{}) {
	heapPushChan <- heapPushChanMsg{
		h: h,
		x: x,
	}
}

// HeapPop - safely pop item from a heap interface
func HeapPop(h heap.Interface) interface{} {
	var result = make(chan interface{})
	heapPopChan <- heapPopChanMsg{
		h:      h,
		result: result,
	}
	return <-result
}

//stopWatchHeapOps - stop watching for heap operations
func stopWatchHeapOps() {
	quitChan <- true
}

// watchHeapOps - watch for push/pops to our heap, and serializing the operations
// with channels
func WatchHeapOps() chan bool {
	var quit = make(chan bool)
	go func() {
		for {
			select {
			case <-quit:
				return
			case popMsg := <-heapPopChan:
				popMsg.result <- heap.Pop(popMsg.h)
			case pushMsg := <-heapPushChan:
				heap.Push(pushMsg.h, pushMsg.x)
			}
		}
	}()
	return quit
}
