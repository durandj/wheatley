package botbuilder

type queueJob struct {
	value    interface{}
	priority int
	index    int
}

type priorityQueue []*queueJob

func newPriorityQueue() *priorityQueue {
	queue := make(priorityQueue, 0)

	return &queue
}

func (queue priorityQueue) Len() int {
	return len(queue)
}

func (queue priorityQueue) Less(i int, j int) bool {
	// TODO(jdurand): it might be helpful to include more complex rules such as time in queue or time submitted
	return queue[i].priority > queue[j].priority
}

func (queue priorityQueue) Swap(i int, j int) {
	queue[i], queue[j] = queue[j], queue[i]

	queue[i].index = j
	queue[j].index = i
}

func (queue *priorityQueue) Push(task interface{}) {
	index := len(*queue)
	job := &queueJob{value: task, index: index}
	*queue = append(*queue, job)
}

func (queue *priorityQueue) Pop() interface{} {
	oldQueue := *queue
	length := len(oldQueue)
	job := oldQueue[length-1]
	oldQueue[length-1] = nil
	job.index = -1
	*queue = oldQueue[0 : length-1]
	return job.value
}
