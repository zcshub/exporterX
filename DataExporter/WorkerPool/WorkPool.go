package workpool

type Task struct {
	Id   string
	Info string
	Err  error
	F    func(int) (string, error)
}

func (task *Task) Do(n int) (string, error) {
	return task.F(n)
}

type WorkPool struct {
	PoolSize    int
	tasksSize   int
	tasksChan   chan Task
	resultsChan chan Task
	Results     func() map[string]string
}

func NewWorkPool(tasks []Task, poolsize int) *WorkPool {
	tasksChan := make(chan Task, len(tasks))
	resultsChan := make(chan Task, len(tasks))
	for _, task := range tasks {
		tasksChan <- task
	}
	close(tasksChan)
	pool := &WorkPool{
		PoolSize:    poolsize,
		tasksSize:   len(tasks),
		tasksChan:   tasksChan,
		resultsChan: resultsChan,
	}
	pool.Results = pool.results
	return pool
}

func (p *WorkPool) Start() {
	for i := 0; i < p.PoolSize; i++ {
		go p.work(i)
	}
}

func (p *WorkPool) work(n int) {
	// defer func() {
	// 	if e := recover(); e != nil {
	// 		os.Exit(1)
	// 	}
	// }()
	for task := range p.tasksChan {
		task.Info, task.Err = task.Do(n)
		p.resultsChan <- task
	}
}

func (p *WorkPool) results() map[string]string {
	result := make(map[string]string)
	for i := 0; i < p.tasksSize; i++ {
		task := <-p.resultsChan
		if _, ok := result[task.Id]; ok {
			panic("Duplicate " + task.Id)
		}
		result[task.Id] = task.Info
	}
	return result
}
