package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"

	"time"

	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/internal/alertmanager"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/internal/lark"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/internal/o11y"
	"source.golabs.io/cloud-platform/observability/katulampa/katulampa-lark-app/pkg/model"
)

func (s *Server) Start() error {
	http.HandleFunc("/ping", s.pingHandler)
	http.HandleFunc("/notify", s.notifyHandler)
	http.HandleFunc("/callback", s.HandleCallback)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", s.port), nil); err != nil {
		return err
	}

	return nil
}

func (s *Server) pingHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func (s *Server) notifyHandler(w http.ResponseWriter, r *http.Request) {
	var webhook model.Webhook

	if err := json.NewDecoder(r.Body).Decode(&webhook); err != nil {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	err := s.notifier.NotifyAlerts(webhook)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}
	o11y.IncreasePostToLarkCounter(webhook.Channel, "webhook", err == nil)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

// HandleCallback handles the callback request and verifies the signature.
func (s *Server) HandleCallback(w http.ResponseWriter, r *http.Request) {
	// Read the request body
	payloadBytes, err := io.ReadAll(r.Body)
	if err != nil {
		slog.Error("error reading payload", "ERROR: ", err)
		lark.WriteToast(w, "error", "Failed to read request body")
		slog.Error("failed to read request body %v", "ERROR: ", err)
		return
	}
	lark.WriteToast(w, "info", "Request received, processing...")

	go func(payloadBytes []byte) {
		var payload model.URLVerificationRequest
		if err := json.Unmarshal(payloadBytes, &payload); err != nil {
			slog.Error("not url verification", "ERROR: ", err)
		}

		if payload.Type == "url_verification" {
			lark.SendChallengeResponse(w, *r, payload.Token, payload.Challenge)
		} else if payload.Type != "url_verification" {
			//read body request
			var payload_event model.CardActionPayload
			if err := json.Unmarshal(payloadBytes, &payload_event); err != nil {
				slog.Info("not event_type")
				slog.Error("failed to unmarshal webhook payload_event", "ERROR: ", err)
				return
			}

			// Process silence creation for both select_static (dropdown) and button actions
			if payload_event.Event.Action.Tag == "select_static" || payload_event.Event.Action.Tag == "button" {
				var duration string
				slog.Info("Raw payload: " + string(payloadBytes))

				if payload_event.Event.Action.Tag == "select_static" {
					duration = payload_event.Event.Action.Option
					slog.Info("Dropdown action detected, duration set to ", "duration: ", duration)
					slog.Info("Dropdown callback payload", "payload_event", payload_event)
				} else {
					duration = "default"
					slog.Info("Button action detected, using default duration")
				}
				alert_id := strings.TrimSuffix(payload_event.Event.Action.Value.AlertID, ",")
				open_id := payload_event.Event.Operator.OpenID
				user, err := s.notifier.GetUserInfo(open_id)
				if err != nil {
					slog.Error("Failed to get user info", "ERROR: ", err)
					return
				}
				email := *user.Email
				am := alertmanager.NewAlertmanager(s.alertmanagerHost)
				silence := lark.NewHandler(am)
				slog.Info("request silence by ", "open_id: ", open_id, ", email: ", email, ", and alert_id: ", alert_id)
				if err := silence.HandleCreateSilence(alert_id, email, duration); err != nil {
					slog.Error("error failed to creating silence ", "ERROR: ", err)
					return
				}

				//reply message confirmation silence created
				messageID := payload_event.Event.Context.OpenMessageID
				chatID := payload_event.Event.Context.OpenChatID

				// endtime add 7 hours for convert it into WIB and another 6 hours for silence duration
				//check value of duration
				var endtime time.Time
				switch duration {
				case "", "default":
					endtime = time.Now().Add(6*time.Hour + 7*time.Hour)
					slog.Info("Silence duration is not set, using default 6 hours")
				case "30m":
					endtime = time.Now().Add(30*time.Minute + 7*time.Hour) // add 7 hours for convert it into WIB
					slog.Info("Silence duration set to 30 minutes")
				case "1h":
					endtime = time.Now().Add(1*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 1 hour")
				case "3h":
					endtime = time.Now().Add(3*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 3 hours")
				case "6h":
					endtime = time.Now().Add(6*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 6 hours")
				case "12h":
					endtime = time.Now().Add(12*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 12 hours")
				case "1d":
					endtime = time.Now().Add(24*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 1 day")
				case "3d":
					endtime = time.Now().Add((24*3)*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 3 days")
				case "1w":
					endtime = time.Now().Add((24*7)*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 1 week")
				case "3w":
					endtime = time.Now().Add((24*7*3)*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 3 weeks")
				case "1M":
					endtime = time.Now().Add((24*30)*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 1 month")
				case "1Y":
					endtime = time.Now().Add((24*365)*time.Hour + 7*time.Hour)
					slog.Info("Silence duration set to 1 year")
				default:
					slog.Error("Invalid silence duration specified ", "duration: ", duration)
					return
				}
				endtimestr := endtime.Format("2006-01-02 15:04:05") + " WIB"
				text := fmt.Sprintf("Silence created successfully. It will expire at %s by %s.", endtimestr, email)
				text_failed := fmt.Sprintf("Failed to create silence for alert %s by %s.", alert_id, email)
				slog.Info("sending silence response message by ", "messageID: ", messageID, ", chatID: ", chatID, ", text: ", text)
				if err := s.notifier.SendResponseCreatedSilence(messageID, chatID, text); err != nil {
					slog.Error("Failed to send response to Lark", "ERROR: ", err)
					s.notifier.SendResponseCreatedSilence(messageID, chatID, text_failed)
				}
			} else {
				slog.Info("non-silence action detected, ignoring callback")
				return
			}
		} else {
			slog.Error("missing or incorrect 'type' or 'event_type' field in payload")
			slog.Info("incorrect field")
			//lark.WriteToast(w, "error", "Missing or incorrect 'type' or 'event_type' field in payload")
			return
		}

	}(payloadBytes)
}
