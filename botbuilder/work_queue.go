package botbuilder

import (
	"sync"
)

type workQueue struct {
	priorityQueue priorityQueue
	lock          sync.RWMutex
}

func newWorkQueue() *workQueue {
	return &workQueue{
		priorityQueue: newPriorityQueue(),
		lock:          sync.RWMutex{},
	}
}

func (queue *workQueue) Push(task Task) {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	queue.priorityQueue.Push(task)
}

func (queue *workQueue) Pop() Task {
	queue.lock.Lock()
	defer queue.lock.Unlock()

	return queue.priorityQueue.Pop().(Task)
}

func (queue *workQueue) Len() int {
	queue.lock.RLock()
	defer queue.lock.RUnlock()

	return queue.priorityQueue.Len()
}

func (queue *workQueue) IsEmpty() bool {
	queue.lock.RLock()
	defer queue.lock.RUnlock()

	return queue.priorityQueue.Len() == 0
}
