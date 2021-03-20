package api

import (
	"encoding/json"
	"fmt"

	"github.com/grafana/grafana/pkg/api/dtos"
	"github.com/grafana/grafana/pkg/models"
	amv2 "github.com/prometheus/alertmanager/api/v2/models"
	"github.com/prometheus/alertmanager/config"
)

// swagger:route POST /alertmanager/{DatasourceId}/config/api/v1/alerts alertmanager RoutePostAlertingConfig
//
// sets an Alerting config
//
//     Responses:
//       201: Ack
//       400: ValidationError

// swagger:route GET /alertmanager/{DatasourceId}/config/api/v1/alerts alertmanager RouteGetAlertingConfig
//
// gets an Alerting config
//
//     Responses:
//       200: GettableUserConfig
//       400: ValidationError

// swagger:route DELETE /alertmanager/{DatasourceId}/config/api/v1/alerts alertmanager RouteDeleteAlertingConfig
//
// deletes the Alerting config for a tenant
//
//     Responses:
//       200: Ack
//       400: ValidationError

// swagger:route GET /alertmanager/{DatasourceId}/api/v2/alerts alertmanager RouteGetAmAlerts
//
// get alertmanager alerts
//
//     Responses:
//       200: GettableAlerts
//       400: ValidationError

// swagger:route GET /alertmanager/{DatasourceId}/api/v2/alerts/groups alertmanager RouteGetAmAlertGroups
//
// get alertmanager alerts
//
//     Responses:
//       200: AlertGroups
//       400: ValidationError

// swagger:route GET /alertmanager/{DatasourceId}/api/v2/silences alertmanager RouteGetSilences
//
// get silences
//
//     Responses:
//       200: GettableSilences
//       400: ValidationError

// swagger:route POST /alertmanager/{DatasourceId}/api/v2/silences alertmanager RouteCreateSilence
//
// create silence
//
//     Responses:
//       201: GettableSilence
//       400: ValidationError

// swagger:route GET /alertmanager/{DatasourceId}/api/v2/silence/{SilenceId} alertmanager RouteGetSilence
//
// get silence
//
//     Responses:
//       200: GettableSilence
//       400: ValidationError

// swagger:route DELETE /alertmanager/{DatasourceId}/api/v2/silence/{SilenceId} alertmanager RouteDeleteSilence
//
// delete silence
//
//     Responses:
//       200: Ack
//       400: ValidationError

// swagger:parameters RouteCreateSilence
type CreateSilenceParams struct {
	// in:body
	Body SilenceBody
}

//swagger:parameters RouteGetSilence RouteDeleteSilence
type GetDeleteSilenceParams struct {
	// in:path
	SilenceId string
}

// swagger:model
type SilenceBody struct {
	Id string `json:"id"`
	amv2.Silence
}

// swagger:model
type GettableSilences amv2.GettableSilences

// swagger:model
type GettableSilence amv2.GettableSilence

// swagger:model
type GettableAlerts amv2.GettableAlerts

// swagger:model
type AlertGroups amv2.AlertGroups

// swagger:parameters RouteGetAmAlerts RouteGetAmAlertGroups
type AlertsParams struct {

	// Show active alerts
	// in: query
	// required: false
	Active bool `json:"active"`

	// Show silenced alerts
	// in: query
	// required: false
	Silenced bool `json:"silenced"`

	// Show inhibited alerts
	// in: query
	// required: false
	Inhibited bool `json:"inhibited"`

	// A list of matchers to filter alerts by
	// in: query
	// required: false
	Matchers []string `json:"matchers"`

	// A list of receivers to filter alerts by
	// in: query
	// required: false
	Receivers []string `json:"receivers"`
}

// swagger:parameters RoutePostAlertingConfig
type BodyAlertingConfig struct {
	// in:body
	Body PostableUserConfig
}

// alertmanager routes
// swagger:parameters RoutePostAlertingConfig RouteGetAlertingConfig RouteDeleteAlertingConfig RouteGetAmAlerts RouteGetAmAlertGroups RouteGetSilences RouteCreateSilence RouteGetSilence RouteDeleteSilence RoutePostAlertingConfig
// ruler routes
// swagger:parameters RouteGetRulesConfig RoutePostNameRulesConfig RouteGetNamespaceRulesConfig RouteDeleteNamespaceRulesConfig RouteGetRulegGroupConfig RouteDeleteRuleGroupConfig
// prom routes
// swagger:parameters RouteGetRuleStatuses RouteGetAlertStatuses
type DatasourceReference struct {
	// in:path
	DatasourceId string
}

// swagger:model
type PostableUserConfig struct {
	TemplateFiles      map[string]string         `yaml:"template_files" json:"template_files"`
	AlertmanagerConfig PostableApiAlertingConfig `yaml:"alertmanager_config" json:"alertmanager_config"`
}

// swagger:model
type GettableUserConfig struct {
	TemplateFiles      map[string]string         `yaml:"template_files" json:"template_files"`
	AlertmanagerConfig GettableApiAlertingConfig `yaml:"alertmanager_config" json:"alertmanager_config"`
}
type GettableApiAlertingConfig struct {
	config.Config

	// Override with our superset receiver type
	Receivers []*GettableApiReceiver `yaml:"receivers,omitempty" json:"receivers,omitempty"`
}

func (c *GettableApiAlertingConfig) UnmarshalJSON(b []byte) error {
	type plain GettableApiAlertingConfig
	if err := json.Unmarshal(b, (*plain)(c)); err != nil {
		return err
	}

	return c.validate()
}

// validate ensures that the two routing trees use the correct receiver types.
func (c *GettableApiAlertingConfig) validate() error {
	receivers := make(map[string]struct{}, len(c.Receivers))

	var hasGrafReceivers, hasAMReceivers bool
	for _, r := range c.Receivers {
		receivers[r.Name] = struct{}{}
		switch r.Type() {
		case GrafanaReceiverType:
			hasGrafReceivers = true
		case AlertmanagerReceiverType:
			hasAMReceivers = true
		}
	}

	if hasGrafReceivers && hasAMReceivers {
		return fmt.Errorf("cannot mix Alertmanager & Grafana receiver types")
	}

	for _, receiver := range AllReceivers(c.Route) {
		_, ok := receivers[receiver]
		if !ok {
			return fmt.Errorf("unexpected receiver (%s) is undefined", receiver)
		}
	}

	return nil
}

// Type requires validate has been called and just checks the first receiver type
func (c *GettableApiAlertingConfig) Type() (backend Backend) {
	for _, r := range c.Receivers {
		switch r.Type() {
		case GrafanaReceiverType:
			return GrafanaBackend
		case AlertmanagerReceiverType:
			return AlertmanagerBackend
		}
	}
	return
}

type PostableApiAlertingConfig struct {
	config.Config

	// Override with our superset receiver type
	Receivers []*PostableApiReceiver `yaml:"receivers,omitempty" json:"receivers,omitempty"`
}

func (c *PostableApiAlertingConfig) UnmarshalJSON(b []byte) error {
	type plain PostableApiAlertingConfig
	if err := json.Unmarshal(b, (*plain)(c)); err != nil {
		return err
	}

	return c.validate()
}

// validate ensures that the two routing trees use the correct receiver types.
func (c *PostableApiAlertingConfig) validate() error {
	receivers := make(map[string]struct{}, len(c.Receivers))

	var hasGrafReceivers, hasAMReceivers bool
	for _, r := range c.Receivers {
		receivers[r.Name] = struct{}{}
		switch r.Type() {
		case GrafanaReceiverType:
			hasGrafReceivers = true
		case AlertmanagerReceiverType:
			hasAMReceivers = true
		}
	}

	if hasGrafReceivers && hasAMReceivers {
		return fmt.Errorf("cannot mix Alertmanager & Grafana receiver types")
	}

	for _, receiver := range AllReceivers(c.Route) {
		_, ok := receivers[receiver]
		if !ok {
			return fmt.Errorf("unexpected receiver (%s) is undefined", receiver)
		}
	}

	return nil
}

// Type requires validate has been called and just checks the first receiver type
func (c *PostableApiAlertingConfig) Type() (backend Backend) {
	for _, r := range c.Receivers {
		switch r.Type() {
		case GrafanaReceiverType:
			return GrafanaBackend
		case AlertmanagerReceiverType:
			return AlertmanagerBackend
		}
	}
	return
}

// AllReceivers will recursively walk a routing tree and return a list of all the
// referenced receiver names.
func AllReceivers(route *config.Route) (res []string) {
	res = append(res, route.Receiver)
	for _, subRoute := range route.Routes {
		res = append(res, AllReceivers(subRoute)...)
	}
	return res
}

type GettableGrafanaReceiver dtos.AlertNotification
type PostableGrafanaReceiver models.CreateAlertNotificationCommand

type ReceiverType int

const (
	GrafanaReceiverType ReceiverType = iota
	AlertmanagerReceiverType
)

type GettableApiReceiver struct {
	config.Receiver
	GettableGrafanaReceivers
}

func (r *GettableApiReceiver) UnmarshalJSON(b []byte) error {
	type plain GettableApiReceiver
	if err := json.Unmarshal(b, (*plain)(r)); err != nil {
		return err
	}

	hasGrafanaReceivers := len(r.GettableGrafanaReceivers.GrafanaManagedReceivers) > 0

	if hasGrafanaReceivers {
		if len(r.EmailConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager EmailConfigs & Grafana receivers together")
		}
		if len(r.PagerdutyConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager PagerdutyConfigs & Grafana receivers together")
		}
		if len(r.SlackConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager SlackConfigs & Grafana receivers together")
		}
		if len(r.WebhookConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager WebhookConfigs & Grafana receivers together")
		}
		if len(r.OpsGenieConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager OpsGenieConfigs & Grafana receivers together")
		}
		if len(r.WechatConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager WechatConfigs & Grafana receivers together")
		}
		if len(r.PushoverConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager PushoverConfigs & Grafana receivers together")
		}
		if len(r.VictorOpsConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager VictorOpsConfigs & Grafana receivers together")
		}

	}

	return nil

}

func (r *GettableApiReceiver) Type() ReceiverType {
	if len(r.GettableGrafanaReceivers.GrafanaManagedReceivers) > 0 {
		return GrafanaReceiverType
	}
	return AlertmanagerReceiverType
}

type PostableApiReceiver struct {
	config.Receiver
	PostableGrafanaReceivers
}

func (r *PostableApiReceiver) UnmarshalJSON(b []byte) error {
	type plain PostableApiReceiver
	if err := json.Unmarshal(b, (*plain)(r)); err != nil {
		return err
	}

	hasGrafanaReceivers := len(r.PostableGrafanaReceivers.GrafanaManagedReceivers) > 0

	if hasGrafanaReceivers {
		if len(r.EmailConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager EmailConfigs & Grafana receivers together")
		}
		if len(r.PagerdutyConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager PagerdutyConfigs & Grafana receivers together")
		}
		if len(r.SlackConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager SlackConfigs & Grafana receivers together")
		}
		if len(r.WebhookConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager WebhookConfigs & Grafana receivers together")
		}
		if len(r.OpsGenieConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager OpsGenieConfigs & Grafana receivers together")
		}
		if len(r.WechatConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager WechatConfigs & Grafana receivers together")
		}
		if len(r.PushoverConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager PushoverConfigs & Grafana receivers together")
		}
		if len(r.VictorOpsConfigs) > 0 {
			return fmt.Errorf("cannot have both Alertmanager VictorOpsConfigs & Grafana receivers together")
		}

	}

	return nil

}

func (r *PostableApiReceiver) Type() ReceiverType {
	if len(r.PostableGrafanaReceivers.GrafanaManagedReceivers) > 0 {
		return GrafanaReceiverType
	}
	return AlertmanagerReceiverType
}

type GettableGrafanaReceivers struct {
	GrafanaManagedReceivers []*GettableGrafanaReceiver `yaml:"grafana_managed_receiver_configs,omitempty" json:"grafana_managed_receiver_configs,omitempty"`
}

type PostableGrafanaReceivers struct {
	GrafanaManagedReceivers []*PostableGrafanaReceiver `yaml:"grafana_managed_receiver_configs,omitempty" json:"grafana_managed_receiver_configs,omitempty"`
}
