package cron

import (
	"sync"
	"time"
)

type Cron struct {
	mu   sync.Mutex
	jobs map[string]job

	stop chan struct{}
}

func NewCron() *Cron {
	return &Cron{
		jobs: make(map[string]job),
		stop: make(chan struct{}),
	}
}

// Register a function to execute when current time matches the pattern specified by ts
func (c *Cron) Register(id string, ts string, fn func()) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	job, err := newJob(ts, fn)
	if err != nil {
		return err
	}
	c.jobs[id] = job
	return nil
}

// Delete a regsitered function from cron.
func (c *Cron) Delete(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.jobs, id)
}

// Stop the cron. Currently running functions will not be stopped.
func (c *Cron) Stop() { close(c.stop) }

// start the cron
//
// Every minute loop through the registered funcions and check its execution time pattern.
// If a time pattern matches the current time then the corresponding function will be called.
//
// Function runs in its own goroutine.
func (c *Cron) Start() {
	ticker := time.Tick(1 * time.Minute)
	for {
		select {
		case <-c.stop:
			return
		case now := <-ticker:
			c.run(now)
		}
	}
}

func (c *Cron) run(now time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, job := range c.jobs {
		if job.match(now) {
			go job.run()
		}
	}
}

type job struct {
	ts   string
	expr CronExpr
	fn   func()
}

func newJob(ts string, fn func()) (j job, err error) {
	j.ts = ts
	j.fn = fn
	j.expr, err = NewCronExpr(ts)
	return
}

func (j *job) run()                   { j.fn() }
func (j *job) match(t time.Time) bool { return j.expr.Match(t) }
