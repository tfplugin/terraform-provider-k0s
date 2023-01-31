package k0s

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

type k0sProvider struct {
}

func New() func() provider.Provider {
	return func() provider.Provider {
		return &k0sProvider{}
	}
}

func (p *k0sProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	//TODO implement me
	panic("implement me")
}

func (p *k0sProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	//TODO implement me
	panic("implement me")
}

func (p *k0sProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	//TODO implement me
	panic("implement me")
}

func (p *k0sProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	//TODO implement me
	panic("implement me")
}

func (p *k0sProvider) Resources(ctx context.Context) []func() resource.Resource {
	//TODO implement me
	panic("implement me")
}
