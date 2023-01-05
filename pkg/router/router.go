package router

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/lazyboson/beta-router/pkg/httpclient"
	"github.com/lazyboson/beta-router/pkg/models"
	pb "github.com/lazyboson/beta-router/pkg/pb/apipb"
)

type Router struct {
	conf *Config
}

type Config struct {
	QueueBaseUrl string
	CtiBaseUrl   string
}

func NewRouter(conf *Config) *Router {
	return &Router{
		conf: conf,
	}
}

func (r *Router) ListenEvents(req *pb.TaskCreationEventRequest) *pb.TaskEventResponse {
	res := &pb.TaskEventResponse{Message: "Code 0, Execution Successful"}
	accountId := req.GetAccountId()
	if accountId == "" {
		log.Printf("empty account id received")
		return res
	}

	go r.handleTask(accountId)

	return res
}

func (r *Router) handleTask(accountId string) {
	time.Sleep(5 * time.Second)
	taskViewUri := r.conf.QueueBaseUrl + taskViewPath
	fmt.Printf("TaskViewUrl: %+v \n", taskViewUri)

	taskViewPayload := &models.TaskViewPayload{
		AccountId: accountId,
		TaskCount: taskCount,
	}

	res, err := httpclient.Post(taskViewPayload, taskViewUri, map[string]string{ContentType: ContentTypeJSON})
	if err != nil {
		fmt.Errorf("failed to fetch task view builder res: %+v \n", err)
		return
	}

	topKTasks := &models.TopKTasks{}
	err = json.Unmarshal(res, topKTasks)
	if err != nil {
		fmt.Errorf("failed to unmarshall k task resp: %+v \n", err)
		return
	}
	// pull context and transferring call to active Agent
	for i := 0; i < len(topKTasks.Tasks); i++ {
		task := topKTasks.Tasks[i]
		contextPullUrl := r.conf.QueueBaseUrl + queuePath + task + contextPull
		fmt.Printf("context Pull Url :%s \n", contextPullUrl)
		res, err = httpclient.Post("", contextPullUrl, map[string]string{ContentType: ContentTypeJSON})
		if err != nil {
			fmt.Errorf("failed to pull context for the call: %+v \n", err)
			continue
		}
		taskContext := &models.TaskContext{}
		err = json.Unmarshal(res, taskContext)
		if err != nil {
			fmt.Errorf("failed to unmarshall call context: %+v \n", err)
			return
		}

		fmt.Printf("Context Pull: data %+v \n", taskContext)
		agentConvParams := prepareAgentServiceConversationParams(taskContext.CallContext)
		err = r.previewCallToAgent(agentConvParams)
		if err != nil {
			fmt.Errorf("failed to prepare task: %+v \n", err)
			return
		}

		hangupPayload := &models.HangupEvent{
			CallStatus: "completed",
			CallUuid:   agentConvParams.Conversation.Interaction.CallId,
		}

		hangupUrl := r.conf.QueueBaseUrl + hangupPath
		fmt.Printf("hangup Path: %v\n", hangupUrl)
		_, err = httpclient.Post(hangupPayload, hangupUrl, map[string]string{ContentType: ContentTypeJSON})
		if err != nil {
			fmt.Errorf("failed to send hangup event: %+v \n", err)
			return
		}

	}
}

func prepareAgentServiceConversationParams(context map[string]interface{}) *models.AgentServiceConvParam {
	agentConvParams := &models.AgentServiceConvParam{}
	var callUUID, ccid, chanType, convFlowId, contextCTx, ctxConv interface{}
	var ok bool

	if callUUID, ok = context["call_uuid"]; ok {
		agentConvParams.Conversation.Interaction.CallId = callUUID.(string)
	}

	if ccid, ok = context["account_id"]; ok {
		agentConvParams.Ccid = ccid.(string)
	}

	if chanType, ok = context["channel_type"]; ok {
		agentConvParams.ChannelId = chanType.(string)
	}

	if contextCTx, ok = context["context"]; !ok {
		return agentConvParams
	}
	mpCtx := contextCTx.(map[string]interface{})
	if convFlowId, ok = mpCtx["conversationFlowId"]; ok {
		agentConvParams.ConversationFlowId = convFlowId.(string)
	}

	if ctxConv, ok = mpCtx["conversation"]; !ok {
		return agentConvParams
	}
	mpConv := ctxConv.(map[string]interface{})
	if convId, ok := mpConv["id"]; ok {
		agentConvParams.Conversation.Id = convId.(string)
	}
	if ctx, ok := mpConv["context"]; ok {
		agentConvParams.Conversation.Context = ctx
	}

	if _, ok := mpConv["interaction"]; !ok {
		return agentConvParams
	}

	mpInteraction := mpConv["interface"].(map[string]interface{})
	if interID, ok := mpInteraction["id"]; ok {
		agentConvParams.Conversation.Interaction.CallId = interID.(string)
	}
	if interDir, ok := mpInteraction["direction"]; ok {
		agentConvParams.Conversation.Interaction.Direction = interDir.(string)
	}
	if interCtx, ok := mpInteraction["context"]; ok {
		agentConvParams.Conversation.Interaction.Context = interCtx
	}
	if interDisName, ok := mpInteraction["displayName"]; ok {
		agentConvParams.Conversation.Interaction.DisplayName = interDisName.(string)
	}
	if interRmtSrtAt, ok := mpInteraction["remoteStartAt"]; ok {
		y := interRmtSrtAt.(float64)
		agentConvParams.Conversation.Interaction.RemoteStartAt = y
	}
	if interInvAs, ok := mpInteraction["invitedAs"]; ok {
		agentConvParams.Conversation.Interaction.InvitedAs = interInvAs.(string)
	}
	if interTo, ok := mpInteraction["toAddress"]; ok {
		agentConvParams.Conversation.Interaction.ToAddress = interTo.(string)
	}
	if interFrom, ok := mpInteraction["fromAddress"]; ok {
		agentConvParams.Conversation.Interaction.FromAddress = interFrom.(string)
	}

	return agentConvParams
}

func (r *Router) previewCallToAgent(body interface{}) error {
	_, err := httpclient.Post(body, r.conf.CtiBaseUrl, map[string]string{ContentType: ContentTypeJSON})
	if err != nil {
		return err
	}

	return err
}
