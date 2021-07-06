package simpleforce

import (
	"log"
	"os"
	"testing"

	"github.com/jarcoal/httpmock"
)

var (
	sfUser  = os.ExpandEnv("${SF_USER}")
	sfPass  = os.ExpandEnv("${SF_PASS}")
	sfToken = os.ExpandEnv("${SF_TOKEN}")
	sfURL   = func() string {
		if os.ExpandEnv("${SF_URL}") != "" {
			return os.ExpandEnv("${SF_URL}")
		} else {
			return DefaultURL
		}
	}()
)

func activateHttpMock() {
}

func registerLoginMock(c *Client) {
	loginResp := `<?xml version="1.0" encoding="utf-8" ?>
		<env:Envelope>
			<env:Body>
				<env:loginResponse>
					<env:result>
						<env:serverUrl>https://na0-api.salesforce.com/services/Soap/c/2.5</env:serverUrl>
						<env:sessionId>sessionId</env:sessionId>
						<env:userId>userId</env:userId>
						<env:userInfo>
							<env:userEmail>userEmail</env:userEmail>
							<env:userFullName>userFullName</env:userFullName>
							<env:userName>userName</env:userName>
						</env:userInfo>
					</env:result>
				</env:loginResponse>
			</env:Body>
		</env:Envelope>`
	mockURL := "https://login.salesforce.com//services/Soap/u/" + c.apiVersion
	httpmock.RegisterResponder("POST", mockURL,
		httpmock.NewStringResponder(200, loginResp))
}

func requireClient(t *testing.T, skippable bool) *Client {
	httpmock.Activate()
	httpmock.Reset()

	client := NewClient(sfURL, DefaultClientID, DefaultAPIVersion)
	if client == nil {
		t.Fail()
	}

	registerLoginMock(client)

	err := client.LoginPassword(sfUser, sfPass, sfToken)
	if err != nil {
		t.Fatal()
	}
	return client
}

func TestClient_LoginPassword_success(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := NewClient(sfURL, DefaultClientID, DefaultAPIVersion)
	if client == nil {
		t.Fatal()
	}

	registerLoginMock(client)

	// Use token
	err := client.LoginPassword(sfUser, sfPass, sfToken)
	if err != nil {
		t.Fail()
	} else {
		log.Println(logPrefix, "sessionID:", client.sessionID)
	}
}

func TestClient_LoginPassword_fail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := NewClient(sfURL, DefaultClientID, DefaultAPIVersion)
	if client == nil {
		t.Fatal()
	}

	loginErrResp := `<?xml version="1.0" encoding="UTF-8"?>
		<soapenv:Envelope
			xmlns:soapenv="http://schemas.xmlsoap.org/soap/envelope/"
			xmlns:sf="urn:fault.partner.soap.sforce.com"
			xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
			<soapenv:Body>
				<soapenv:Fault>
					<faultcode>INVALID_LOGIN</faultcode>
					<faultstring>INVALID_LOGIN: Invalid username, password, security token; or user locked out.</faultstring>
					<detail>
						<sf:LoginFault xsi:type="sf:LoginFault">
						<sf:exceptionCode>INVALID_LOGIN</sf:exceptionCode>
						<sf:exceptionMessage>Invalid username, password, security token; or user locked out.</sf:exceptionMessage>
						</sf:LoginFault>
					</detail>
				</soapenv:Fault>
			</soapenv:Body>
		</soapenv:Envelope>`
	mockURL := "https://login.salesforce.com//services/Soap/u/" + client.apiVersion
	httpmock.RegisterResponder("POST", mockURL,
		httpmock.NewStringResponder(500, loginErrResp))

	err := client.LoginPassword("__INVALID_USER__", "__INVALID_PASS__", "__INVALID_TOKEN__")
	if err == nil {
		t.Fail()
	}
}

func TestClient_LoginPasswordNoToken(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := NewClient(sfURL, DefaultClientID, DefaultAPIVersion)
	if client == nil {
		t.Fatal()
	}

	registerLoginMock(client)

	// Trusted IP must be configured AND the request must be initiated from the trusted IP range.
	err := client.LoginPassword(sfUser, sfPass, "")
	if err != nil {
		t.FailNow()
	} else {
		log.Println(logPrefix, "sessionID:", client.sessionID)
	}
}

func TestClient_LoginOAuth(t *testing.T) {

}

func TestClient_Query(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	client := requireClient(t, true)

	mockURL := "https://na0-api.salesforce.com/services/data/v" + client.apiVersion + "/query?q=SELECT%20Id%2CLastModifiedById%2CLastModifiedDate%2CParentId%2CCommentBody%20FROM%20CaseComment"

	httpmock.RegisterResponder("GET", mockURL,
		httpmock.NewStringResponder(200, `{"TotalSize": 0, "Done": true, "NextRecordsURL": "NextRecordsURL", "records": []}`))

	q := "SELECT Id,LastModifiedById,LastModifiedDate,ParentId,CommentBody FROM CaseComment"
	result, err := client.Query(q)
	if err != nil {
		log.Println(logPrefix, "query failed,", err)
		t.FailNow()
	}
}

func TestMain(m *testing.M) {
	m.Run()
}
