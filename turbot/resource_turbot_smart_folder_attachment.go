package turbot

import (
	"fmt"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/terraform-provider-turbot/apiClient"
	"strings"
)

var smartFolderAttachProperties = map[string]string{
	"resource":     "resource",
	"smart_folder": "smartFolders",
}

func resourceTurbotSmartFolderAttachemnt() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotSmartFolderAttachmentCreate,
		Read:   resourceTurbotSmartFolderAttachmentRead,
		Delete: resourceTurbotSmartFolderAttachmentDelete,
		Exists: resourceTurbotSmartFolderAttachmentExists,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotSmartFolderAttachmentImport,
		},
		Schema: map[string]*schema.Schema{
			"resource": {
				Type:             schema.TypeString,
				Required:         true,
				ForceNew:         true,
				DiffSuppressFunc: suppressIfAkaMatches("resource_akas"),
			},
			"smart_folder": {
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
		},
	}
}

func resourceTurbotSmartFolderAttachmentExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	client := meta.(*apiClient.Client)
	smartFolderId, resource := parseSmartFolderId(d.Id())
	// execute api call
	smartFolder, err := client.ReadSmartFolder(smartFolderId)
	if err != nil {
		return false, fmt.Errorf("error reading smart folder: %s", err.Error())
	}

	//find resource aka in list of attached resources
	for _, attachedResource := range smartFolder.AttachedResources.Items {
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

func resourceTurbotSmartFolderAttachmentCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	resource := d.Get("resource").(string)
	smartFolder := d.Get("smart_folder").(string)
	input := mapFromResourceDataWithPropertyMap(d, smartFolderAttachProperties)

	_, err := client.CreateSmartFolderAttachment(input)
	if err != nil {
		return err
	}

	// set resource_akas property by loading resource and fetching the akas
	if err := storeAkas(resource, "resource_akas", d, meta); err != nil {
		return err
	}
	// assign the id
	var stateId = buildId(smartFolder, resource)
	d.SetId(stateId)
	d.Set("resource", resource)
	d.Set("smart_folder", smartFolder)
	return nil
}

func resourceTurbotSmartFolderAttachmentRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	// NOTE: This will not be called if the attachment does not exist
	smartFolder, resource := parseSmartFolderId(d.Id())

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
	d.Set("smart_folder", smartFolder)
	return nil
}

func resourceTurbotSmartFolderAttachmentDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	input := mapFromResourceDataWithPropertyMap(d, smartFolderAttachProperties)
	err := client.DeleteSmartFolderAttachment(input)
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")
	return nil
}

func resourceTurbotSmartFolderAttachmentImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotSmartFolderAttachmentRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func buildId(smartFolder, resource string) string {
	return smartFolder + "_" + resource
}

func parseSmartFolderId(id string) (smartFolder, resource string) {
	segments := strings.Split(id, "_")
	smartFolder = segments[0]
	resource = segments[1]
	return
}
