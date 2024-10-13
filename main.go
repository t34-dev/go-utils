package main

import (
	"encoding/json"
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/t34-dev/go-utils/pkg/proxy"
	"go.uber.org/zap"
	"time"
)

type User struct {
	Name string
}

func GetCurrentIP(client proxy.Client) (string, error) {
	resp, err := client.Get("https://httpbin.org/get", nil, &User{Name: "test"})
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
		//log.Info(fmt.Sprintf("[%s] %s: %s", level, proxy, msg))
	}
	debuggerMiddleware := func(method, url string, req *resty.Request, userData interface{}) {
		fmt.Printf("DEBBUG: %s %s BODY:%+v UserData:%+v\n", method, req.URL, req.Body, userData)
	}

	// Создаем первый клиент
	client1 := proxy.NewClient(
		resty.New().
			SetTimeout(10*time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(1*time.Second).
			SetRetryMaxWaitTime(3*time.Second),
		proxy.WithProxy(proxies),
		proxy.WithLogFunc(logFunc),
		proxy.WithMiddleware(debuggerMiddleware),
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
		fmt.Printf("%s: Working: %v, Error: %#v\n", proxy.MaskProxyPassword(elem.URL), elem.Working, elem.Error)
	}

	// Формируем список рабочих прокси
	workingProxies := []string{}
	for _, elem := range proxyStatus {
		if elem.Working {
			workingProxies = append(workingProxies, elem.URL)
		}
	}

	// Создаем второй клиент только с рабочими прокси
	client2 := proxy.NewClient(
		resty.New().
			SetTimeout(10*time.Second).
			SetRetryCount(3).
			SetRetryWaitTime(1*time.Second).
			SetRetryMaxWaitTime(3*time.Second),
		proxy.WithProxy(workingProxies),
		proxy.WithLogFunc(logFunc),
		proxy.WithMiddleware(debuggerMiddleware),
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

	fmt.Println("==========")
	fmt.Println("Test middleware:")
	name := struct {
		Name string
		Age  int
	}{
		Name: "John",
		Age:  30,
	}
	req := client2.R()
	req.SetBody(name)
	resp, err := client2.Post("https://jsonplaceholder.typicode.com/posts", req, &name)
	fmt.Println(resp, err)
}
