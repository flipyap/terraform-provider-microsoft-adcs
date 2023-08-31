package provider

import (
	"context"
	"fmt"

	"github.com/flipyap/microsoft-adcs-client/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ datasource.DataSource              = &certificateDataSource{}
	_ datasource.DataSourceWithConfigure = &certificateDataSource{}
)

// NewCertificateDataSource is a helper function to simplify the provider implementation.
func NewCertificateDataSource() datasource.DataSource {
	return &certificateDataSource{}
}

// certificateDataSource is the data source implementation.
type certificateDataSource struct {
	client *client.ADCSClient
}

// coffeesModel maps coffees schema data.
type certificateModel struct {
	ID                  types.String `tfsdk:"id"`
	CertificateB64      types.String `tfsdk:"certificate_b64"`
	CertificateChainB64 types.String `tfsdk:"certificate_chain_b64"`
}

// Configure adds the provider configured client to the data source.
func (d *certificateDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.ADCSClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.ADCSClient, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.client = client
}

// Metadata returns the data source type name.
func (d *certificateDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

// Schema defines the schema for the data source.
func (d *certificateDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the certificate that was generated.",
				Required:    true,
			},
			"certificate_b64": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate returned from ADCS as base64 encoded.",
			},
			"certificate_chain_b64": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate chain returned from ADCS as base64 encoded.",
			},
		},
	}
}

// Read refreshes the Terraform state with the latest data.
func (d *certificateDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Retrieve values from plan
	var data certificateModel
	// Read Terraform configuration data into the model
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	reqID := data.ID.ValueString()

	certificates, err := d.client.RetrieveCertificates(reqID)
	if err != nil {
		// diagError = "Unable to Read certificates for " + reqID
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to Read certificates for %d", &reqID),
			err.Error(),
		)
		return
	}

	// Map response body to model
	state := certificateModel{
		ID:                  types.StringValue(certificates.ID),
		CertificateB64:      types.StringValue(certificates.CertificateB64),
		CertificateChainB64: types.StringValue(certificates.CertificateChainB64),
	}

	// Set state
	diags := resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

}
