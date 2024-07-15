package cron

import (
	"fmt"
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
	Jobs   []*CronManagerJob
	cron   *cron.Cron
	mu     sync.RWMutex
	logger Logger
}

func NewCronManager(logger Logger) *CronManager {
	cm := &CronManager{
		Jobs:   []*CronManagerJob{},
		cron:   cron.New(),
		logger: logger,
	}
	cm.cron.Start() // Start the cron scheduler
	return cm
}

func (cm *CronManager) DestroyJob(groupId string, endpointId string) {
	cm.mu.Lock()

	job := cm.getJob(groupId, endpointId)
	if job != nil {
		cm.cron.Remove(job.ID)
		for i, j := range cm.Jobs {
			if j.GroupID == groupId && j.EndpointID == endpointId {
				cm.Jobs = append(cm.Jobs[:i], cm.Jobs[i+1:]...)
				cm.logger.Info("Job destroyed", "groupId", groupId, "endpointId", endpointId)
				break
			}
		}
	}
	cm.mu.Unlock()
	cm.formatAllJobs()
}

func (cm *CronManager) AddJob(job CronManagerJob) {
	cm.mu.Lock()

	cm.Jobs = append(cm.Jobs, &job)
	addedJob := cm.getJob(job.GroupID, job.EndpointID)
	cm.logger.Info("Adding new job", "groupId", job.GroupID, "endpointId", job.EndpointID, "interval", job.Interval)
	fmt.Println("here")

	entryId, err := cm.cron.AddFunc(addedJob.Interval, func() {
		if addedJob.Status == CronJobStatusRunning {
			cm.logger.Info("Job already running", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID)
			return
		}
		addedJob.Status = CronJobStatusRunning
		cm.logger.Info("Starting job execution", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID)
		err := addedJob.Job()
		if err != nil {
			cm.logger.Error("Error running job", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID, "error", err)
		} else {
			cm.logger.Info("Job ran successfully", "groupId", addedJob.GroupID, "endpointId", addedJob.EndpointID)
			cm.mu.Lock()
			job := cm.getJob(addedJob.GroupID, addedJob.EndpointID)
			if job != nil {
				addedJob.LastRun = time.Now().Format(time.RFC3339)
			}
			addedJob.Status = CronJobStatusIdle
			cm.mu.Unlock()
		}
	})

	if err != nil {
		cm.logger.Error("Error adding job", "groupId", job.GroupID, "endpointId", job.EndpointID, "error", err)
	} else {
		newJob := cm.getJob(job.GroupID, job.EndpointID)
		if newJob != nil {
			newJob.ID = entryId
			cm.logger.Info("Job added successfully", "groupId", job.GroupID, "endpointId", job.EndpointID, "entryId", entryId)
		} else {
			cm.logger.Error("Failed to retrieve newly added job", "groupId", job.GroupID, "endpointId", job.EndpointID)
		}
	}
	cm.mu.Unlock()
	cm.formatAllJobs()
}

func (cm *CronManager) Stop() {
	for _, job := range cm.Jobs {
		cm.DestroyJob(job.GroupID, job.EndpointID)
	}
	cm.cron.Stop()
}

func (cm *CronManager) UpdateJob(job CronManagerJob) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	cm.destroyJob(job.GroupID, job.EndpointID)
	cm.AddJob(job)
	cm.logger.Info("Job updated", "groupId", job.GroupID, "endpointId", job.EndpointID)
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

		cm.cron.Remove(job.ID)
		cm.logger.Info("Job destroyed", "groupId", groupId, "endpointId", endpointId)

		for i, j := range cm.Jobs {
			if j.GroupID == groupId && j.EndpointID == endpointId {

				cm.Jobs = append(cm.Jobs[:i], cm.Jobs[i+1:]...)
				break
			}
		}
	}
	cm.formatAllJobs()
}

func (cm *CronManager) formatAllJobs() {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

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

	for _, log := range jobsToLog {
		fmt.Printf("Job: %v\n", log)
	}
}
