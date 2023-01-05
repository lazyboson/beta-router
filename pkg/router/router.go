package router

import (
	"encoding/json"
	"fmt"
	"log"

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
	taskViewUri := r.conf.QueueBaseUrl + taskViewPath
	taskViewPayload := &models.TaskViewPayload{
		AccountId: accountId,
		TaskCount: taskCount,
	}

	res, err := httpclient.Post(taskViewPayload, taskViewUri, map[string]string{ContentType: ContentTypeJSON})
	if err != nil {
		log.Fatal(err)
	}

	topKTasks := &models.TopKTasks{}
	err = json.Unmarshal(res, topKTasks)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Top K Tasks: %+v", topKTasks)
	// pull context and transferring call to active Agent
	for i := 0; i < len(topKTasks.Tasks); i++ {
		task := topKTasks.Tasks[i]
		contextPullUrl := r.conf.QueueBaseUrl + queuePath + task + contextPull
		res, err = httpclient.Post("", contextPullUrl, map[string]string{ContentType: ContentTypeJSON})
		if err != nil {
			log.Fatal(err)
		}
		taskContext := &models.TaskContext{}
		err = json.Unmarshal(res, taskContext)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("Context Pull: data %+v", taskContext)
		agentConvParams := prepareAgentServiceConversationParams(taskContext.CallContext)
		err := r.previewCallToAgent(agentConvParams)
		if err != nil {
			return
		}

		hangupPayload := &models.HangupEvent{
			CallStatus: "completed",
			CallUuid:   agentConvParams.Conversation.Interaction.CallId,
		}

		hangupUrl := r.conf.QueueBaseUrl + hangupPath
		_, err = httpclient.Post(hangupPayload, hangupUrl, map[string]string{ContentType: ContentTypeJSON})
		if err != nil {
			log.Println(err)
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
