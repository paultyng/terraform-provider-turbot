package turbot

import (
	"fmt"

	"github.com/go-yaml/yaml"
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/turbot/terraform-provider-turbot/apiClient"
	"github.com/turbot/terraform-provider-turbot/errors"
	"github.com/turbot/terraform-provider-turbot/helpers"
)

var policySettingInputProperties = []interface{}{"value", "precedence", "template", "template_input", "note", "valid_from_timestamp", "valid_to_timestamp", "type", "resource"}

func getPolicySettingUpdateProperties() []interface{} {
	excludedProperties := []string{"type", "resource"}
	return helpers.RemoveProperties(policySettingInputProperties, excludedProperties)
}
func resourceTurbotPolicySetting() *schema.Resource {
	return &schema.Resource{
		Create: resourceTurbotPolicySettingCreate,
		Read:   resourceTurbotPolicySettingRead,
		Update: resourceTurbotPolicySettingUpdate,
		Delete: resourceTurbotPolicySettingDelete,
		Exists: resourceTurbotPolicySettingExists,
		Importer: &schema.ResourceImporter{
			State: resourceTurbotPolicySettingImport,
		},
		Schema: map[string]*schema.Schema{
			"type": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"resource": {
				Type:             schema.TypeString,
				Required:         true,
				DiffSuppressFunc: suppressIfAkaMatches("resource_akas"),
				ForceNew:         true,
			},
			"resource_akas": {
				Type:     schema.TypeList,
				Computed: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"value": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressIfEncryptedOrValueSourceMatches,
			},
			"value_source": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"value_source_key_fingerprint": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"precedence": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "REQUIRED",
			},
			"template": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"template_input": {
				Type:             schema.TypeString,
				Optional:         true,
				DiffSuppressFunc: suppressIfTemplateInputEquivalent,
			},
			"note": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_from_timestamp": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"valid_to_timestamp": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"value_source_used": {
				Type:     schema.TypeBool,
				Computed: true,
			},
			"pgp_key": {
				Type:     schema.TypeString,
				ForceNew: true,
				Optional: true,
			},
			"enforce": {
				Type:     schema.TypeList,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"value": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"template": {
							Type:     schema.TypeString,
							Optional: true,
						},
						"template_input": {
							Type:     schema.TypeString,
							Optional: true,
						},
					},
				},
			},
		},
	}
}

func resourceTurbotPolicySettingExists(d *schema.ResourceData, meta interface{}) (b bool, e error) {
	// Exists - This is called to verify a resource still exists. It is called prior to Read,
	// and lowers the burden of Read to be able to assume the resource exists.
	client := meta.(*apiClient.Client)
	id := d.Id()

	_, err := client.ReadPolicySetting(id)
	if err != nil {
		if errors.NotFoundError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func resourceTurbotPolicySettingCreate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	policyTypeUri := d.Get("type").(string)
	resourceAka := d.Get("resource").(string)

	// first check if the folder exists - search by parent and foldere title
	existingSetting, err := client.FindPolicySetting(policyTypeUri, resourceAka)
	if err != nil {
		return err
	}
	if existingSetting.Value != nil {
		return fmt.Errorf("A policy setting for policy type: '%s', resource: '%s' already exists ( id: %s ). To manage the existing setting using Terraform, import it using command 'terraform import <resource_address> <id>'",
			policyTypeUri, resourceAka, existingSetting.Turbot.Id)
	}

	// NOTE:  turbot policy settings have a value and a valueSource property
	// - value is the type property value, with the type dependent on the policy schema
	// - valueSource is the yaml representation of the policy.
	//	as we are not sure of the value format provided, we try multiple times, relying on the policy validation to
	//	reject invalid values
	// 1) pass value as 'value'
	// 2) pass value as 'valueSource'. update d.value to be the yaml parsed version of 'value'
	input := mapFromResourceData(d, policySettingInputProperties)

	var customInputPhase map[string]string
	var enforceValString, enforceTemplate, enforceTemplateInput string
	enforceValueList := d.Get("enforce").([]interface{})

	if len(enforceValueList) > 0 {
		enforceValue := enforceValueList[0].(map[string]interface{})

		if len(enforceValue) > 0 {
			if enforceValue["template"] != nil {
				customInputPhase = map[string]string{
					"template":       enforceValue["template"].(string),
					"template_input": enforceValue["template_input"].(string),
				}
				enforceTemplate = enforceValue["template"].(string)
				enforceTemplateInput = enforceValue["template_input"].(string)
			}

			if enforceValue["value"] != nil {
				customInputPhase = map[string]string{
					"value": enforceValue["value"].(string),
				}

				enforceValString = enforceValue["value"].(string)
			}
			input["enforce"] = customInputPhase

		}
	}

	if value, ok := d.GetOk("template_input"); ok {
		// NOTE: ParseYamlString doesn't validate input as valid YAML format, on error it returns value
		valueString := fmt.Sprintf("%v", value)
		input["templateInput"], err = helpers.ParseYamlString(valueString)
	}

	policySetting, err := client.CreatePolicySetting(input)
	if err != nil {
		if !errors.FailedValidationError(err) {
			d.SetId("")
			return err
		}
		// so we have a data validation error, try the value source
		input["valueSource"] = input["value"]
		delete(input, "value")
		// try again
		policySetting, err = client.CreatePolicySetting(input)
		if err != nil {
			d.SetId("")
			return err
		}
		// update state value setting with yaml parsed valueSource
		setValueFromValueSource(input["valueSource"].(string), d)
	}
	// if pgp_key has been supplied, encrypt value and value_source
	storeValue(d, policySetting)
	// set akas properties by loading resource and fetching the akas
	if err := storeAkas(resourceAka, "resource_akas", d, meta); err != nil {
		return err
	}

	// NOTE: TemplateInput can be string or array of strings
	// - In case of string, we return string
	// - In array of strings, we return a valid YAML string
	templateInput, err := helpers.InterfaceToStringOrYaml(policySetting.TemplateInput)
	if err != nil {
		return err
	}
	// assign read properties
	d.Set("precedence", policySetting.Precedence)
	d.Set("template", policySetting.Template)
	d.Set("template_input", templateInput)
	d.Set("note", policySetting.Note)
	d.Set("valid_from_timestamp", policySetting.ValidFromTimestamp)
	d.Set("valid_to_timestamp", policySetting.ValidToTimestamp)
	d.Set("type", policySetting.Type.Uri)

	d.Set("phase", map[string]map[string]string{
		"enforce": {
			"value":          enforceValString,
			"template":       enforceTemplate,
			"template_input": enforceTemplateInput,
		},
	})

	// assign the id
	d.SetId(policySetting.Turbot.Id)

	return nil
}

func resourceTurbotPolicySettingRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()

	policySetting, err := client.ReadPolicySetting(id)
	if err != nil {
		if errors.NotFoundError(err) {
			// setting was not found - clear id
			d.SetId("")
		}
		return err
	}

	// set akas properties by loading resource and fetching the akas
	if err := storeAkas(policySetting.Turbot.ResourceId, "resource_akas", d, meta); err != nil {
		return err
	}

	// NOTE: TemplateInput can be string or array of strings
	// - In case of string, we return string
	// - In array of strings, we return a valid YAML string
	templateInput, err := helpers.InterfaceToStringOrYaml(policySetting.TemplateInput)

	if err != nil {
		return err
	}
	// assign results back into ResourceData
	// if pgp_key has been supplied, encrypt value and value_source
	storeValue(d, policySetting)
	d.Set("precedence", policySetting.Precedence)
	d.Set("resource", policySetting.Turbot.ResourceId)
	d.Set("template", policySetting.Template)
	d.Set("template_input", templateInput)
	d.Set("note", policySetting.Note)
	d.Set("valid_from_timestamp", policySetting.ValidFromTimestamp)
	d.Set("valid_to_timestamp", policySetting.ValidToTimestamp)
	d.Set("type", policySetting.Type.Uri)

	if value, ok := d.GetOk("phase"); ok {
		// If ok is true, 'example_attribute' is set in the configuration
		d.Set("phase", value)
	}

	return nil
}

func resourceTurbotPolicySettingUpdate(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()

	// NOTE:  turbot policy settings have a value and a valueSource property
	// - value is the type property value, with the type dependent on the policy schema
	// - valueSource is the yaml representation of the policy.
	//	as we are not sure of the value format provided, we try multiple times, relying on the policy validation to
	//	reject invalid values
	// 1) pass value as 'value'
	// 2) pass value as 'valueSource'. update d.value to be the yaml parsed version of 'value'
	input := mapFromResourceData(d, getPolicySettingUpdateProperties())
	input["id"] = id

	var err error
	if value, ok := d.GetOk("template_input"); ok {
		// NOTE: ParseYamlString doesn't validate input as valid YAML format, on error it just returns the value
		valueString := fmt.Sprintf("%v", value)
		input["templateInput"], err = helpers.ParseYamlString(valueString)
	}

	policySetting, err := client.UpdatePolicySetting(input)
	if err != nil {
		if !errors.FailedValidationError(err) {
			d.SetId("")
			return err
		}
		// so we have a data validation error - try using value as valueSource
		input["valueSource"] = input["value"]
		delete(input, "value")
		// try again
		policySetting, err = client.UpdatePolicySetting(input)
		if err != nil {
			d.SetId("")
			return err
		}
		// update state value setting with yaml parsed valueSource
		setValueFromValueSource(input["valueSource"].(string), d)
	}

	// NOTE: TemplateInput can be string or array of strings
	// - In case of string, we return string
	// - In array of strings, we return a valid YAML string
	templateInput, err := helpers.InterfaceToStringOrYaml(policySetting.TemplateInput)
	if err != nil {
		return err
	}

	//assign read properties
	d.Set("precedence", policySetting.Precedence)
	d.Set("template", policySetting.Template)
	d.Set("template_input", templateInput)
	d.Set("note", policySetting.Note)
	d.Set("valid_from_timestamp", policySetting.ValidFromTimestamp)
	d.Set("valid_to_timestamp", policySetting.ValidToTimestamp)
	d.Set("type", policySetting.Type.Uri)
	return nil
}

func setValueFromValueSource(valueSource string, d *schema.ResourceData) {
	var i interface{}
	yaml.Unmarshal([]byte(valueSource), &i)
	d.Set("value", fmt.Sprintf("%v", i))
	d.Set("value_source", valueSource)
	d.Set("value_source_used", true)
}

func resourceTurbotPolicySettingDelete(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*apiClient.Client)
	id := d.Id()
	err := client.DeletePolicySetting(id)
	if err != nil {
		return err
	}

	// clear the id to show we have deleted
	d.SetId("")

	return nil
}

func resourceTurbotPolicySettingImport(d *schema.ResourceData, meta interface{}) ([]*schema.ResourceData, error) {
	if err := resourceTurbotPolicySettingRead(d, meta); err != nil {
		return nil, err
	}
	return []*schema.ResourceData{d}, nil
}

func buildPayload(d *schema.ResourceData) map[string]string {
	commandPayload := map[string]string{
		"value":              d.Get("value").(string),
		"precedence":         d.Get("precedence").(string),
		"template":           d.Get("template").(string),
		"templateInput":      d.Get("template_input").(string),
		"note":               d.Get("note").(string),
		"validFromTimestamp": d.Get("valid_from_timestamp").(string),
		"validToTimestamp":   d.Get("valid_to_timestamp").(string),
	}
	// remove nil entries from commandPayload
	for k, v := range commandPayload {
		if v == "" {
			delete(commandPayload, k)
		}
	}
	return commandPayload
}

// If a pgp key is present, value will be encrypted so we cannot perform diff
// If valueSource was used, suppress diff if value source matches
func suppressIfEncryptedOrValueSourceMatches(_, old, new string, d *schema.ResourceData) bool {
	// if old value is not set, do not suppress - cannot be encrypted and value source will not have been used
	if old == "" {
		return false
	}

	_, keyPresent := d.GetOk("pgp_key")

	// Return true if the diff should be suppressed, false to retain it.
	if d.Get("value_source_used").(bool) {
		old = d.Get("value_source").(string)
	}
	return keyPresent || new == old
}

// write value and value_source to ResourceData, encrypting if a pgp key was provided
func storeValue(d *schema.ResourceData, setting *apiClient.PolicySetting) error {
	// NOTE: turbot policy settings have a value and a valueSource property
	// - value is the type property value, with the type dependent on the policy schema
	// - valueSource is the yaml representation of the policy.

	if pgpKey, ok := d.GetOk("pgp_key"); ok {
		// NOTE: If it is a complex type (object/array) then the diff calculation will use the value_source so the precise format is not critical
		// format the value as a string to allow us to handle object/array values using a string schema
		valueFingerprint, encryptedValue, err := helpers.EncryptValue(pgpKey.(string), helpers.InterfaceToString(setting.Value))
		if err != nil {
			return err
		}
		d.Set("value", encryptedValue)
		d.Set("value_key_fingerprint", valueFingerprint)

		valueSourceFingerprint, encryptedValueSource, err := helpers.EncryptValue(pgpKey.(string), setting.ValueSource)
		if err != nil {
			return err
		}
		d.Set("value_source", encryptedValueSource)
		d.Set("value_source_key_fingerprint", valueSourceFingerprint)
	} else {
		d.Set("value", helpers.InterfaceToString(setting.Value))
		d.Set("value_source", setting.ValueSource)
	}

	return nil
}

func suppressIfTemplateInputEquivalent(k, old, new string, d *schema.ResourceData) bool {
	if old == "" {
		return false
	}
	equivalent, err := helpers.YamlStringsAreEqual(old, new)
	if err != nil {
		return false
	}

	return equivalent
}
