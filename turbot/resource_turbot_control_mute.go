package turbot

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/terraform-provider-turbot/apiClient"
	"github.com/turbot/terraform-provider-turbot/errors"
)

var controlProperties = map[string]string{
	"control_id":   "id",
	"resource":     "resource",
	"control_type": "controlType",
	"note":         "note",
	"to_timestamp": "toTimestamp",
	"until_states": "untilStates",
}

func getControlUpdateProperties() map[string]string {
	excludedProperties := []string{"control_id", "resource", "control_type"}

	// Remove the excluded properties from controlProperties
	for _, key := range excludedProperties {
		delete(controlProperties, key)
	}

	return controlProperties
}

func getControlDeleteProperties() map[string]string {
	excludedProperties := []string{"to_timestamp", "note", "until_states", "resource", "control_type"}

	// Remove the excluded properties from controlProperties
	for _, key := range excludedProperties {
		delete(controlProperties, key)
	}

	return controlProperties
}

func resourceTurbotControlMute() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotControlMuteCreate,
		Read:   resourceTurbotControlMuteRead,
		Update: resourceTurbotControlMuteUpdate,
		Delete: resourceTurbotControlMuteDelete,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotControlMuteImport,
		},
		Schema: map[string]*schema.Schema{
			"control_id": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"resource": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"control_type": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"note": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"to_timestamp": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"until_states": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"mute_state": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceTurbotControlMuteCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)

	// build input map to pass to mutation
	if value, ok := d.GetOk("control_id"); ok {
		d.Set("control_id", value.(string))
		delete(controlProperties, "resource")
		delete(controlProperties, "control_type")
	}

	if value, ok := d.GetOk("resource"); ok {
		d.Set("resource", value.(string))
		delete(controlProperties, "id")
	}

	if value, ok := d.GetOk("control_type"); ok {
		d.Set("control_type", value.(string))
		delete(controlProperties, "id")
	}

	// build mutation input data by parsing the resource schema and
	// excluding top level properties which must not be sent to update - `type`
	input := mapFromResourceDataWithPropertyMap(d, controlProperties)

	muteControl, err := client.MuteControl(input)
	if err != nil {
		return err
	}

	// Set the id
	d.SetId(muteControl.Turbot["id"])

	d.Set("control_id", muteControl.Turbot["id"])
	d.Set("resource", muteControl.Turbot["resourceId"])
	d.Set("mute_state", muteControl.Turbot["muteState"])
	d.Set("to_timestamp", muteControl.Turbot["muteToTimestamp"])
	d.Set("control_type", muteControl.Type.Uri)

	// Check control mute status
	muteConfig := muteControl.Mute.(map[string]interface{})
	if muteConfig["note"] != nil {
		d.Set("note", muteConfig["note"].(string))
	}

	if muteConfig["untilStates"] != nil {
		d.Set("until_states", muteConfig["untilStates"].([]interface{}))
	}

	return nil
}

func resourceTurbotControlMuteRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)

	// Get the control id
	controlId := d.Get("control_id").(string)

	control, err := client.ReadControl(controlId)
	if err != nil {
		if errors.NotFoundError(err) {
			// folder was not found - clear id
			d.Set("id", "")
		}
		return err
	}

	d.SetId(control.Turbot["id"])
	d.Set("control_id", control.Turbot["id"])
	d.Set("control_type", control.Type.Uri)
	d.Set("resource", control.Turbot["resourceId"])

	muteConfig := control.Mute.(map[string]interface{})
	if muteConfig["state"] != nil {
		d.Set("mute_state", muteConfig["state"].(string))
	}
	if muteConfig["toTimestamp"] != nil {
		d.Set("to_timestamp", muteConfig["toTimestamp"].(string))
	}
	if muteConfig["untilStates"] != nil {
		d.Set("until_states", muteConfig["untilStates"].([]interface{}))
	}
	if muteConfig["note"] != nil {
		d.Set("note", muteConfig["note"].(string))
	}

	return nil
}

func resourceTurbotControlMuteUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	controlId := d.Get("control_id").(string)

	// build map of folder properties
	input := mapFromResourceDataWithPropertyMap(d, getControlUpdateProperties())
	input["id"] = controlId

	muteControl, err := client.MuteControl(input)
	if err != nil {
		return err
	}
	// Set the id
	d.SetId(controlId)

	d.Set("control_id", muteControl.Turbot["id"])
	d.Set("resource", muteControl.Turbot["resourceId"])
	d.Set("mute_state", muteControl.Turbot["muteState"])
	d.Set("to_timestamp", muteControl.Turbot["muteToTimestamp"])
	d.Set("control_type", muteControl.Type.Uri)

	// Check control mute status
	muteConfig := muteControl.Mute.(map[string]interface{})
	if muteConfig["note"] != nil {
		d.Set("note", muteConfig["note"].(string))
	}

	if muteConfig["untilStates"] != nil {
		d.Set("until_states", muteConfig["untilStates"].([]interface{}))
	}

	return nil
}

func resourceTurbotControlMuteDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	controlId := d.Get("control_id").(string)

	// build map of folder properties
	input := mapFromResourceDataWithPropertyMap(d, getControlDeleteProperties())
	input["id"] = controlId

	_, err := client.UnMuteControl(input)
	if err != nil {
		return err
	}

	return nil
}

func resourceTurbotControlMuteImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotControlMuteRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}
