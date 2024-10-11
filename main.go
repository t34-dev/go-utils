package main

import (
	"encoding/json"
	"fmt"
	"github.com/t34-dev/go-utils/pkg/http"
	"go.uber.org/zap"
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

	log, _ := zap.NewDevelopment()
	defer log.Sync()

	logFunc := func(level string, msg string, fields ...interface{}) {
		zapFields := make([]zap.Field, len(fields)/2)
		for i := 0; i < len(fields); i += 2 {
			zapFields[i/2] = zap.Any(fields[i].(string), fields[i+1])
		}
		switch level {
		case "info":
			log.Info(msg, zapFields...)
		case "warn":
			log.Warn(msg, zapFields...)
		case "error":
			log.Error(msg, zapFields...)
		}
	}
	client := http.NewClient(
		http.WithTimeout(15*time.Second),
		http.WithProxy(proxies),
		http.WithRetryCount(5),
		http.WithRetryWaitTime(2*time.Second, 10*time.Second),
		http.WithLogFunc(logFunc),
	)

	ip, err := GetCurrentIP(client)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
	} else {
		fmt.Printf("Current IP: %s\n", ip)
	}

	fmt.Println("Proxy elem:")
	for _, elem := range client.GetProxyStatus() {
		fmt.Printf("%s: %v, Error: %v\n", elem.URL, elem.Working, elem.Error)
	}
}
