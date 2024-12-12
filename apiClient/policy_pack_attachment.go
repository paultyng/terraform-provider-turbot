package apiClient

import (
	"fmt"
)

func (client *Client) CreatePolicyPackAttachment(input map[string]interface{}) (*TurbotResourceMetadata, error) {
	query := createPolicyPackAttachmentMutation()
	responseData := &CreatePolicyPackAttachResponse{}

	variables := map[string]interface{}{
		"input": input,
	}

	// execute api call
	if err := client.doRequest(query, variables, responseData); err != nil {
		return nil, client.handleCreateError(err, input, "policy pack attachment")
	}
	return &responseData.PolicyPackAttach.Turbot, nil
}

func (client *Client) DeletePolicyPackAttachment(input map[string]interface{}) error {
	query := detachPolicyPackAttachmentMutation()
	var responseData interface{}

	variables := map[string]interface{}{
		"input": input,
	}
	// execute api call
	if err := client.doRequest(query, variables, responseData); err != nil {
		return fmt.Errorf("error deleting policy pack attachment: %s", err.Error())
	}
	return nil
}
