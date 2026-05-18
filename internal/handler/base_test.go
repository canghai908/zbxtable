package handler

import (
	"testing"
	"zbxtable/internal/model"
)

func TestReceiveWebhookTokenValidation(t *testing.T) {
	tests := []struct {
		name     string
		instance *model.ZabbixInstance
		token    string
		wantOK   bool
	}{
		{
			name: "accept matching webhook token",
			instance: &model.ZabbixInstance{
				Token:        "zabbix-api-token",
				WebhookToken: "webhook-token",
			},
			token:  "webhook-token",
			wantOK: true,
		},
		{
			name: "reject legacy api token",
			instance: &model.ZabbixInstance{
				Token:        "zabbix-api-token",
				WebhookToken: "webhook-token",
			},
			token:  "zabbix-api-token",
			wantOK: false,
		},
		{
			name: "reject when webhook token missing",
			instance: &model.ZabbixInstance{
				Token: "zabbix-api-token",
			},
			token:  "zabbix-api-token",
			wantOK: false,
		},
		{
			name: "reject empty request token",
			instance: &model.ZabbixInstance{
				WebhookToken: "webhook-token",
			},
			token:  "",
			wantOK: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOK := tt.token != "" && tt.instance.WebhookToken != "" && tt.token == tt.instance.WebhookToken
			if gotOK != tt.wantOK {
				t.Fatalf("webhook token validation = %v, want %v", gotOK, tt.wantOK)
			}
		})
	}
}
