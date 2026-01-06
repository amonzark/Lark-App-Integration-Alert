package lark

import (
	"context"
	"encoding/json"
	"fmt"

	"log"
	"log/slog"
	"net/http"
	"os"
	"time"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkcontact "github.com/larksuite/oapi-sdk-go/v3/service/contact/v3"
	larkim "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
	"github.com/prometheus/alertmanager/api/v2/models"

	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/internal/alertmanager"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg/model"
)

func (l *Lark) NotifyAlerts(webhook model.Webhook) error {
	for _, alert := range webhook.Alerts {
		if err := l.notifyAlert(alert, webhook.Channel); err != nil {
			return err
		}
	}
	return nil
}

func (l *Lark) notifyAlert(alert model.WebhookAlert, channel string) error {
	content, err := l.cardBuilder.BuildJSON(&alert)
	if err != nil {
		slog.Error(err.Error())
		return err
	}

	if err := l.sendAlert(alert, channel, content); err != nil {
		slog.Error(err.Error())
		return err
	}

	return nil
}

func (l *Lark) sendAlert(alert model.WebhookAlert, channel, content string) error {
	logger := l.logger.With(
		slog.String("alert_id", alert.CallbackID),
		slog.String("chat_id", channel),
	)

	if alert.Color == "green" {
		messageID, err := l.repository.GetMessageID(alert.CallbackID)
		logger.With(
			slog.String("message_id", messageID),
		)
		if err != nil {
			logger.Warn("failed to get message id from repository [fallback to create message]",
				slog.String("error", err.Error()),
			)
			err, _ := l.sendAlertMessage(alert.CallbackID, channel, content)
			return err
		}

		if err := l.sendResolvedMessage(messageID, content, alert.CallbackID, channel); err != nil {
			return err
		}
		logger.Debug("deleting message id from repository")
		if err := l.repository.DeleteMessageID(messageID); err != nil {
			logger.Warn("failed to delete reply message request",
				slog.String("error", err.Error()),
			)
		}
	} else {
		err, messageID := l.sendAlertMessage(alert.CallbackID, channel, content)
		if err != nil {
			return err
		}
		logger = logger.With(
			slog.String("message_id", *messageID),
		)
		logger.Debug("saving message id to repository")
		if err := l.repository.SetMessageID(alert.CallbackID, *messageID); err != nil {
			logger.Warn("failed to save message id to repository",
				slog.String("error", err.Error()),
			)
		}
	}

	return nil
}

func (l *Lark) sendAlertMessage(alertID, channel, content string) (error, *string) {
	logger := l.logger.With(
		slog.String("alert_id", alertID),
		slog.String("chat_id", channel),
	)
	logger.Info("sending create alert message request")

	logger.Debug("creating new create message request")
	req := larkim.NewCreateMessageReqBuilder().
		ReceiveIdType("chat_id").
		Body(larkim.NewCreateMessageReqBodyBuilder().
			ReceiveId(channel).
			MsgType("interactive").
			Content(content).
			Build()).
		Build()

	logger.Debug("sending create message request")
	resp, err := l.client.Im.Message.Create(context.Background(), req)
	if err != nil {
		logger.Error("failed to send create message request",
			slog.String("error", err.Error()),
		)
		// TODO: create retry fallback mechanism
		return err, nil
	}
	if !resp.Success() {
		logger.Error("failed to send create message request",
			slog.String("logId", resp.RequestId()),
			slog.String("error_response", larkcore.Prettify(resp.CodeError)),
		)
		return err, nil
	}

	return nil, resp.Data.MessageId
}

func (l *Lark) sendResolvedMessage(messageID, content, alertID, channel string) error {
	logger := l.logger.With(
		slog.String("message_id", messageID),
		slog.String("alert_id", alertID),
		slog.String("chat_id", channel),
	)
	logger.Info("sending reply resolved message request")

	messageID, err := l.repository.GetMessageID(alertID)
	if err != nil {
		logger.Warn("failed to get message id from repository [fallback to create message]",
			slog.String("error", err.Error()),
		)
		err, _ := l.sendAlertMessage(alertID, channel, content)
		return err
	}
	logger = l.logger.With(
		slog.String("message_id", messageID),
	)

	logger.Debug("creating new reply message request")
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType("interactive").
			Content(content).
			ReplyInThread(true).
			Build()).
		Build()

	logger.Debug("sending reply message request")
	resp, err := l.client.Im.Message.Reply(context.Background(), req)
	if err != nil {
		logger.Error("failed to send reply message request",
			slog.String("error", err.Error()),
		)
		// TODO: create retry fallback mechanism
		return err
	}
	if !resp.Success() {
		logger.Error("failed to send reply message request",
			slog.String("logId", resp.RequestId()),
			slog.String("error_response", larkcore.Prettify(resp.CodeError)),
		)
		return err
	}

	return nil
}

// silence
func NewHandler(alertmanager alertmanager.Alertmanager) *Handler {
	return &Handler{
		alertmanager: alertmanager,
	}
}

type EventSilence interface {
	HandleCreateSilence(alertID string, email string, duration string) error
}

type Handler struct {
	alertmanager alertmanager.Alertmanager
}

func (h *Handler) HandleCreateSilence(alertID string, email string, duration string) error {
	alerts, err := h.alertmanager.GetAlertsByFingerprints([]string{alertID})
	if err != nil {
		slog.Info("alertID is not found ", "alertID: ", alertID)
		slog.Error("Failed to get alerts by fingerprints", "ERROR: ", err)
		return err
	}
	//check alert is fetched or not
	if alerts == nil || len(*alerts) == 0 {
		slog.Info("No alerts found for the given fingerprint")
		return fmt.Errorf("no alerts found for alertID: %s", alertID)
	}

	for _, alert := range *alerts {
		if name, ok := alert.Labels["alertname"]; ok {
			slog.Info("Alert Name ", "name: ", name)
		} else {
			slog.Error("Alert Name not found in labels")
		}
	}
	// create matchers
	matchers := []*models.Matcher{}
	for _, alert := range *alerts {
		for key, value := range alert.Labels {
			matchers = append(matchers, &models.Matcher{
				Name:    pointerString(key),
				Value:   pointerString(value),
				IsRegex: pointerBool(false),
			})
		}
	}

	//create silence
	startsAt := time.Now()
	//check and set duration
	var endsAt time.Time
	switch duration {
	case "", "default":
		endsAt = startsAt.Add(6 * time.Hour)
		slog.Info("Silence duration is not set, using default 6 hours")
	case "30m":
		endsAt = startsAt.Add(30 * time.Minute)
		slog.Info("Silence duration set to 30 minutes")
	case "1h":
		endsAt = startsAt.Add(1 * time.Hour)
		slog.Info("Silence duration set to 1 hour")
	case "3h":
		endsAt = startsAt.Add(3 * time.Hour)
		slog.Info("Silence duration set to 3 hours")
	case "6h":
		endsAt = startsAt.Add(6 * time.Hour)
		slog.Info("Silence duration set to 6 hours")
	case "12h":
		endsAt = startsAt.Add(12 * time.Hour)
		slog.Info("Silence duration set to 12 hours")
	case "1d":
		endsAt = startsAt.Add(24 * time.Hour)
		slog.Info("Silence duration set to 1 day")
	case "3d":
		endsAt = startsAt.Add((24 * 3) * time.Hour)
		slog.Info("Silence duration set to 3 days")
	case "1w":
		endsAt = startsAt.Add((24 * 7) * time.Hour)
		slog.Info("Silence duration set to 1 week")
	case "3w":
		endsAt = startsAt.Add((24 * 7 * 3) * time.Hour)
		slog.Info("Silence duration set to 3 weeks")
	case "1M":
		endsAt = startsAt.Add((24 * 30) * time.Hour)
		slog.Info("Silence duration set to 1 month")
	case "1Y":
		endsAt = startsAt.Add((24 * 365) * time.Hour)
		slog.Info("Silence duration set to 1 year")
	default:
		slog.Error("Invalid silence duration specified ", "duration: ", duration)
		return fmt.Errorf("invalid silence duration: %s", duration)
	}

	comment := "create silence"

	silenceID, err := h.alertmanager.Silence(comment, email, matchers, startsAt, endsAt)
	slog.Info("Creating Silence ID: ", "silenceID: ", silenceID, ", startsAt: ", startsAt, ", endsAt: ", endsAt)
	if err != nil {
		slog.Error("Failed to post silence: ", "ERROR: ", err)
	}
	slog.Info("Successfully created silence ID: %s\n", "silenceID: ", silenceID)
	return nil
}

func pointerString(s string) *string {
	return &s
}
func pointerBool(s bool) *bool {
	return &s
}

func (l *Lark) GetUserInfo(openID string) (*larkcontact.User, error) {
	//create new request to get user info
	req := larkcontact.NewGetUserReqBuilder().
		UserId(openID).
		UserIdType(`open_id`).
		Build()

	//send request
	resp, err := l.client.Contact.V3.User.Get(context.Background(), req)
	if err != nil {
		slog.Error("API call failed to get user info", "ERROR: ", err)
		return nil, err
	}
	//return user info
	return resp.Data.User, nil
}

func (l *Lark) ResponseError(w http.ResponseWriter, errors map[string]string) error {
	w.Header().Add("Content-type", "application/json")

	errorJSON, err := json.Marshal(errors)
	if err != nil {
		slog.Error("Failed to marshall view submission resposne: %v", "ERROR: ", err)
		return err
	}
	body := fmt.Sprintf(`{"response_action": "errors", "errors": %s}`, errorJSON)
	w.Write([]byte(body))
	log.Println(body)
	return nil
}

func SendChallengeResponse(w http.ResponseWriter, r http.Request, veriftoken string, challenge string) {
	if veriftoken != os.Getenv("VERIFICATION_TOKEN") {
		slog.Error("Invalid verification token", "received", veriftoken)
		http.Error(w, "Unauthorized: invalid verification token", http.StatusUnauthorized)
		return
	}
	// Respond with the challenge value
	response := map[string]string{
		"challenge": challenge,
	}
	// Send the challenge back to Lark Open Platform
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// response after create silence
func (l *Lark) SendResponseCreatedSilence(message_id string, chat_id string, text string) error {
	// Create a new reply message request
	req := larkim.NewReplyMessageReqBuilder().
		MessageId(message_id).
		Body(larkim.NewReplyMessageReqBodyBuilder().
			MsgType("text").
			Content(fmt.Sprintf(`{"text": "%s"}`, text)).
			Build()).
		Build()

	// Send the reply message
	resp, err := l.client.Im.Message.Reply(context.Background(), req)
	if err != nil {
		slog.Error("Failed to send reply message", "error", err)
		return err
	}

	// Check if the response was successful
	if !resp.Success() {
		slog.Error("Failed to send reply message",
			"logId", resp.RequestId(),
			"error_response", larkcore.Prettify(resp.CodeError),
		)
		return fmt.Errorf("failed to send reply message: %s", larkcore.Prettify(resp.CodeError))
	}

	return nil
}

func WriteToast(w http.ResponseWriter, toastType, content string) {
	resp := model.CallbackResponse{
		Toast: &model.Toast{
			Type:    toastType,
			Content: content,
		},
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("Failed to encode toast response ", "ERROR: ", err)
	}
}
