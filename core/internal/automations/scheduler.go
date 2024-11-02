package automations

import (
	"context"
	"errors"
	"fmt"
	"node-herder/utils"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.uber.org/atomic"
)

var (
	ErrUnsupportedTimeFormat = errors.New("scheduler: the given time format is not supported")
)

type TimeSchedule struct {
	Enabled bool   `json:"enabled"`
	Start   string `json:"start"`
	End     string `json:"end"`
}

func NewTimeSchedule() *TimeSchedule {
	return &TimeSchedule{
		Enabled: false,
		Start:   "",
		End:     "",
	}
}

type job struct {
	Name            string
	Id              uuid.UUID
	Action          func() error
	StartAtDuration time.Duration
	RepeatEvery     time.Duration
	timer           *time.Timer
	Error           error
	ctx             context.Context
	cancel          context.CancelFunc
	lock            *sync.RWMutex
	isRunning       *atomic.Bool
}

func newJob() *job {
	ctx, cancel := context.WithCancel(context.Background())

	id := uuid.New()
	return &job{
		Name: id.String(),
		Id:   id,
		Action: func() error {
			return nil
		},
		StartAtDuration: 0,
		RepeatEvery:     0,
		timer:           &time.Timer{},
		Error:           nil,
		ctx:             ctx,
		cancel:          cancel,
		lock:            &sync.RWMutex{},
		isRunning:       atomic.NewBool(false),
	}
}

func (j *job) Stop() {
	defer j.lock.Unlock()

	j.lock.Lock()
	if j.timer != nil {
		j.timer.Stop()
	}

	if j.cancel != nil {
		j.cancel()
		j.ctx, j.cancel = context.WithCancel(context.Background())
	}

	j.isRunning.Store(true)
}

func (j *job) IsRunning() bool {
	return j.isRunning.Load()
}

func (j *job) Start() error {

	if j.Error != nil {
		utils.LogErrorf("Job %s failed: %s", j.Name, j.Error.Error())
		return j.Error
	}

	j.timer = time.AfterFunc(j.StartAtDuration, func() {

		select {

		case <-j.ctx.Done():

			utils.LogInfo("Scheduler context cancel requested")
			j.Stop()

			return

		default:
			utils.LogInfo("Executing job ", j.Name)

			err := j.Action()
			if err != nil {
				utils.LogErrorf("Job %s failed: %s", j.Name, err.Error())
				j.Error = errors.Join(j.Error, err)
			}

			if j.RepeatEvery > 0 {
				j.timer.Reset(j.RepeatEvery)
			}
		}
	})

	return nil
}

type Scheduler struct {
	ctx             context.Context
	jobs            map[uuid.UUID]*job
	inScheduleChain *uuid.UUID
	isRunning       *atomic.Bool
	jobsLock        *sync.RWMutex
}

func NewScheduler(ctx context.Context) *Scheduler {
	return &Scheduler{
		ctx:             ctx,
		jobs:            map[uuid.UUID]*job{},
		inScheduleChain: nil,
		isRunning:       atomic.NewBool(false),
		jobsLock:        &sync.RWMutex{},
	}
}

func (s *Scheduler) getCurrentJob() *job {
	if s.inScheduleChain != nil {
		return s.jobs[*s.inScheduleChain]
	}

	j := newJob()
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

	timestamp, err := parserTime(timeString)
	if err != nil {
		utils.LogError("Error parsing start time:", err)
		job.Error = err
		return s
	}

	now := time.Now().UTC()
	startAtTime := time.Date(now.Year(), now.Month(), now.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0, time.UTC)

	job.StartAtDuration = startAtTime.Sub(now)

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

// // time , repeat, func, name
// func (s *Scheduler) AddJob(timeString string, repeat time.Duration, action func() error) error {

// 	now := time.Now().UTC()
// 	timestamp, err := parserTime(timeString)

// 	if err != nil {
// 		utils.LogError("Error parsing start time:", err)
// 		return err
// 	}

// 	stm := time.Date(now.Year(), now.Month(), now.Day(), timestamp.Hour(), timestamp.Minute(), timestamp.Second(), 0, time.UTC)
// 	duration := stm.Sub(now)
// 	var job *time.Timer

// 	fmt.Println(now, " trigger time: ", timestamp, " - ", stm, " duration: ", duration)
// 	job = time.AfterFunc(duration, func() {

// 		select {

// 		case <-s.ctx.Done():

// 			utils.LogInfo("Scheduler context cancel requested")
// 			s.Stop()

// 			return

// 		default:
// 			utils.LogInfo("Executing job ", stm, " - ", time.Now().UTC())

// 			err = action()
// 			if err != nil {
// 				utils.LogErrorf("Job %s failed: %s", timeString, err.Error())
// 			}

// 			job.Reset(repeat)
// 		}
// 	})

// 	//name := fmt.Sprintf("Job %s", timeString)

// 	job, err := s.cronScheduler.Name(name).Every(1).Day().At(time).Do(func() {
// 	// 	select {

// 	// 	case <-s.ctx.Done():

// 	// 		utils.LogInfo("Scheduler context cancel requested")
// 	// 		s.Stop()

// 	// 		return

// 	// 	default:
// 	// 		utils.LogInfo("Executing job ", time)

// 	// 		err = action()
// 	// 		if err != nil {
// 	// 			utils.LogErrorf("Job %s failed: %s", timeString, err.Error())
// 	// 		}
// 	// 	}

// 	// })

// 	if err != nil {
// 		utils.LogErrorf("Job %s failed to schedule: %s", timeString, err.Error())
// 		return err
// 	}
// 	s.jobs = append(s.jobs, job)

// 	return nil
// }

func parserTime(timeString string) (time.Time, error) {
	layout := "15:04:05.000"
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

func (s *Scheduler) Start() {

	// TODO: use enabled/disabled logic ???

	if s.IsRunning() {
		utils.LogInfo("Scheduler already running")
		return
	}
	if len(s.jobs) == 0 {
		utils.LogInfo("No jobs scheduled")
		return
	}

	s.jobsLock.Lock()

	defer s.jobsLock.Unlock()

	for _, job := range s.jobs {
		job.Start()
	}

	s.setRunning(true)

	utils.LogInfo("Scheduler started")
}

func (s *Scheduler) stopJobs() {
	s.jobsLock.RLock()
	defer s.jobsLock.RUnlock()
	for _, job := range s.jobs {
		job.Stop()
	}
}
func (s *Scheduler) Stop() error {

	if !s.IsRunning() {
		return fmt.Errorf("scheduler is not running")
	}

	if len(s.jobs) == 0 {
		return fmt.Errorf("no jobs scheduled")
	}

	s.stopJobs()

	// check if all jobs are stopped
	for _, job := range s.jobs {
		if job.IsRunning() {
			return fmt.Errorf("job %s is still running", job.Error)
		}
	}

	s.setRunning(false)
	utils.LogInfo("Scheduler stopped")

	return nil
}

// type DaySchedule struct {
// 	Start int `json:"start"`
// 	End   int `json:"end"`
// }
// day schedule
// start day
// end day

// time schedule
// start time
// end time
