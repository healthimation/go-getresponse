package getresponse

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/healthimation/go-client/client"
	"github.com/healthimation/go-glitch/glitch"
)

//Error codes
const (
	ErrorAPI = "ERROR_API"

	// described @ https://apidocs.getresponse.com/v3/errors
	ErrorInternalError           = 1
	ErrorValidationError         = 1000
	ErrorRelatedResourceNotFound = 1001
	ErrorForbidden               = 1002
	ErrorInvalidParameterFormat  = 1003
	ErrorInvalidHash             = 1004
	ErrorMissingParameter        = 1005
	ErrorInvalidParameterType    = 1006
	ErrorInvalidParameterLength  = 1007
	ErrorResourceAlreadyExists   = 1008
	ErrorResourceInUse           = 1009
	ErrorExternalError           = 1010
	ErrorMessageAlreadySending   = 1011
	ErrorMessageParsing          = 1012
	ErrorResourceNotFound        = 1013
	ErrorAuthenticationFailure   = 1014
	ErrorRequestQuotaReached     = 1015
	ErrorTemporarilyBlocked      = 1016
	ErrorPermanentlyBlocked      = 1017
	ErrorIPBlocked               = 1018
	ErrorInvalidRequestHeaders   = 1021
)

// Client can make requests to the GR api
type Client interface {
	// CreateContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.create
	CreateContact(ctx context.Context, email string, name *string, dayOfCycle *int32, campaignID string, customFields []CustomField, ipAddress *string) glitch.DataError

	// GetContacts - https://apidocs.getresponse.com/v3/resources/contacts#contacts.get.all
	GetContacts(ctx context.Context, queryHash map[string]string, fields []string, sortHash map[string]string, page int32, perPage int32, additionalFlags *string) ([]Contact, glitch.DataError)

	// Get Contact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.get
	GetContact(ctx context.Context, ID string, fields []string) (Contact, glitch.DataError)

	// UpdateContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.update
	UpdateContact(ctx context.Context, ID string, newData Contact) (Contact, glitch.DataError)

	// UpdateContactCustomFields - https://apidocs.getresponse.com/v3/resources/contacts#contacts.upsert.custom-fields
	UpdateContactCustomFields(ctx context.Context, ID string, customFields []CustomField) (Contact, glitch.DataError)

	// DeleteContact - https://apidocs.getresponse.com/v3/resources/contacts#contacts.delete
	DeleteContact(ctx context.Context, ID string, messageID string, ipAddress string) glitch.DataError
}

type getResponseClient struct {
	c      client.BaseClient
	apiKey string
}

// NewClient returns a new pushy client
func NewClient(apiKey string, timeout time.Duration) Client {
	return &getResponseClient{
		c:      client.NewBaseClient(findGetResponse, "getresponse", true, timeout),
		apiKey: apiKey,
	}
}

func (g *getResponseClient) CreateContact(ctx context.Context, email string, name *string, dayOfCycle *int32, campaignID string, customFields []CustomField, ipAddress *string) glitch.DataError {
	slug := "/v3/contacts"
	h := http.Header{}
	h.Set("Content-type", "application/json")
	h.Set("X-Auth-Token", fmt.Sprintf("api-key %s", g.apiKey))

	bodyObj := createContactRequest{
		Email:             email,
		Name:              name,
		DayOfCycle:        dayOfCycle,
		Campaign:          Campaign{CampaignID: campaignID},
		CustomFieldValues: customFields,
		IPAddress:         ipAddress,
	}

	body, err := client.ObjectToJSONReader(bodyObj)
	if err != nil {
		return err
	}

	status, ret, err := g.c.MakeRequest(ctx, http.MethodPost, slug, nil, h, body)
	if err != nil {
		return err
	}

	if status < 200 || status >= 400 {
		//parse error
		return g.parseError(ret)
	}

	return nil
}

func (g *getResponseClient) GetContacts(ctx context.Context, queryHash map[string]string, fields []string, sortHash map[string]string, page int32, perPage int32, additionalFlags *string) ([]Contact, glitch.DataError) {
	slug := "/v3/contacts"
	h := http.Header{}
	h.Set("Content-type", "application/json")
	h.Set("X-Auth-Token", fmt.Sprintf("api-key %s", g.apiKey))

	query := url.Values{}
	for k, v := range queryHash {
		query.Set(fmt.Sprintf("query[%s]", k), v)
	}

	for k, v := range sortHash {
		query.Set(fmt.Sprintf("sort[%s]", k), v)
	}

	if len(fields) > 0 {
		query.Set("fields", strings.Join(fields, ","))
	}

	query.Set("page", strconv.Itoa(int(page)))
	query.Set("perPage", strconv.Itoa(int(perPage)))

	if additionalFlags != nil {
		query.Set("additionalFlags", *additionalFlags)
	}

	result := make([]Contact, 0)
	status, ret, err := g.c.MakeRequest(ctx, http.MethodGet, slug, query, h, nil)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, glitch.NewDataError(err, client.ErrorDecodingResponse, fmt.Sprintf("Could not unmarshal response: %s", ret))
	}

	return result, nil
}

func (g *getResponseClient) GetContact(ctx context.Context, ID string, fields []string) (Contact, glitch.DataError) {
	slug := fmt.Sprintf("/v3/contacts/%s", ID)
	h := http.Header{}
	h.Set("Content-type", "application/json")
	h.Set("X-Auth-Token", fmt.Sprintf("api-key %s", g.apiKey))

	query := url.Values{}
	if len(fields) > 0 {
		query.Set("fields", strings.Join(fields, ","))
	}

	result := Contact{}
	status, ret, err := g.c.MakeRequest(ctx, http.MethodGet, slug, query, h, nil)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, glitch.NewDataError(err, client.ErrorDecodingResponse, fmt.Sprintf("Could not unmarshal response: %s", ret))
	}

	return result, nil
}

func (g *getResponseClient) UpdateContact(ctx context.Context, ID string, newData Contact) (Contact, glitch.DataError) {
	result := Contact{}
	slug := fmt.Sprintf("/v3/contacts/%s", ID)
	h := http.Header{}
	h.Set("Content-type", "application/json")
	h.Set("X-Auth-Token", fmt.Sprintf("api-key %s", g.apiKey))

	body, err := client.ObjectToJSONReader(newData)
	if err != nil {
		return result, err
	}

	status, ret, err := g.c.MakeRequest(ctx, http.MethodPost, slug, nil, h, body)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, glitch.NewDataError(err, client.ErrorDecodingResponse, fmt.Sprintf("Could not unmarshal response: %s", ret))
	}

	return result, nil
}

func (g *getResponseClient) UpdateContactCustomFields(ctx context.Context, ID string, customFields []CustomField) (Contact, glitch.DataError) {
	result := Contact{}
	slug := fmt.Sprintf("/v3/contacts/%s/custom-fields", ID)
	h := http.Header{}
	h.Set("Content-type", "application/json")
	h.Set("X-Auth-Token", fmt.Sprintf("api-key %s", g.apiKey))

	bodyObj := updateCustomFieldRequest{customFields}
	body, err := client.ObjectToJSONReader(bodyObj)
	if err != nil {
		return result, err
	}

	status, ret, err := g.c.MakeRequest(ctx, http.MethodPost, slug, nil, h, body)
	if err != nil {
		return result, err
	}

	if status < 200 || status >= 400 {
		//parse error
		return result, g.parseError(ret)
	}

	jErr := json.Unmarshal(ret, &result)
	if jErr != nil {
		return result, glitch.NewDataError(err, client.ErrorDecodingResponse, fmt.Sprintf("Could not unmarshal response: %s", ret))
	}

	return result, nil
}

func (g *getResponseClient) DeleteContact(ctx context.Context, ID string, messageID string, ipAddress string) glitch.DataError {
	slug := fmt.Sprintf("/v3/contacts/%s", ID)
	h := http.Header{}
	h.Set("Content-type", "application/json")
	h.Set("X-Auth-Token", fmt.Sprintf("api-key %s", g.apiKey))

	query := url.Values{}
	query.Set("messageId", messageID)
	query.Set("ipAddress", ipAddress)

	status, ret, err := g.c.MakeRequest(ctx, http.MethodDelete, slug, query, h, nil)
	if err != nil {
		return err
	}

	if status < 200 || status >= 400 {
		//parse error
		return g.parseError(ret)
	}

	return nil
}

func (g *getResponseClient) parseError(resp []byte) glitch.DataError {
	errRet := ErrorResponse{}
	err := json.Unmarshal(resp, &errRet)
	if err != nil {
		return glitch.NewDataError(err, client.ErrorDecodingError, fmt.Sprintf("Could not unmarshal error response: %s", resp))
	}
	return glitch.NewDataError(nil, fmt.Sprintf("%d", errRet.ErrorCode), fmt.Sprintf("%s | context: %s", errRet.Message, strings.Join(errRet.Context, ", ")))
}
