package routes

import (
	"encoding/json"
	"flare-indexer/services/api"
	"flare-indexer/services/utils"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/go-cmp/cmp"
)

type testData struct {
	request  string
	response string
}

var processAttestationRequests map[string]testData = map[string]testData{
	"request1": {
		request: `{
				"request": "0x0500a200000054141f408ab37e4ed1411c9db2f875c064f67024654a42bda597298628eefb8c213686a516f05a706ed2ae1f20b3bb69d1916a59468a3b8d31790e0269fa2c88bc000000"
			}`,
		response: `{
				"stateConnectorRound": 0,
				"merkleProof": null,
				"blockNumber": 188,
				"transactionHash": "0x213686a516f05a706ed2ae1f20b3bb69d1916a59468a3b8d31790e0269fa2c88",
				"transactionType": 0,
				"nodeId": "0xf29bce5f34a74301eb0de716d5194e4a4aea5d7a",
				"startTime": 1678825429,
				"endTime": 1678826629,
				"weight": 100000000000,
				"sourceAddress": "0x3cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c"
			}`,
	},
}

var prepareRequests map[string]testData = map[string]testData{
	"request1": {
		request: `{
				"attestationType": 5,
				"sourceId": 162,
				"messageIntegrityCode": "",
				"id": "0x213686a516f05a706ed2ae1f20b3bb69d1916a59468a3b8d31790e0269fa2c88",
				"blockNumber": 188
			}`,
		response: `{
				"stateConnectorRound": 0,
				"merkleProof": null,
				"blockNumber": 188,
				"transactionHash": "0x213686a516f05a706ed2ae1f20b3bb69d1916a59468a3b8d31790e0269fa2c88",
				"transactionType": 0,
				"nodeId": "0xf29bce5f34a74301eb0de716d5194e4a4aea5d7a",
				"startTime": 1678825429,
				"endTime": 1678826629,
				"weight": 100000000000,
				"sourceAddress": "0x3cb7d3842e8cee6a0ebd09f1fe884f6861e1b29c"
			}`,
	},
}

var integrityRequests map[string]testData = map[string]testData{
	"request1": {
		request: `{
				"attestationType": 5,
				"sourceId": 162,
				"messageIntegrityCode": "",
				"id": "0x213686a516f05a706ed2ae1f20b3bb69d1916a59468a3b8d31790e0269fa2c88",
				"blockNumber": 188
			}`,
		response: `"0x54141f408ab37e4ed1411c9db2f875c064f67024654a42bda597298628eefb8c"`,
	},
}

var prepareAttestationRequests map[string]testData = map[string]testData{
	"request1": {
		request: `{
				"attestationType": 5,
				"sourceId": 162,
				"messageIntegrityCode": "",
				"id": "0x213686a516f05a706ed2ae1f20b3bb69d1916a59468a3b8d31790e0269fa2c88",
				"blockNumber": 188
			}`,
		response: `"0x0500a200000054141f408ab37e4ed1411c9db2f875c064f67024654a42bda597298628eefb8c213686a516f05a706ed2ae1f20b3bb69d1916a59468a3b8d31790e0269fa2c88bc000000"`,
	},
}

func TestProcessAttestationRequest(t *testing.T) {
	qr := newQueryRouteHandlers(testContext)

	r := httptest.NewRequest(http.MethodPost, "/", utils.JsonToReader(t, processAttestationRequests["request1"].request))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	qr.processAttestationRequest(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
	verifyVerificationResponse(t, w.Result().Body, processAttestationRequests["request1"].response)
}

func TestPrepareRequest(t *testing.T) {
	qr := newQueryRouteHandlers(testContext)

	r := httptest.NewRequest(http.MethodPost, "/prepare", utils.JsonToReader(t, prepareRequests["request1"].request))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	qr.prepareRequest(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
	verifyVerificationResponse(t, w.Result().Body, prepareRequests["request1"].response)
}

func TestIntegrityRequest(t *testing.T) {
	qr := newQueryRouteHandlers(testContext)

	r := httptest.NewRequest(http.MethodPost, "/integrity", utils.JsonToReader(t, integrityRequests["request1"].request))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	qr.integrityRequest(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
	VerifyData(
		t,
		w.Result().Body,
		func(w api.ApiResponseWrapper[string]) string { return w.Data },
		integrityRequests["request1"].response,
	)
}

func TestPrepareAttestationRequest(t *testing.T) {
	qr := newQueryRouteHandlers(testContext)

	r := httptest.NewRequest(http.MethodPost, "/prepareAttestation", utils.JsonToReader(t, prepareAttestationRequests["request1"].request))
	r.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	qr.prepareAttestationRequest(w, r)

	if w.Result().StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, w.Result().StatusCode)
	}
	VerifyData(
		t,
		w.Result().Body,
		func(w api.ApiResponseWrapper[string]) string { return w.Data },
		prepareAttestationRequests["request1"].response,
	)

}

// Verify if actual response body's data matches expected data.
// testFunc is a function that takes the actual response body and returns the data to be compared to
// the expected data
func VerifyData[R any, T any](t *testing.T, actualBody io.Reader, testFunc func(api.ApiResponseWrapper[R]) T, expected string) {
	var wResponse api.ApiResponseWrapper[R]
	utils.DecodeStruct(t, actualBody, &wResponse)

	if wResponse.Status != api.ApiResStatusOk {
		t.Errorf("Expected status %s, got %s", api.ApiResStatusOk, wResponse.Status)
	}

	var response T
	err := json.Unmarshal([]byte(expected), &response)
	if err != nil {
		t.Fatal(err)
	}
	if !cmp.Equal(testFunc(wResponse), response) {
		t.Errorf("Response mismatch: %s", cmp.Diff(testFunc(wResponse), response))
	}
}

func verifyVerificationResponse(t *testing.T, actualBody io.Reader, expected string) {
	VerifyData(
		t,
		actualBody,
		func(r api.ApiResponseWrapper[api.APIVerification[api.ARPChainStaking, api.DHPChainStaking]]) api.DHPChainStaking {
			return *r.Data.Response
		},
		expected,
	)
}
