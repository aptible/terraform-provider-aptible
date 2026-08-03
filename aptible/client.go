package aptible

import (
	"context"

	"github.com/aptible/aptible-api-go/aptibleapi"
	"github.com/aptible/aptible-api-go/helpers"
)

type providerMetadata struct {
	*aptibleapi.APIClient
	Token   string
	APIRoot string
}

func newProviderMetadata(token, apiRoot string) *providerMetadata {
	cfg := aptibleapi.NewAPIConfiguration()
	cfg.Servers = aptibleapi.ServerConfigurations{{
		URL:         apiRoot,
		Description: "Aptible API",
	}}
	cfg.AddDefaultHeader("Authorization", "Bearer "+token)

	return &providerMetadata{
		APIClient: aptibleapi.NewAPIClient(cfg),
		Token:     token,
		APIRoot:   apiRoot,
	}
}

func (m *providerMetadata) WaitForOperation(ctx context.Context, operationID int32) (bool, error) {
	return helpers.WaitForOperation(ctx, m.APIClient, operationID)
}

func (m *providerMetadata) DeleteApp(ctx context.Context, appID int32) (bool, error) {
	return helpers.DeleteApp(ctx, m.APIClient, appID)
}

func (m *providerMetadata) DeleteDatabase(ctx context.Context, databaseID int32) (bool, error) {
	return helpers.DeleteDatabase(ctx, m.APIClient, databaseID)
}

func (m *providerMetadata) DeleteEndpoint(ctx context.Context, vhostID int32) (bool, error) {
	return helpers.DeleteEndpoint(ctx, m.APIClient, vhostID)
}

func (m *providerMetadata) DeleteLogDrain(ctx context.Context, logDrainID int32) (bool, error) {
	return helpers.DeleteLogDrain(ctx, m.APIClient, logDrainID)
}

func (m *providerMetadata) DeleteMetricDrain(ctx context.Context, metricDrainID int32) (bool, error) {
	return helpers.DeleteMetricDrain(ctx, m.APIClient, metricDrainID)
}

func (m *providerMetadata) GetStackByName(ctx context.Context, name string) (*aptibleapi.Stack, error) {
	return helpers.GetStackByName(ctx, m.APIClient, name)
}

func (m *providerMetadata) GetDatabaseImageByTypeAndVersion(ctx context.Context, imageType, version string) (*aptibleapi.DatabaseImage, error) {
	return helpers.GetDatabaseImageByTypeAndVersion(ctx, m.APIClient, imageType, version)
}

func (m *providerMetadata) GetServiceForAppByName(ctx context.Context, appID int32, processType string) (*aptibleapi.Service, error) {
	return helpers.GetServiceForAppByName(ctx, m.APIClient, appID, processType)
}

func (m *providerMetadata) GetReplicaByHandle(ctx context.Context, databaseID int32, handle string) (*aptibleapi.Database, error) {
	return helpers.GetReplicaByHandle(ctx, m.APIClient, databaseID, handle)
}

func (m *providerMetadata) GetOrganizationFromAuthAPI() (string, error) {
	return helpers.GetOrganizationFromAuthAPI(m.Token, helpers.GetAuthURL())
}

func (m *providerMetadata) GetStackOrganizationID(stack *aptibleapi.Stack) string {
	return helpers.GetStackOrganizationID(stack)
}
