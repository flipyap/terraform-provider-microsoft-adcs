// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/flipyap/microsoft-adcs-client/client"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure MicrosoftADCSProvider satisfies various provider interfaces.
var _ provider.Provider = &MicrosoftADCSProvider{}

// MicrosoftADCSProvider defines the provider implementation.
type MicrosoftADCSProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// MicrosoftADCSProviderModel describes the provider data model.
type MicrosoftADCSProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
	Krb5Conf types.String `tfsdk:"krb5conf"`
	Ntlm     types.Bool   `tfsdk:"use_ntlm"`
}

func (p *MicrosoftADCSProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "microsoftadcs"
	resp.Version = p.version
}

func (p *MicrosoftADCSProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				MarkdownDescription: "Hostname of the Server hosting the Active Directory Certificate Services",
				Optional:            true,
			},
			"username": schema.StringAttribute{
				MarkdownDescription: "Active Directory Username for Kerberos authentication",
				Optional:            true,
			},
			"password": schema.StringAttribute{
				MarkdownDescription: "Active Directory Password for Kerberos authentication",
				Optional:            true,
				Sensitive:           true,
			},
			"krb5conf": schema.StringAttribute{
				MarkdownDescription: "Kerberos Config to use for authentication",
				Optional:            true,
			},
			"use_ntlm": schema.BoolAttribute{
				MarkdownDescription: "Use NTLM authentication",
				Optional:            true,
			},
		},
	}
}

func (p *MicrosoftADCSProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	tflog.Info(ctx, "Configuring Active Directory Certificate Services client")
	// Retrieve provider data from configuration
	var config MicrosoftADCSProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown Active Directory Certificate Services Host",
			"The provider cannot create the ADCS API client as there is an unknown configuration value for the ADCS host. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADCS_HOST environment variable.",
		)
	}

	if config.Username.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown Active Directory Certificate Services Username",
			"The provider cannot create the ADCS API client as there is an unknown configuration value for the ADCS username. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADCS_USERNAME environment variable.",
		)
	}

	if config.Password.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown Active Directory Certificate Services Password",
			"The provider cannot create the ADCS API client as there is an unknown configuration value for the ADCS password. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the ADCS_PASSWORD environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	host := os.Getenv("ADCS_HOST")
	username := os.Getenv("ADCS_USERNAME")
	password := os.Getenv("ADCS_PASSWORD")
	krb5conf := os.Getenv("ADCS_KRB5CONF")
	useNtlm := config.Ntlm.ValueBool()

	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}

	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}

	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if !config.Krb5Conf.IsNull() {
		krb5conf = config.Krb5Conf.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing Active Directory Certificate Services Host",
			"The provider cannot create the ADCS API client as there is a missing or empty value for the ADCS API host. "+
				"Set the host value in the configuration or use the ADCS_HOST environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if username == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Missing Active Directory Certificate Services Username",
			"The provider cannot create the ADCS API client as there is a missing or empty value for the ADCS username. "+
				"Set the username value in the configuration or use the ADCS_USERNAME environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Missing Active Directory Certificate Services Password",
			"The provider cannot create the ADCS API client as there is a missing or empty value for the ADCS password. "+
				"Set the password value in the configuration or use the ADCS_PASSWORD environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}
	ctx = tflog.SetField(ctx, "adcs_host", host)
	ctx = tflog.SetField(ctx, "adcs_username", username)
	ctx = tflog.SetField(ctx, "adcs_password", password)
	ctx = tflog.MaskFieldValuesWithFieldKeys(ctx, "adcs_password")

	tflog.Debug(ctx, "Creating Active Directory Certificate Services client")

	// Create a new ADCS client using the configuration values.
	clientConfig := client.ClientConfig{
		Host:     host,
		Username: username,
		Password: password,
		Krb5Conf: krb5conf,
		Ntlm:     useNtlm,
	}
	client, err := client.NewClient(&clientConfig)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create Active Directory Certificate Services API Client",
			"An unexpected error occurred when creating the Active Directory Certificate Services API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"ADCS Client Error: "+err.Error(),
		)
		return
	}

	// Make the adcs client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client

	tflog.Info(ctx, "Configured Active Directory Certificate Services client", map[string]any{"success": true})
}

func (p *MicrosoftADCSProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCertificateResource,
	}
}

func (p *MicrosoftADCSProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCertificateDataSource,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &MicrosoftADCSProvider{
			version: version,
		}
	}
}
