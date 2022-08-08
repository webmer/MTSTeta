package models

type KafkaAnalyticMessage struct {
	UUID      string `json:"uuid"`
	Timestamp int64  `json:"timestamp"`
	Type      string `json:"type"`
	Value     string `json:"value"`
}
