package provider

import (
	"bytes"
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/k0sproject/k0sctl/cmd"
	"os"
	"strings"
)

const (
	clusterResourceName = "cluster"
)

type clusterResource struct {
}

type clusterResourceModel struct {
	Config        string       `tfsdk:"config"`
	SSHPrivateKey string       `tfsdk:"ssh_private_key"`
	Kubeconfig    types.String `tfsdk:"kubeconfig"`
}

func NewClusterResource() resource.Resource {
	return &clusterResource{}
}

func (r *clusterResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_" + clusterResourceName
}

func (r *clusterResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"config": schema.StringAttribute{
				Description: "k0sctl cluster configuration. Make sure to set host private key path to ./id_dsa",
				Required:    true,
			},
			"ssh_private_key": schema.StringAttribute{
				Description: "SSH private key to authenticate with the hosts defined in configuration.",
				Required:    true,
				Sensitive:   true,
			},
			"kubeconfig": schema.StringAttribute{
				Description: "The created k0s cluster admin kubeconfig",
				Computed:    true,
				Sensitive:   true,
			},
		},
	}
}

func (r *clusterResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	r.apply(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	// Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if err := r.prepare(state.Config, state.SSHPrivateKey); err != nil {
		resp.Diagnostics.AddError("Failed to create temporary files", err.Error())
		return
	}
	defer func() {
		if err := r.cleanUp(); err != nil {
			resp.Diagnostics.AddWarning("Failed to clean up temporary files", err.Error())
		}
	}()

	kubeconfig, err := r.getKubeconfig()
	if err != nil {
		// if k0s cluster doesn't exist, clean resource state
		if strings.Contains(err.Error(), "failed to read file /var/lib/k0s/kubelet.conf") {
			resp.State.RemoveResource(ctx)
		} else {
			resp.Diagnostics.AddError("Failed to read kubeconfig", err.Error())
		}
		return
	}

	state.Kubeconfig = types.StringValue(kubeconfig)
	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

func (r *clusterResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan clusterResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	r.apply(&plan, &resp.Diagnostics)
	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

func (r *clusterResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state clusterResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if diags.HasError() {
		return
	}

	if err := r.prepare(state.Config, state.SSHPrivateKey); err != nil {
		resp.Diagnostics.AddError("Failed to create temporary files", err.Error())
		return
	}
	defer func() {
		if err := r.cleanUp(); err != nil {
			resp.Diagnostics.AddWarning("Failed to clean up temporary files", err.Error())
		}
	}()

	// Run k0sctl reset
	if err := cmd.App.Run([]string{"k0sctl", "reset", "--force"}); err != nil {
		resp.Diagnostics.AddError("Failed to reset k0s cluster.", err.Error())
	}
}

func (r *clusterResource) prepare(config, sshPrivateKey string) error {
	// Save config file
	if err := os.WriteFile("k0sctl.yaml", []byte(config), 0644); err != nil {
		return fmt.Errorf("failed to write config to k0sctl.yaml: %v", err)
	}

	// Save SSH private key
	if err := os.WriteFile("id_dsa", []byte(sshPrivateKey), 0644); err != nil {
		return fmt.Errorf("failed to write config to id_dsa: %v", err)
	}

	return nil
}

func (r *clusterResource) cleanUp() error {
	if err := os.Remove("k0sctl.yaml"); err != nil {
		return fmt.Errorf("failed to remove k0sctl.yaml: %v", err)
	}
	if err := os.Remove("id_dsa"); err != nil {
		return fmt.Errorf("failed to remove id_dsa: %v", err)
	}
	return nil
}

func (r *clusterResource) apply(model *clusterResourceModel, diagnostics *diag.Diagnostics) *clusterResourceModel {
	if err := r.prepare(model.Config, model.SSHPrivateKey); err != nil {
		diagnostics.AddError("Failed to create temporary files", err.Error())
		return nil
	}
	defer func() {
		// Remove local k0sctl.yaml and SSH private key
		if err := r.cleanUp(); err != nil {
			diagnostics.AddWarning("Failed to clean up temporary files", err.Error())
		}
	}()

	// Run kubectl apply
	if err := cmd.App.Run([]string{"k0sctl", "apply"}); err != nil {
		diagnostics.AddError("failed to run k0sctl apply --kubeconfig-out kubeconfig", err.Error())
		return nil
	}

	// Get kubeconfig
	kubeconfig, err := r.getKubeconfig()
	if err != nil {
		diagnostics.AddError("Failed to get kubeconfig", err.Error())
		return nil
	}
	model.Kubeconfig = types.StringValue(kubeconfig)

	return model
}

func (r *clusterResource) getKubeconfig() (string, error) {
	// Redirect app writer to buffer
	buf := new(bytes.Buffer)
	writer := cmd.App.Writer
	cmd.App.Writer = buf
	defer func() {
		cmd.App.Writer = writer
	}()

	// Run k0sctl kubeconfig
	if err := cmd.App.Run([]string{"k0sctl", "kubeconfig"}); err != nil {
		return "", err
	}
	return buf.String(), nil
}
