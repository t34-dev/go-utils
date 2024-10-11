package main

import (
	"encoding/json"
	"fmt"
	"github.com/t34-dev/go-utils/pkg/http"
	"time"
)

func GetCurrentIP(client *http.Client) (string, error) {
	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.Unmarshal([]byte(resp), &result)
	if err != nil {
		return "", err
	}

	if origin, ok := result["origin"].(string); ok {
		return origin, nil
	}

	return "", fmt.Errorf("unable to parse IP from response")
}

func main() {
	proxies := []string{
		"socks5://4b077H:qslYR7aFsO@46.8.56.219:1051",
		"http://4b077H:qslYR7aFsO@46.8.56.219:1050",
	}

	client := http.NewClient(
		http.WithTimeout(15*time.Second),
		http.WithProxy(proxies),
		http.WithRetryCount(2),
		http.WithRetryWaitTime(2*time.Second, 10*time.Second),
	)

	ip, err := GetCurrentIP(client)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Current IP: %s\n", ip)
	}

	fmt.Println("Proxy status:")
	for _, status := range client.GetProxyStatus() {
		fmt.Printf("%s: %v, Error: %v\n", status.URL, status.Working, status.Error)
	}
}
