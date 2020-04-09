package configs

import (
	"encoding/json"
	"time"
)

type Config struct {
	PreviewURL string `json:"preview_url"`
	Listen     string `json:"listen"`

	MongoURI          string `json:"mongo_uri"`
	MongoDB           string `json:"mongo_db"`
	APIToken          string `json:"api_token"`
	UseLocalDashboard bool   `json:"use_local_dashboard"`

	AggregationPeriod     time.Duration `json:"aggregation_period"`
	ActionTriggerPeriod   time.Duration `json:"action_trigger_period"`
	QueueJobMaxRetryTimes int           `json:"queue_job_max_retry_times"`
	QueueWorkerNum        int           `json:"queue_worker_num"`

	Migrate bool `json:"migrate"`
}

func (conf *Config) Serialize() string {
	rs, _ := json.Marshal(conf)
	return string(rs)
}
