package common

const (
	JOB_SAVE_DIR = "/cron/job/"
	JOB_KILL_DIR = "/cron/kill/"
	JOB_LOCK_DIR = "/cron/lock/"
	JOB_WORKER_DIR = "/cron/workers/"

	JOB_EVENT_SAVE = 1
	JOB_EVENT_DELETE = 2
	JOB_EVENT_KILL = 3
)