package cron

import (
	"fmt"
	taskqueue "scrapeit/internal/task-queue"
	"sync"
	"time"

	"github.com/robfig/cron/v3"
)

// Logger is an interface for logging operations
type Logger interface {
	Info(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

type CronJobStatus string

const (
	CronJobStatusRunning CronJobStatus = "running"
	CronJobStatusIdle    CronJobStatus = "idle"
)

type CronManagerJob struct {
	ID         cron.EntryID
	GroupID    string
	EndpointID string
	Interval   string
	LastRun    string
	Active     bool
	Status     CronJobStatus
	Job        func() error
}

type CronManager struct {
	Jobs      []*CronManagerJob
	TaskQueue *taskqueue.TaskQueue
	cron      *cron.Cron
	mu        sync.RWMutex
	logger    Logger
}

func NewCronManager(logger Logger, maxConcurrentTasks, numWorkers int) *CronManager {
	cm := &CronManager{
		Jobs:      []*CronManagerJob{},
		TaskQueue: taskqueue.NewTaskQueue(maxConcurrentTasks, numWorkers),
		cron:      cron.New(),
		logger:    logger,
	}
	cm.cron.Start() // Start the cron scheduler
	return cm
}

func (cm *CronManager) AddJob(job CronManagerJob) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.Jobs = append(cm.Jobs, &job)
	addedJob := cm.getJob(job.GroupID, job.EndpointID)
	cm.logger.Info("Adding new job", "groupId", job.GroupID, "endpointId", job.EndpointID, "interval", job.Interval)

	entryId, err := cm.cron.AddFunc(addedJob.Interval, func() {
		if !addedJob.Active {
			cm.logger.Info("Job inactive", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID)
			return
		}
		if addedJob.Status == CronJobStatusRunning {
			cm.logger.Info("Job already running", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID)
			return
		}
		addedJob.Status = CronJobStatusRunning
		cm.logger.Info("Queueing job execution", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID)

		cm.TaskQueue.AddTask(func() error {
			defer func() {
				cm.mu.Lock()
				addedJob.Status = CronJobStatusIdle
				addedJob.LastRun = time.Now().Format(time.RFC3339)
				cm.mu.Unlock()
			}()

			err := addedJob.Job()
			if err != nil {
				cm.logger.Error("Error running job", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID, "error", err)
			} else {
				cm.logger.Info("Job ran successfully", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID)
			}
			return err
		})
	})

	if err != nil {
		cm.logger.Error("Error adding job", "groupId", job.GroupID, "endpointId", job.EndpointID, "error", err)
	} else {
		addedJob.ID = entryId
		cm.logger.Info("Job added successfully", "groupId", job.GroupID, "endpointId", job.EndpointID, "entryId", entryId)
	}

	cm.formatAllJobs()
}

func (cm *CronManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, job := range cm.Jobs {
		cm.cron.Remove(job.ID)
	}
	cm.cron.Stop()
	cm.TaskQueue.Stop()
	cm.logger.Info("Cron manager and task queue stopped")
}

func (cm *CronManager) DestroyJob(groupId string, endpointId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	job := cm.getJob(groupId, endpointId)
	if job != nil {
		cm.cron.Remove(job.ID)

		// Remove the job from the Jobs slice
		for i, j := range cm.Jobs {
			if j.GroupID == groupId && j.EndpointID == endpointId {
				cm.Jobs = append(cm.Jobs[:i], cm.Jobs[i+1:]...)
				break
			}
		}

		cm.logger.Info("Job destroyed", "groupId", groupId, "endpointId", endpointId)
	} else {
		cm.logger.Info("Job not found for destruction", "groupId", groupId, "endpointId", endpointId)
	}
	cm.formatAllJobs()
}

func (cm *CronManager) UpdateJob(job CronManagerJob) {
	cm.mu.Lock()

	cm.destroyJob(job.GroupID, job.EndpointID)
	cm.AddJob(job)
	cm.logger.Info("Job updated", "groupId", job.GroupID, "endpointId", job.EndpointID)
	cm.mu.Unlock()
	cm.formatAllJobs()
}

func (cm *CronManager) UpdateJobInterval(groupId, endpointId, interval string) {
	cm.mu.Lock()

	job := *cm.getJob(groupId, endpointId)

	job.Interval = interval
	cm.destroyJob(groupId, endpointId)

	cm.mu.Unlock()
	cm.AddJob(job)

	cm.logger.Info("Job interval updated", "groupId", groupId, "endpointId", endpointId, "newInterval", interval)
	cm.formatAllJobs()
}

func (cm *CronManager) StopJob(groupId string, endpointId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	job := cm.getJob(groupId, endpointId)
	if job != nil {
		job.Active = false
		cm.logger.Info("Job stopped", "groupId", groupId, "endpointId", endpointId)
	}

}

func (cm *CronManager) StartJob(groupId string, endpointId string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	job := cm.getJob(groupId, endpointId)
	if job != nil {
		job.Active = true
		cm.logger.Info("Job started", "groupId", groupId, "endpointId", endpointId)
	}

}

func (cm *CronManager) GetJob(groupId string, endpointId string) *CronManagerJob {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	return cm.getJob(groupId, endpointId)
}

// Private methods

func (cm *CronManager) getJob(groupId string, endpointId string) *CronManagerJob {
	for _, job := range cm.Jobs {
		if job.GroupID == groupId && job.EndpointID == endpointId {
			return job
		}
	}
	return nil
}

func (cm *CronManager) destroyJob(groupId string, endpointId string) {
	job := cm.getJob(groupId, endpointId)
	if job != nil {
		// Stop the cron job
		cm.cron.Remove(job.ID)

		// If the job is currently running, wait for it to finish
		if job.Status == CronJobStatusRunning {
			cm.logger.Info("Waiting for job to finish before destroying", "groupId", groupId, "endpointId", endpointId)
			for job.Status == CronJobStatusRunning {
				cm.mu.Unlock()
				time.Sleep(100 * time.Millisecond)
				cm.mu.Lock()
			}
		}

		// Remove the job from the Jobs slice
		for i, j := range cm.Jobs {
			if j.GroupID == groupId && j.EndpointID == endpointId {
				cm.Jobs = append(cm.Jobs[:i], cm.Jobs[i+1:]...)
				break
			}
		}

		cm.logger.Info("Job destroyed", "groupId", groupId, "endpointId", endpointId)
	} else {
		cm.logger.Info("Job not found for destruction", "groupId", groupId, "endpointId", endpointId)
	}
}

func (cm *CronManager) formatAllJobs() {

	type formattedJob struct {
		ID         cron.EntryID
		GroupID    string
		EndpointID string
		Interval   string
		LastRun    string
	}
	jobsToLog := []*formattedJob{}
	for _, job := range cm.Jobs {
		jobsToLog = append(jobsToLog, &formattedJob{
			ID:         job.ID,
			GroupID:    job.GroupID,
			EndpointID: job.EndpointID,
			Interval:   job.Interval,
			LastRun:    job.LastRun,
		})
	}

	fmt.Println("--------------------")
	for _, log := range jobsToLog {
		fmt.Printf("Job: %v\n", log)
	}
	fmt.Println("--------------------")
}
