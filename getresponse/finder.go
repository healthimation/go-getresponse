package getresponse

import (
	"net/url"
)

func findGetResponse(serviceName string, useTLS bool) (url.URL, error) {
	ret, err := url.Parse("https://api.getresponse.com/")
	if err != nil || ret == nil {
		return url.URL{}, err
	}
	return *ret, err
}
