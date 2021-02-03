package turbot

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/terraform-provider-turbot/apiClient"
	"github.com/turbot/terraform-provider-turbot/errors"
)

// map of Terraform properties to Turbot properties that we pass to create and update mutations
// NOTE: use a map instead of an array like other resources as we cannot automatically map the names
var grantInputProperties = []interface{}{"identity", "type", "level", "resource"}

func resourceTurbotGrant() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotGrantCreate,
		Read:   resourceTurbotGrantRead,
		Delete: resourceTurbotGrantDelete,
		Exists: resourceTurbotGrantExists,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotGrantImport,
		},
		Schema: map[string]*schema.Schema{
			// aka of the resource resource
			"resource": {
				Type:     schema.TypeString,
				Required: true,
				// when doing a diff, the state file will contain the id of the resource but the config contains the aka,
				// so we need custom diff code
				DiffSuppressFunc: suppressIfAkaMatches("resource_akas"),
				ForceNew:         true,
			},
			"type": {
				Type:     schema.TypeString,
				Required: true,
				// when doing a diff, the state file will contain the id of the permission type but the config contains the aka,
				// so we need custom diff code
				DiffSuppressFunc: suppressIfAkaMatches("permission_type_akas"),
				ForceNew:         true,
			},
			"level": {
				Type:     schema.TypeString,
				Required: true,
				// when doing a diff, the state file will contain the id of the permission level but the config contains the aka,
				// so we need custom diff code
				DiffSuppressFunc: suppressIfAkaMatches("permission_level_akas"),
				ForceNew:         true,
			},
			"identity": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
				// when doing a diff, the state file will contain the id of the identity but the config contains the aka,
				// so we need custom diff code
				DiffSuppressFunc: suppressIfAkaMatches("identity_akas"),
			},
			"resource_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"permission_type_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"permission_level_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"identity_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
	}
}

func resourceTurbotGrantExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*apiClient.Client)
	id := d.Id()
	return client.GrantExists(id)
}

func resourceTurbotGrantCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	resourceAka := d.Get("resource").(string)
	identityAka := d.Get("identity").(string)
	permissionTypeAka := d.Get("type").(string)
	permissionLevelAka := d.Get("level").(string)
	// build map of Grant properties
	input := mapFromResourceData(d, grantInputProperties)
	// create Grant returns turbot resource metadata containing the id
	TurbotGrantMetadata, err := client.CreateGrant(input)
	if err != nil {
		return err
	}

	// set akas properties by loading resource and fetching the akas
	if err := storeAkas(resourceAka, "resource_akas", d, meta); err != nil {
		return err
	}
	if err := storeAkas(identityAka, "identity_akas", d, meta); err != nil {
		return err
	}
	if err := storeAkas(permissionTypeAka, "permission_type_akas", d, meta); err != nil {
		return err
	}
	if err := storeAkas(permissionLevelAka, "permission_level_akas", d, meta); err != nil {
		return err
	}

	// assign the id
	d.SetId(TurbotGrantMetadata.Id)
	return nil
}

func resourceTurbotGrantRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()

	Grant, err := client.ReadGrant(id)
	if err != nil {
		if errors.NotFoundError(err) {
			// Grant was not found - clear id
			d.SetId("")
		}
		return err
	}

	// assign results back into ResourceData
	d.Set("level", Grant.PermissionLevelId)
	d.Set("type", Grant.PermissionTypeId)
	d.Set("identity", Grant.Turbot.ProfileId)
	d.Set("resource", Grant.Turbot.ResourceId)

	// set akas properties by loading resource and fetching the akas
	if err := storeAkas(Grant.Turbot.ResourceId, "resource_akas", d, meta); err != nil {
		return err
	}
	if err := storeAkas(Grant.Turbot.ProfileId, "identity_akas", d, meta); err != nil {
		return err
	}
	if err := storeAkas(Grant.PermissionTypeId, "permission_type_akas", d, meta); err != nil {
		return err
	}
	return storeAkas(Grant.PermissionLevelId, "permission_level_akas", d, meta)
}

func resourceTurbotGrantDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()
	err := client.DeleteGrant(id)
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")

	return nil
}

func resourceTurbotGrantImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotGrantRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
