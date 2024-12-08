package pubsub

type SimpleQueueType int

const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTranscient
)

type AckType int

const (
	Ack AckType = iota
	NackDiscard
	NackRequeue
)
