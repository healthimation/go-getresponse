# go-getresponse
A golang API client for [GetResponse V3](https://apidocs.getresponse.com/v3)

## Supported APIs
- [Contacts](https://apidocs.getresponse.com/v3/resources/contacts)

## Usage

```sh
go get github.com/healthimation/go-getresponse/getresponse
```

```golang
import (
    "log"
    "time"
    "context"

    "github.com/healthimation/go-getresponse/getresponse"
)

func main() {
    timeout := 5 * time.Second
    client := getresponse.NewClient("my get response api key", timeout)

    campaignID := "123"
    err := client.(context.Background(), "jsmith@example.com", "John Smith", nil, campaignID, nil, nil)
    if err != nil {
        log.Printf("Error creating contact in GR: %s", err.Error())
    }
}
```
