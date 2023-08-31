package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/fatih/structs"
	"github.com/flipyap/microsoft-adcs-client/client"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ resource.Resource                = &certificateResource{}
	_ resource.ResourceWithConfigure   = &certificateResource{}
	_ resource.ResourceWithImportState = &certificateResource{}
)

// NewCertificateResource is a helper function to simplify the provider implementation.
func NewCertificateResource() resource.Resource {
	return &certificateResource{}
}

// certificateResource is the resource implementation.
type certificateResource struct {
	client *client.ADCSClient
}

type certificateCreateModel struct {
	ID                  types.String `tfsdk:"id"`
	Attributes          types.String `tfsdk:"attributes"`
	CSR                 types.String `tfsdk:"certificate_signing_request"`
	Template            types.String `tfsdk:"template"`
	CertificateB64      types.String `tfsdk:"certificate_b64"`
	CertificateChainB64 types.String `tfsdk:"certificate_chain_b64"`
	LastUpdated         types.String `tfsdk:"last_updated"`
}

// Metadata returns the resource type name.
func (r *certificateResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_certificate"
}

// Schema defines the schema for the resource.
func (r *certificateResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Description: "Numeric identifier of the generated certificate.",
				Computed:    true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"certificate_signing_request": schema.StringAttribute{
				Required:    true,
				Description: "The certificate signing request used to create a certificate ",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
				// TODO: make a validator that can validate base64: https://developer.hashicorp.com/terraform/plugin/framework/handling-data/types/custom
			},
			"template": schema.StringAttribute{
				Required: true,
				Description: `There are usually several predefined templates that make it easier to request certificates 
depending on what they are needed for.`,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"attributes": schema.StringAttribute{
				Optional:    true,
				Description: "Extra attributes to add to the certificate",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"certificate_b64": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate returned from ADCS as base64 encoded.",
			},
			"certificate_chain_b64": schema.StringAttribute{
				Computed:    true,
				Description: "The certificate chain returned from ADCS as base64 encoded.",
			},
			"last_updated": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

// Configure adds the provider configured client to the resource.
func (r *certificateResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

	r.client = client
}

// Create creates the resource and sets the initial Terraform state.
func (r *certificateResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// Retrieve values from plan
	var plan certificateCreateModel
	attr := ""
	diags := req.Plan.Get(ctx, &plan)

	resp.Diagnostics.Append(diags...)
	// Add attributes if provided
	if !plan.Attributes.IsNull() && !plan.Attributes.IsUnknown() {
		tflog.Debug(ctx, "Adding attributes to certificate creation", map[string]interface{}{
			"attributes": plan.Attributes,
		})
		attr = plan.Attributes.ValueString()
	}
	// Create new certificate
	tflog.Info(ctx, "Requesting certificate from ADCS server.")
	tflog.Debug(ctx, "Certificate request Data", structs.Map(plan))
	certificates, err := r.client.RequestCertificate(plan.CSR.ValueString(), client.TemplateName(plan.Template.ValueString()), attr)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating certificate from singing request",
			"Could not create certificate, unexpected error: "+err.Error(),
		)
		return
	}

	plan.ID = types.StringValue(certificates.ID)
	plan.CertificateB64 = types.StringValue(certificates.CertificateB64)
	plan.CertificateChainB64 = types.StringValue(certificates.CertificateChainB64)
	plan.LastUpdated = types.StringValue(time.Now().Format(time.RFC850))

	// Set state to fully populated data
	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Read refreshes the Terraform state with the latest data.
func (r *certificateResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// Get current state
	var state certificateCreateModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
	reqID := state.ID.ValueString()

	// Get refreshed order value from HashiCups
	certificates, err := r.client.RetrieveCertificates(reqID)

	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Certificate",
			fmt.Sprintf("Could not read Certificate ID %s", state.ID.ValueString())+":"+err.Error(),
		)
		return
	}

	// Overwrite items with refreshed state
	state.ID = types.StringValue(certificates.ID)
	state.CertificateB64 = types.StringValue(strings.Replace(certificates.CertificateB64, `\r`, "", -1))
	state.CertificateChainB64 = types.StringValue(strings.Replace(certificates.CertificateChainB64, `\r`, "", -1))

	// Set refreshed state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// Update updates the resource and sets the updated Terraform state on success.
func (r *certificateResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//There is no updates that can be performed on the adcs side
}

// Delete deletes the resource and removes the Terraform state on success.
func (r *certificateResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	// If we delete there is nothing to be done on the ADCS side.. State management is handled by terraform so we don't
	// need to do anything here
}

func (r *certificateResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	// Retrieve import ID and save to id attribute
	resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
