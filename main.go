package main

import (
	"encoding/json"
	"fmt"
	"github.com/t34-dev/go-utils/pkg/proxy"
	proxy_resty "github.com/t34-dev/go-utils/pkg/proxy/resty"
	"go.uber.org/zap"
	"time"
)

func GetCurrentIP(client proxy.HTTPClient) (string, error) {
	resp, err := client.Get("https://httpbin.org/get")
	if err != nil {
		return "", err
	}

	var result map[string]interface{}
	err = json.Unmarshal(resp.Body(), &result)
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
		"socks5://4b077H:qslYR7aFsO@46.8.56.219:1052",
		"socks5://4b077H:qslYR7aFsO1@46.8.56.219:1051",
		"socks5://4b077H1:qslYR7aFsO@46.8.56.219:1051",
		"socks5://4b077H:qslYR7aFsO@46.8.56.220:1051",
		"http://4b077H:qslYR7aFsO@46.8.56.219:1051",
		"http://4b077H:qslYR7aFsO1@46.8.56.219:1050",
		"http://4b077H1:qslYR7aFsO@46.8.56.220:1050",
		"http://4b077H:qslYR7aFsO@46.8.56.219:1050",
		"socks5://4b077H:qslYR7aFsO@46.8.56.219:1051",
	}
	log, _ := zap.NewDevelopment()
	defer log.Sync()
	logFunc := func(level, msg, proxy string) {
		log.Info(fmt.Sprintf("[%s] %s: %s", level, proxy, msg))
	}

	// Создаем первый клиент
	client1 := proxy_resty.NewRestyClient(
		proxy_resty.WithTimeout(10*time.Second),
		proxy_resty.WithRetryCount(3),
		proxy_resty.WithRetryWaitTime(1*time.Second, 3*time.Second),
		proxy.WithProxy(proxies),
		proxy.WithLogFunc(logFunc),
	)

	// Выполняем запрос с первым клиентом
	ip, err := GetCurrentIP(client1)
	if err != nil {
		fmt.Printf("Error with client1: %v\n", err)
	} else {
		fmt.Printf("Current IP with client1: %s\n", ip)
	}

	// Получаем статус прокси после выполнения запросов
	proxyStatus := client1.GetProxyStatus()
	fmt.Println("==========")
	fmt.Println("Proxy status after client1:")
	for _, elem := range proxyStatus {
		fmt.Printf("%s: Working: %v, Error: %v\n", proxy.MaskProxyPassword(elem.URL), elem.Working, elem.Error)
	}

	// Формируем список рабочих прокси
	workingProxies := []string{}
	for _, elem := range proxyStatus {
		if elem.Working {
			workingProxies = append(workingProxies, elem.URL)
		}
	}

	// Создаем второй клиент только с рабочими прокси
	client2 := proxy_resty.NewRestyClient(
		proxy_resty.WithTimeout(10*time.Second),
		proxy_resty.WithRetryCount(3),
		proxy_resty.WithRetryWaitTime(1*time.Second, 5*time.Second),
		proxy.WithProxy(workingProxies),
		proxy.WithLogFunc(logFunc),
	)

	// Выполняем запрос со вторым клиентом
	ip, err = GetCurrentIP(client2)
	if err != nil {
		fmt.Printf("Error with client2: %v\n", err)
	} else {
		fmt.Printf("Current IP with client2: %s\n", ip)
	}

	// Получаем финальный статус прокси
	finalProxyStatus := client2.GetProxyStatus()
	fmt.Println("==========")
	fmt.Println("Final proxy status:")
	for _, elem := range finalProxyStatus {
		fmt.Printf("%s: Working: %v, Error: %v\n", proxy.MaskProxyPassword(elem.URL), elem.Working, elem.Error)
	}
}
