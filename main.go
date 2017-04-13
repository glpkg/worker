package worker

import (
	"fmt"
	"sync"
	"time"
)

type Tasker interface {
	Do(c chan bool)
}

type Work struct {
	debug     bool
	sleepTime time.Duration
	lock      sync.Mutex
	limit     int
	taskList  []Tasker
}

// 设置没有任务的时候的睡眠时间
func (w *Work) SetSleepTime(t time.Duration) {
	w.sleepTime = t
}

// 设置是否为debug
func (w *Work) Debug(debug bool) *Work {
	w.debug = debug
	return w
}

// 设置goroutine速度
func (w *Work) SetLimit(limit int) {
	w.limit = limit
}

// 增加一个任务到列表里面
func (w *Work) AddTask(t Tasker) {
	w.taskList = append(w.taskList, t)
	if w.debug {
		fmt.Println("增加了一个任务，现在任务总个数为：", len(w.taskList))
	}
}

// 获取一个任务，以及是否获取成功
func (w *Work) getTask() (Tasker, bool) {
	w.lock.Lock()
	defer w.lock.Unlock()
	if len(w.taskList) == 0 {
		return nil, false
	}
	t := w.taskList[0]
	w.taskList = w.taskList[1:]
	return t, true
}

// 任务开始
func (w *Work) Start() {
	c := make(chan bool, w.limit)

	for i := 0; i < w.limit; i++ {
		for {
			t, ok := w.getTask()
			if ok {
				go t.Do(c)
				break
			}
			time.Sleep(w.sleepTime)
		}
	}
	for {
		<-c
		t, ok := w.getTask()
		if ok {
			go t.Do(c)
		}
		time.Sleep(w.sleepTime)
	}

}

// 返回任务列表现有的个数
func (w *Work) Len() int {
	return len(w.taskList)
}

// 创建一个Work
func NewWork() *Work {
	w := new(Work)
	w.sleepTime = time.Second
	w.limit = 1
	return w
}
