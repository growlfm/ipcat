package ipcat

import (
    "encoding/json"
    "errors"
    "fmt"
    "io/ioutil"
    "log"
    "net/http"
    "regexp"
    "strings"
)

type AzureValueProperties struct {
    Region          string   `json:"region"`
    Platform        string   `json:"platform"`
    SystemService   string   `json:"systemService"`
    AddressPrefixes []string `json:"addressPrefixes"`
    NetworkFeatures []string `json:"networkFeatures"`
}

// AzureValue is Azure value in their IP ranges file
type AzureValue struct {
    Name       string                 `json:"name"`
    Id         string                 `json:"id"`
    Properties AzureValueProperties   `json:"properties"`
}

// Azure is main record for Azure IP info
type Azure struct {
    ChangeNumber int          `json:"changeNumber"`
    Cloud        string       `json:"cloud"`
    Values       []AzureValue `json:"values"`
}

var retried bool

var findPublicIPsURL = func() (string, error) {
    //  Ref: Azure IP Ranges and Service Tags – Public Cloud
    //  https://www.microsoft.com/en-us/download/details.aspx?id=56519
    downloadPage := "https://www.microsoft.com/en-us/download/confirmation.aspx?id=56519"

    resp, err := http.Get(downloadPage)
    if err != nil {
        return "", err
    }
    b, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return "", err
    }

    re := regexp.MustCompile("url=https://download.microsoft.com/download/.*/ServiceTags_Public_.*.json")
    addr := re.Find(b)

    if string(addr) == "" {
        return "", errors.New("could not find PublicIPs address on download page")
    }

    return string(addr)[4:], nil
}

// DownloadAzure downloads and returns raw bytes of the MS Azure ip
// range list
func DownloadAzure() ([]byte, error) {
    url, err := findPublicIPsURL()
    if err != nil {
        return nil, fmt.Errorf("Failed to find public IPs url: %s", err)
    }

    log.Printf("Fetching url, %s", url)
    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    if resp.StatusCode != 200 {
        return nil, fmt.Errorf("Failed to download Azure ranges: status code %s", resp.Status)
    }
    body, err := ioutil.ReadAll(resp.Body)
    if err != nil {
        return nil, err
    }
    resp.Body.Close()
    return body, nil
}

// UpdateAzure takes a raw data, parses it and updates the ipmap
func UpdateAzure(ipmap *IntervalSet, body []byte) error {
    const (
        dcName = "Microsoft Azure"
        dcURL  = "http://www.windowsazure.com/en-us/"
    )

    azure := Azure{}
    err := json.Unmarshal(body, &azure)
    if err != nil {
        log.Printf("Unable to unmarshall body, %s", body)
        return err
    }

    // delete all existing records
    ipmap.DeleteByName(dcName)

    for _, value := range azure.Values {
        if value.Id == "AzureCloud" {
            prop := value.Properties
            for _, prefix := range prop.AddressPrefixes {
                if strings.Count(prefix, ":") == 0 {
                    err = ipmap.AddCIDR(prefix, dcName, dcURL)
                    if err != nil {
                        return err
                    }
                }
            }
        }
    }

    return nil
}
