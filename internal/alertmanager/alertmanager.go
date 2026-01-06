package alertmanager

import (
	"log/slog"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/prometheus/alertmanager/api/v2/client"
	"github.com/prometheus/alertmanager/api/v2/client/alert"
	"github.com/prometheus/alertmanager/api/v2/client/silence"
	"github.com/prometheus/alertmanager/api/v2/models"
)

type Alertmanager interface {
	GetAlerts(param *alert.GetAlertsParams) (*alert.GetAlertsOK, error)
	Silence(comment, createdBy string, matchers []*models.Matcher, startAt time.Time, endsAt time.Time) (string, error)
	GetSilence(silenceID string) (*models.GettableSilence, error)
	GetAlertsByFingerprints(fingerprints []string) (*models.GettableAlerts, error)
}

type alertmanager struct {
	client *client.AlertmanagerAPI
}

func NewAlertmanager(hostname string) *alertmanager {
	alertClient := client.NewHTTPClientWithConfig(nil, &client.TransportConfig{
		Host:     hostname,
		BasePath: "/api/v2",
		Schemes:  client.DefaultSchemes,
	})
	if alertClient == nil {
		slog.Error("Failed to create Alertmanager client")
	} else {
		slog.Info("Successfully created Alertmanager client", "hostname", hostname)
	}
	return &alertmanager{
		client: alertClient,
	}
}

func (am *alertmanager) GetAlerts(param *alert.GetAlertsParams) (*alert.GetAlertsOK, error) {
	return am.client.Alert.GetAlerts(param)
}

func (am *alertmanager) GetSilence(silenceID string) (*models.GettableSilence, error) {
	param := silence.NewGetSilenceParams()
	param.SilenceID = strfmt.UUID(silenceID)
	resp, err := am.client.Silence.GetSilence(param)
	if err != nil {
		return nil, err
	}
	return resp.GetPayload(), nil
}

func (am *alertmanager) Silence(comment, createdBy string, matchers []*models.Matcher, startAt time.Time, endsAt time.Time) (string, error) {
	startAtStr := strfmt.DateTime(startAt)
	endsAtStr := strfmt.DateTime(endsAt)
	for _, matcher := range matchers {
		slog.Info("Matcher Details silence",
			"Name", matcher.Name,
			"Value", matcher.Value,
			"IsRegex", matcher.IsRegex,
		)
	}
	silenceObj := models.PostableSilence{
		Silence: models.Silence{
			Comment:   &comment,
			CreatedBy: &createdBy,
			Matchers:  matchers,
			StartsAt:  &startAtStr,
			EndsAt:    &endsAtStr,
		},
	}
	params := silence.NewPostSilencesParams()
	params.SetSilence(&silenceObj)
	resp, err := am.client.Silence.PostSilences(params)
	if err != nil {
		slog.Error("Error post silences: %v", "ERROR: ", err)
		return "", err
	}
	return resp.GetPayload().SilenceID, nil
}

func (am *alertmanager) GetAlertsByFingerprints(fingerprints []string) (*models.GettableAlerts, error) {
	alerts := models.GettableAlerts{}

	resp, err := am.GetAlerts(nil)
	if err != nil {
		slog.Error("Failed to fetch alerts from alertmanager %v", "ERROR: ", err)
		return nil, err
	}
	if resp == nil || len(resp.Payload) == 0 {
		slog.Error("No alerts found.")
		return &alerts, nil
	}
	target := fingerprints[0]
	for _, alert := range resp.Payload {
		if alert.Fingerprint != nil && *alert.Fingerprint == target {
			alerts = append(alerts, alert)
			slog.Info("Alert fingerprint found %s", "fingerprint: ", target)
			slog.Info("Details of ", "alert: ", alert)
			break // Found it, no need to continue
		}
	}
	if len(alerts) == 0 {
		slog.Info("Alert fingerprint not found %s", "fingerprint: ", target)
	}
	return &alerts, nil
}
