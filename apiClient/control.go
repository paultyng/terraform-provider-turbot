package apiClient

import (
	"fmt"
)

func (client *Client) ReadControl(args string) (*Control, error) {
	query := readControlQuery(args)
	var responseData = &ReadControlResponse{}

	// execute api call
	err := client.doRequest(query, nil, responseData)
	if err != nil {
		return nil, fmt.Errorf("error reading control: %s", err.Error())
	}
	control := responseData.Control

	return &control, nil
}

func (client *Client) MuteControl(input map[string]interface{}) (*MuteControl, error) {
	query := muteControlMutation()
	responseData := &MuteControlResponse{}
	variables := map[string]interface{}{
		"input": input,
	}

	// execute api call
	if err := client.doRequest(query, variables, responseData); err != nil {
		return nil, client.handleUpdateError(err, input, "control")
	}
	return &responseData.MuteControl, nil
}

func (client *Client) UnMuteControl(input map[string]interface{}) (*MuteControl, error) {
	query := unMuteControlMutation()
	responseData := &MuteControlResponse{}
	variables := map[string]interface{}{
		"input": input,
	}

	// execute api call
	if err := client.doRequest(query, variables, responseData); err != nil {
		return nil, client.handleUpdateError(err, input, "control")
	}
	return &responseData.MuteControl, nil
}
