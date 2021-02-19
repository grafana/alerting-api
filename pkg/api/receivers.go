package api

import (
	"github.com/grafana/grafana/pkg/models"
	"github.com/prometheus/alertmanager/config"
)

// swagger:route POST /api/v1/alerts AlertingConfig RoutePostAlertingConfig
//
// sets an Alerting config
//
//     Responses:
//       201: Ack
//       400: ValidationError

// swagger:route GET /api/v1/alerts AlertingConfig RouteGetAlertingConfig
//
// gets an Alerting config
//
//     Responses:
//       200: AlertingConfigResponse
//       400: ValidationError

// swagger:route DELETE /api/v1/alerts AlertingConfig RouteDeleteAlertingConfig
//
// deletes the Alerting config for a tenant
//
//     Responses:
//       200: Ack
//       400: ValidationError

// swagger:parameters RoutePostAlertingConfig
type BodyAlertingConfig struct {
	// in:body
	Body UserConfig
}

// swagger:model
type UserConfig struct {
	TemplateFiles      map[string]string `yaml:"template_files" json:"template_files"`
	AlertmanagerConfig ApiAlertingConfig `yaml:"alertmanager_config" json:"alertmanager_config"`
}

// swagger:model
type AlertingConfigResponse struct {
	BodyAlertingConfig
}

type ApiAlertingConfig struct {
	config.Config

	// Override with our superset receiver type
	Receivers []*ApiReceiver `yaml:"receivers,omitempty" json:"receivers,omitempty"`
}

type GrafanaReceiver models.CreateAlertNotificationCommand

type ApiReceiver struct {
	config.Receiver
	GrafanaManagedReceivers []*GrafanaReceiver `yaml:"grafana_managed_receiver_configs,omitempty" json:"grafana_managed_receiver_configs,omitempty"`
}
