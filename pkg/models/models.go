package models

type TaskViewPayload struct {
	AccountId string `json:"account_id"`
	TaskCount uint32 `json:"task_count"`
}

type TopKTasks struct {
	Tasks   []string `json:"tasks"`
	Message string   `json:"message"`
}

type TaskContext struct {
	CallContext map[string]interface{} `json:"call_context"`
	TaskId      string                 `json:"task_id"`
	QueueId     string                 `json:"queue_id"`
}

type Interaction struct {
	CallId        string      `json:"id"` // unique id
	Direction     string      `json:"direction"`
	Context       interface{} `json:"context"`
	DisplayName   string      `json:"displayName"` // name to be displayed on cti
	RemoteStartAt float64     `json:"remoteStartAt"`
	InvitedAs     string      `json:"invitedAs"`
	ToAddress     string      `json:"toAddress"`
	FromAddress   string      `json:"fromAddress"`
}

type Conversation struct {
	Id          string      `json:"id"` // unique uuid for a conversation (Business flow id)
	Context     interface{} `json:"context"`
	Interaction Interaction `json:"interaction"`
}

type Task struct {
	TaskId    string `json:"id"`
	QueueId   string `json:"queueId"`
	QueueName string `json:"queueName"`
	Priority  uint32 `json:"priority"`
}

type AgentServiceConvParam struct {
	Ccid               string       `json:"ccid"`
	ConversationFlowId string       `json:"conversationFlowId"` // id of portal flow/project
	ChannelId          string       `json:"channelId"`
	Task               Task         `json:"task"`
	Conversation       Conversation `json:"conversation"`
}

type HangupEvent struct {
	CallUuid   string `json:"call_uuid"`
	CallStatus string `json:"call_status"`
}
