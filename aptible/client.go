package aptible

import (
	"context"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/aptible-api-go/helpers"
)

type client struct {
	*aptibleapi.APIClient
	Token   string
	APIRoot string
}

func newClient(token, apiRoot string) *client {
	cfg := aptibleapi.NewAPIConfiguration()
	cfg.Servers = aptibleapi.ServerConfigurations{{
		URL:         apiRoot,
		Description: "Aptible API",
	}}
	cfg.AddDefaultHeader("Authorization", "Bearer "+token)

	return &client{
		APIClient: aptibleapi.NewAPIClient(cfg),
		Token:     token,
		APIRoot:   apiRoot,
	}
}

func (m *client) WaitForOperation(ctx context.Context, operationID int32) (bool, error) {
	return helpers.WaitForOperation(ctx, m.APIClient, operationID)
}

func (m *client) DeleteApp(ctx context.Context, appID int32) (bool, error) {
	return helpers.DeleteApp(ctx, m.APIClient, appID)
}

func (m *client) DeleteDatabase(ctx context.Context, databaseID int32) (bool, error) {
	return helpers.DeleteDatabase(ctx, m.APIClient, databaseID)
}

func (m *client) DeleteEndpoint(ctx context.Context, vhostID int32) (bool, error) {
	return helpers.DeleteEndpoint(ctx, m.APIClient, vhostID)
}

func (m *client) DeleteLogDrain(ctx context.Context, logDrainID int32) (bool, error) {
	return helpers.DeleteLogDrain(ctx, m.APIClient, logDrainID)
}

func (m *client) DeleteMetricDrain(ctx context.Context, metricDrainID int32) (bool, error) {
	return helpers.DeleteMetricDrain(ctx, m.APIClient, metricDrainID)
}

func (m *client) GetStackByName(ctx context.Context, name string) (*aptibleapi.Stack, error) {
	return helpers.GetStackByName(ctx, m.APIClient, name)
}

func (m *client) GetDatabaseImageByTypeAndVersion(ctx context.Context, imageType, version string) (*aptibleapi.DatabaseImage, error) {
	return helpers.GetDatabaseImageByTypeAndVersion(ctx, m.APIClient, imageType, version)
}

func (m *client) GetServiceForAppByName(ctx context.Context, appID int32, processType string) (*aptibleapi.Service, error) {
	return helpers.GetServiceForAppByName(ctx, m.APIClient, appID, processType)
}

func (m *client) GetReplicaByHandle(ctx context.Context, databaseID int32, handle string) (*aptibleapi.Database, error) {
	return helpers.GetReplicaByHandle(ctx, m.APIClient, databaseID, handle)
}

func (m *client) GetOrganizationFromAuthAPI() (string, error) {
	return helpers.GetOrganizationFromAuthAPI(m.Token, helpers.GetAuthURL())
}

func (m *client) GetStackOrganizationID(stack *aptibleapi.Stack) string {
	return helpers.GetStackOrganizationID(stack)
}
