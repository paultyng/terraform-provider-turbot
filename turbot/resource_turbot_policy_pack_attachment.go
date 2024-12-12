package turbot

import (
	"fmt"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/terraform-provider-turbot/apiClient"
)

var policyPackAttachProperties = map[string]string{
	"resource": "resource",
}

func resourceTurbotPolicyPackAttachment() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotPolicyPackAttachmentCreate,
		Read:   resourceTurbotPolicyPackAttachmentRead,
		Delete: resourceTurbotPolicyPackAttachmentDelete,
		Exists: resourceTurbotPolicyPackAttachmentExists,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotPolicyPackAttachmentImport,
		},
		Schema: map[string]*schema.Schema{
			"resource": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressIfAkaMatches("resource_akas"),
			},
			"policy_pack": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"phase": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
		},
	}
}

func resourceTurbotPolicyPackAttachmentExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*apiClient.Client)
	policyPackId, resource := parsePolicyPackId(d.Id())
	// execute api call
	policyPack, err := client.ReadPolicyPack(policyPackId)
	if err != nil {
		return false, fmt.Errorf("error reading policy pack: %s", err.Error())
	}

	// find resource aka in list of attached resources
	for _, attachedResource := range policyPack.AttachedResources.Items {
		if resource == attachedResource.Turbot.Id {
			return true, nil
		}
		for _, aka := range attachedResource.Turbot.Akas {
			if aka == resource {
				return true, nil
			}
		}
	}
	return false, nil
}

func resourceTurbotPolicyPackAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	resource := d.Get("resource").(string)
	policyPack := d.Get("policy_pack").(string)
	input := mapFromResourceDataWithPropertyMap(d, policyPackAttachProperties)

	policyPackByPhase := []apiClient.PolicyPackByPhase{
		{
			Id: policyPack,
		},
	}

	if phase, ok := d.GetOk("phase"); ok {
		policyPackByPhase[0].Phase = phase.(string)
	}
	input["policyPacksPhase"] = policyPackByPhase

	_, err := client.CreatePolicyPackAttachment(input)
	if err != nil {
		return err
	}

	// set resource_akas property by loading resource and fetching the akas
	if err := storeAkas(resource, "resource_akas", d, meta); err != nil {
		return err
	}
	// assign the id
	var stateId = buildPolicyPackId(policyPack, resource)
	d.SetId(stateId)
	d.Set("resource", resource)
	d.Set("policy_pack", policyPack)
	d.Set("phase", d.Get("phase").(string))
	return nil
}

func resourceTurbotPolicyPackAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	// NOTE: This will not be called if the attachment does not exist
	policyPack, resource := parsePolicyPackId(d.Id())

	turbotResource, err := client.ReadResource(resource, nil)
	if err != nil {
		return err
	}
	// set resource_akas property by loading resource and fetching the akas
	if err := storeAkas(turbotResource.Turbot.Id, "resource_akas", d, meta); err != nil {
		return err
	}
	// assign results directly back into ResourceData
	d.Set("resource", resource)
	d.Set("policy_pack", policyPack)
	return nil
}

func resourceTurbotPolicyPackAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	input := mapFromResourceDataWithPropertyMap(d, policyPackAttachProperties)
	err := client.DeletePolicyPackAttachment(input)
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")
	return nil
}

func resourceTurbotPolicyPackAttachmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotPolicyPackAttachmentRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func buildPolicyPackId(policyPack, resource string) string {
	return policyPack + "_" + resource
}

func parsePolicyPackId(id string) (policyPack, resource string) {
	// Get the index of the first underscore
	index := strings.Index(id, "_")
	policyPack = id[:index]
	resource = id[index+1:]
	return
}
