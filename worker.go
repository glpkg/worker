package worker

import (
	"fmt"
	"sync"
	"time"
)

type Tasker interface {
	Do() bool
}

type Work struct {
	debug         bool //是否开启debu，如果开启debug，则会在对应的地方执行log函数。
	isStop        bool
	retryTimes    uint8         //如果任务失败重试的次数，默认为0。
	c             chan bool     //用于通讯的chan，长度根据limit长度在start的时候进行make
	idleSleepTime time.Duration //没有获取到任务再次获取的间隔时间，默认为1s。
	taskSleepTime time.Duration //每个任务之间间隔的时间，默认为0。
	lock          sync.Mutex    //获取task的时候要进行并发处理，所以先将slice锁起来。
	limit         int           //最多同时有几个任务
	taskList      []Tasker      //任务列表
}

// 设置失败重试的次数
func (w *Work) RetryTimes(t uint8) {
	w.retryTimes = t
}

// 设置没有任务的时候的睡眠时间
func (w *Work) IdleSleepTime(t time.Duration) {
	w.idleSleepTime = t
}

// 设置每个任务之间的间隔时间
func (w *Work) TaskSleepTime(t time.Duration) {
	w.taskSleepTime = t
}

// 设置是否为debug
func (w *Work) Debug(debug bool) *Work {
	w.debug = debug
	return w
}

// 设置goroutine速度
func (w *Work) Limit(limit int) {
	w.limit = limit
}

// 增加一个任务到列表里面
func (w *Work) AddTask(t Tasker) {
	w.taskList = append(w.taskList, t)
	if w.debug {
		w.log()
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
	w.c = make(chan bool, w.limit)

	for i := 0; i < w.limit; i++ {
		go w.Do()
	}

	for {
		<-w.c
		if !w.isStop {
			go w.Do()
		} else {
			break
		}
	}

}

// 获取执行任务进行执行
func (w *Work) Do() {
	for {
		t, ok := w.getTask()
		try := w.retryTimes
		ret := false
		if ok {
			for {
				ret = t.Do()
				if w.taskSleepTime != 0 {
					time.Sleep(w.taskSleepTime)
				}
				if ret {
					break
				} else {
					if try <= 0 {
						break
					} else {
						continue
					}
				}
			}
			w.c <- ret
			break
		}
		time.Sleep(w.idleSleepTime)
	}

}

// 停止所有任务，已经执行的任务会等执行完毕之后才会停止。
func (w *Work) Stop() {
	w.isStop = true
}

// 返回任务列表现有的个数
func (w *Work) Len() int {
	return len(w.taskList)
}

// debug的时候会调用这个函数
func (w *Work) log() {
	fmt.Println("task len:", w.Len(), "limit", w.limit, "idleSleepTime:", w.idleSleepTime, "taskSleepTime:", w.taskSleepTime)
}

// 创建一个Work
func NewWork() *Work {
	w := new(Work)
	w.idleSleepTime = time.Second
	w.taskSleepTime = 0
	w.limit = 1
	w.retryTimes = 0
	w.isStop = false
	return w
}
