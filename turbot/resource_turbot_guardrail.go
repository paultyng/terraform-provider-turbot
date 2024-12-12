package turbot

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/terraform-provider-turbot/apiClient"
	"github.com/turbot/terraform-provider-turbot/errors"
	"github.com/turbot/terraform-provider-turbot/helpers"
)

// properties which must be passed to a create/update call
var guardrailProperties = []interface{}{"title", "description", "parent", "filter", "targets", "akas"}

func getGuardrailUpdateProperties() []interface{} {
	excludedProperties := []string{"parent"}
	return helpers.RemoveProperties(guardrailProperties, excludedProperties)
}

func resourceTurbotGuardrail() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotGuardrailCreate,
		Read:   resourceTurbotGuardrailRead,
		Update: resourceTurbotGuardrailUpdate,
		Delete: resourceTurbotGuardrailDelete,
		Exists: resourceTurbotGuardrailExists,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotGuardrailImport,
		},
		Schema: map[string]*schema.Schema{
			//aka of the parent resource
			"parent": {
				Type: schema.TypeString,
				// when doing a diff, the state file will contain the id of the parent but the config contains the aka,
				// so we need custom diff code
				DiffSuppressFunc: suppressIfAkaMatches("parent_akas"),
				Optional:         true,
				Default:          "tmod:@turbot/turbot#/",
			},
			//when doing a read, fetch the parent akas to use in suppressIfAkaMatches
			"parent_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"title": {
				Type:     schema.TypeString,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"filter": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"targets": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"akas": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				DiffSuppressFunc: suppressIfAkaRemoved(),
			},
		},
	}
}

func resourceTurbotGuardrailExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*apiClient.Client)
	id := d.Id()
	return client.ResourceExists(id)
}

func resourceTurbotGuardrailCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	// build map of folder properties
	input := mapFromResourceData(d, guardrailProperties)

	policyPack, err := client.CreatePolicyPack(input)
	if err != nil {
		return err
	}

	// assign the id
	d.SetId(policyPack.Turbot.Id)
	// TODO Remove Read call once schema changes are In.
	return resourceTurbotGuardrailRead(d, meta)
}

func resourceTurbotGuardrailUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()

	// build map of folder properties
	input := mapFromResourceData(d, getGuardrailUpdateProperties())
	input["id"] = id

	_, err := client.UpdatePolicyPack(input)
	if err != nil {
		return err
	}
	// set 'Read' Properties
	// TODO Remove Read call once schema changes are In.
	return resourceTurbotGuardrailRead(d, meta)
}

func resourceTurbotGuardrailRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()

	policyPack, err := client.ReadPolicyPack(id)
	if err != nil {
		if errors.NotFoundError(err) {
			// folder was not found - clear id
			d.SetId("")
		}
		return err
	}

	// assign results back into ResourceData
	// set parent_akas property by loading resource and fetching the akas
	if err := storeAkas(policyPack.Turbot.ParentId, "parent_akas", d, meta); err != nil {
		return err
	}
	// NOTE currently turbot accepts array of filters but only uses the first
	if len(policyPack.Filters) > 0 {
		d.Set("filter", policyPack.Filters[0])
	}
	d.Set("parent", policyPack.Parent)
	d.Set("title", policyPack.Title)
	d.Set("description", policyPack.Description)
	d.Set("akas", policyPack.Turbot.Akas)

	// The "targets" parameter can be provided as either an ID or the URI of the resource.
	// However, the query response always returns IDs, which can cause a mismatch between
	// the input and the state file. To maintain consistency, the input values are directly
	// saved to the state file as-is.
	if value, ok := d.GetOk("targets"); ok {
		targets := make([]string, len(value.([]interface{})))
		for i, target := range value.([]interface{}) {
			targets[i] = target.(string)
		}
		d.Set("targets", targets)
	}

	return nil
}

func resourceTurbotGuardrailDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()
	err := client.DeleteResource(id)
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")

	return nil
}

func resourceTurbotGuardrailImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotGuardrailRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
