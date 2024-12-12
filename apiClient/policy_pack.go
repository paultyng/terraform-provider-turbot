package apiClient

func (client *Client) CreatePolicyPack(input map[string]interface{}) (*PolicyPack, error) {
	query := createPolicyPackMutation()
	responseData := &PolicyPackResponse{}
	variables := map[string]interface{}{
		"input": input,
	}

	// execute api call
	if err := client.doRequest(query, variables, responseData); err != nil {
		return nil, client.handleCreateError(err, input, "policy pack")
	}
	return &responseData.PolicyPack, nil
}

func (client *Client) ReadPolicyPack(id string) (*PolicyPack, error) {
	query := readPolicyPackQuery(id)
	responseData := &PolicyPackResponse{}

	// execute api call
	if err := client.doRequest(query, nil, responseData); err != nil {
		return nil, client.handleReadError(err, id, "smart folder")
	}
	return &responseData.PolicyPack, nil
}

func (client *Client) UpdatePolicyPack(input map[string]interface{}) (*PolicyPack, error) {
	query := updatePolicyPackMutation()
	responseData := &PolicyPackResponse{}
	variables := map[string]interface{}{
		"input": input,
	}
	// execute api call
	if err := client.doRequest(query, variables, responseData); err != nil {
		return nil, client.handleUpdateError(err, input, "smart folder")
	}
	return &responseData.PolicyPack, nil
}
