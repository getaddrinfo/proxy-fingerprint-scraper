package common

import (
	"encoding/json"
	"io"
	"net/http"
)

type FingerprintResponse struct {
	Fingerprint *string `json:"fingerprint"`
	Raw         []byte
}

type GetExperimentsResponse struct {
	Fingerprint *string `json:"fingerprint,omitempty"`
}

type Storable struct {
	Fingerprint string
	Proxy       string
}

func DecodeExperimentResponse(resp *http.Response) (GetExperimentsResponse, error) {
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return GetExperimentsResponse{}, err
	}

	var data GetExperimentsResponse
	err = json.Unmarshal(body, &data)

	if err != nil {
		return GetExperimentsResponse{}, err
	}

	return data, nil
}

func DecodeFingerprintBody(resp *http.Response) (FingerprintResponse, error) {
	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return FingerprintResponse{}, err
	}

	var data FingerprintResponse
	err = json.Unmarshal(body, &data)

	if err != nil {
		return FingerprintResponse{Raw: body}, err
	}

	data.Raw = body

	return data, nil
}
