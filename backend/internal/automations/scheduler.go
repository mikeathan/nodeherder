package automations

import (
	"context"
	"errors"
	"fmt"
	"node-herder/utils"
	"regexp"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/atomic"
)

var (
	ErrUnsupportedTimeFormat = errors.New("scheduler: the given time format is not supported")
	ErrIsRunning             = errors.New("scheduler: the scheduler is already running")
	ErrNotRunning            = errors.New("scheduler: the scheduler is not running")
	ErrNoJobsScheduled       = errors.New("scheduler: no jobs scheduled")
	ErrJobIsRunning          = errors.New("scheduler: the job is already running")

	timeWithSeconds      = regexp.MustCompile(`(?m)^\d{1,2}:\d\d:\d\d$`)
	timeWithMilliseconds = regexp.MustCompile(`^\d{2}:\d{2}:\d{2}\.\d{3}$`)
	timeWithoutSeconds   = regexp.MustCompile(`(?m)^\d{1,2}:\d\d$`)
)

type job struct {
	Name            string
	Id              uuid.UUID
	Action          func() error
	StartAtDuration time.Duration
	startAtTime     string
	RepeatEvery     time.Duration
	timer           utils.Timer
	Error           error
	lock            *sync.RWMutex
	isRunning       *atomic.Bool
	clock           utils.Clock
}

func newJob(clock utils.Clock) *job {

	id := uuid.New()
	return &job{
		Name: id.String(),
		Id:   id,
		Action: func() error {
			return nil
		},
		StartAtDuration: 0,
		RepeatEvery:     0,
		timer:           nil,
		Error:           nil,
		lock:            &sync.RWMutex{},
		isRunning:       atomic.NewBool(false),
		clock:           clock,
	}
}

func (j *job) IsRunning() bool {
	return j.isRunning.Load()
}

func (j *job) Stop() {
	defer j.lock.Unlock()

	j.lock.Lock()
	if j.timer != nil {
		j.timer.Stop()
	}

	j.isRunning.Store(false)
}

func (j *job) Start() error {

	if j.Error != nil {
		utils.LogErrorf("Job %s failed: %s", j.Name, j.Error.Error())
		return j.Error
	}

	// j.timer = time.AfterFunc(j.StartAtDuration, func() {

	// 	utils.LogInfo("Executing job ", j.Name)

	// 	err := j.Action()
	// 	if err != nil {
	// 		utils.LogErrorf("Job %s failed: %s", j.Name, err.Error())
	// 		j.Error = errors.Join(j.Error, err)
	// 	}

	// 	if j.RepeatEvery > 0 {
	// 		nexStartAtDuration, err := j.getStartAtDuration()
	// 		if err != nil {
	// 			j.Error = errors.Join(j.Error, err)
	// 			return
	// 		}

	// 		j.timer.Reset(nexStartAtDuration)
	// 	}
	// })

	j.timer = j.clock.AfterFunc(j.StartAtDuration, func() {
		utils.LogInfo("Executing job ", j.Name)

		err := j.Action()
		if err != nil {
			utils.LogErrorf("Job %s failed: %s", j.Name, err.Error())
			j.Error = errors.Join(j.Error, err)
		}

		if j.RepeatEvery > 0 {
			nextStart, err := j.getStartAtDuration()
			if err != nil {
				j.Error = errors.Join(j.Error, err)
				return
			}
			j.timer.Reset(nextStart)
		}
	})

	j.isRunning.Store(true)
	return nil
}

func (j *job) getStartAtDuration() (time.Duration, error) {

	if j.startAtTime == "" {
		return 0, nil
	}

	startAtTime, err := ConvertStringToTime(j.startAtTime)
	if err != nil {
		utils.LogError("Error parsing start time:", err)
		return 0, err
	}

	now := j.clock.Now()
	startTime := time.Date(now.Year(), now.Month(), now.Day(), startAtTime.Hour(), startAtTime.Minute(), startAtTime.Second(), 0, time.UTC)

	if startTime.Before(now) {
		nextTime := now.Truncate(time.Second).Add(j.RepeatEvery)
		nextDuration := nextTime.Sub(now)

		return nextDuration, nil
	}

	return startTime.Sub(now), nil
}

type Scheduler struct {
	ctx             context.Context
	jobs            map[uuid.UUID]*job
	inScheduleChain *uuid.UUID
	isRunning       *atomic.Bool
	jobsLock        *sync.RWMutex
	clock           utils.Clock
}

func NewScheduler(clock utils.Clock, ctx context.Context) *Scheduler {

	s := &Scheduler{
		ctx:             ctx,
		jobs:            map[uuid.UUID]*job{},
		inScheduleChain: nil,
		isRunning:       atomic.NewBool(false),
		jobsLock:        &sync.RWMutex{},
		clock:           clock,
	}

	go func() {
		select {
		case <-ctx.Done():
			utils.LogError("Scheduler context cancel requested")
			err := s.Stop()
			if err != nil {
				utils.LogError("Error stopping scheduler: ", err)
			}
			return
		}
	}()

	return s
}

func (s *Scheduler) FindJobByStartTime(startTime string) *job {
	for _, job := range s.jobs {
		if job.startAtTime == startTime {
			return job
		}
	}
	return nil
}

func (s *Scheduler) JobByStartAt(startAt string) *job {
	for _, job := range s.jobs {
		if job.startAtTime == startAt {
			return job
		}
	}
	return nil
}

func (s *Scheduler) getCurrentJob() *job {
	if s.inScheduleChain != nil {
		return s.jobs[*s.inScheduleChain]
	}

	j := newJob(s.clock)
	s.jobs[j.Id] = j
	s.inScheduleChain = &j.Id

	return j
}

func (s *Scheduler) Name(name string) *Scheduler {
	job := s.getCurrentJob()

	job.Name = name
	return s
}

func (s *Scheduler) Every(duration time.Duration) *Scheduler {
	job := s.getCurrentJob()

	job.RepeatEvery = duration
	return s
}

func (s *Scheduler) At(timeString string) *Scheduler {
	job := s.getCurrentJob()

	job.startAtTime = timeString
	startAtTime, err := job.getStartAtDuration()
	if err != nil {
		utils.LogError("Error parsing start time:", err)
		job.Error = err
		return s
	}

	job.StartAtDuration = startAtTime
	return s
}

func (s *Scheduler) Do(action func() error) error {
	job := s.getCurrentJob()

	job.Action = action

	s.inScheduleChain = nil

	if job.Error != nil {
		return fmt.Errorf("job %s failed: %s", job.Name, job.Error.Error())
	}
	return nil
}

func ConvertStringToTime(timeString string) (time.Time, error) {

	var layout string
	if timeWithMilliseconds.MatchString(timeString) {
		layout = "15:04:05.000"
	} else if timeWithSeconds.MatchString(timeString) {
		layout = "15:04:05"
	} else if timeWithoutSeconds.MatchString(timeString) {
		layout = "15:04"
	} else {
		return time.Time{}, ErrUnsupportedTimeFormat
	}

	t, err := time.Parse(layout, timeString)
	if err != nil {
		return time.Time{}, ErrUnsupportedTimeFormat
	}

	return t, nil
}

func (s *Scheduler) IsRunning() bool {
	return s.isRunning.Load()
}

func (s *Scheduler) setRunning(vaue bool) {
	s.isRunning.Store(vaue)
}

func (s *Scheduler) Start() error {

	if s.IsRunning() {
		utils.LogInfo("Scheduler already running")
		return ErrIsRunning
	}

	if len(s.jobs) == 0 {
		utils.LogInfo("No jobs scheduled")
		return ErrNoJobsScheduled
	}

	s.jobsLock.Lock()

	defer s.jobsLock.Unlock()

	for _, job := range s.jobs {
		err := job.Start()
		if err != nil {
			return err
		}
	}

	s.setRunning(true)

	utils.LogInfo("Scheduler started")
	return nil
}

func (s *Scheduler) stopJobs() {
	s.jobsLock.RLock()
	defer s.jobsLock.RUnlock()
	for _, job := range s.jobs {
		utils.LogDebug("Scheduler stopping job: ", job.Name)
		job.Stop()
	}
}
func (s *Scheduler) Stop() error {

	if !s.IsRunning() {
		return fmt.Errorf("Scheduler is not running")
	}

	if len(s.jobs) == 0 {
		return fmt.Errorf("no jobs scheduled")
	}

	s.stopJobs()

	// check if all jobs are stopped
	for _, job := range s.jobs {
		if job.IsRunning() {
			return ErrJobIsRunning
		}
	}

	s.setRunning(false)
	utils.LogInfo("Scheduler stopped")

	return nil
}
