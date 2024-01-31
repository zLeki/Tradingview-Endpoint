package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	ratelimitBucket []string
	accountData     AccountData
	client          http.Client
	openPositions   []*Position
)

const (
	BUY  = "POSITION_TYPE_BUY"
	SELL = "POSITION_TYPE_SELL"
)

func init() {
	GetPositions()
	Update()
}
func GetPositions() {
	url := fmt.Sprintf("%v/users/current/accounts/%v/positions", REGION, ACCOUNT)
	var err error
	_, openPositions, err = HandleHTTP(openPositions, url, "GET")
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range openPositions {
		if v.Type == BUY {
			v.Buy = true
		} else {
			v.Buy = false
		}
	}
	fmt.Println("Positions fetched successfully")
}
func HandleHTTP[T any](t T, url string, method string, payload ...io.Reader) (*http.Response, T, error) {
	var req *http.Request
	var err error
	if len(payload) > 0 {
		req, err = http.NewRequest(method, url, payload[0])
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	if err != nil {
		return nil, t, nil
	}
	req.Header.Add("Accept", "application/json")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("auth-token", AUTH)
	res, err := client.Do(req)
	if err != nil {
		return nil, t, nil
	}
	err = json.NewDecoder(res.Body).Decode(&t)
	if err != nil {
		return res, t, nil
	}
	return res, t, nil
}
func Update() {
	url := fmt.Sprintf("%s/users/current/accounts/%s/account-information", REGION, ACCOUNT)
	var err error
	_, accountData, err = HandleHTTP(accountData, url, "GET")
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Account data fetched successfully")
}

func main() {
	router := gin.New()
	gin.SetMode(gin.ReleaseMode)
	router.Use(gin.Logger())
	// router.POST("/contact", Email())
	router.POST("/execute", Execute())
	router.GET("/dash", Dashboard())
	router.StaticFile("/dash/logo.jpg", "./logo.jpg")
	//Handle 404
	router.NoRoute(func(c *gin.Context) {
		c.Redirect(http.StatusMovedPermanently, "/home")
	})
	fmt.Println(`
██╗     ███████╗██╗  ██╗██╗███████╗██╗  ██╗
██║     ██╔════╝██║ ██╔╝██║██╔════╝╚██╗██╔╝
██║     █████╗  █████╔╝ ██║█████╗   ╚███╔╝ 
██║     ██╔══╝  ██╔═██╗ ██║██╔══╝   ██╔██╗ 
███████╗███████╗██║  ██╗██║██║     ██╔╝ ██╗
╚══════╝╚══════╝╚═╝  ╚═╝╚═╝╚═╝     ╚═╝  ╚═╝`)
	fmt.Println(len(openPositions), "position(s) open!")
	router.Run(":80")
}

func Dashboard() gin.HandlerFunc {
	return func(c *gin.Context) {
		Update()
		GetPositions()
		c.HTML(http.StatusOK, "home.html", gin.H{
			"Account":   accountData,
			"Positions": openPositions,
		})
	}
}

// ---------------------------------- TRADE SERVER ----------------------------------
// ------------------------------- DO NOT EDIT WHILE TIRED --------------------------
const (
	QUANTITIY = 0.50
	LONG      = 1    // DO NOT CHANGE
	SHORT     = 2    // DO NOT CHANGE
	EXIT      = iota // DO NOT CHNAGE
	REGION    = "https://mt-client-api-v1.london.agiliumtrade.ai"
	ACCOUNT   = "" // CHANGE
	AUTH      = "" // CHANGE
	REVERSE   = false
)

var (
	Positions    = make(map[string]string, 0)
	positionSize = 0.00
	msg          string
)

func Execute() gin.HandlerFunc {
	return func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		msg = string(body)
		GetPositions()
		if strings.Contains(msg, "Buy") {
			for _, v := range openPositions {
				if v.Type == SELL && !REVERSE {
					Exit(v.Id)
				} else if REVERSE && v.Type == BUY {
					Exit(v.Id)
				}
			}
			if !REVERSE {
				ExecuteTrade(LONG)
			} else {
				ExecuteTrade(SHORT)
			}
		} else if strings.Contains(msg, "Sell") {
			for _, v := range openPositions {
				if v.Type == BUY && !REVERSE {
					Exit(v.Id)
				} else if REVERSE && v.Type == SELL {
					Exit(v.Id)
				}
			}
			if !REVERSE {
				ExecuteTrade(SHORT)
			} else {
				ExecuteTrade(LONG)
			}
		} else if strings.Contains(msg, "Exit") {
			for _, v := range openPositions {
				Exit(v.Id)
			}
			fmt.Println("All positions exited, have a good day.")
		} else {
			fmt.Println("Invalid command")
		}

	}
}

func ExecuteTrade(Position int) error {
	switch Position {
	case LONG:
		go Market("BUY")
	case SHORT:
		go Market("SELL")
	case EXIT:

	}
	return nil
}

func Market(condition string, exitAmt ...float64) {
	fmt.Println("Executing", condition)
	url := REGION + "/users/current/accounts/" + ACCOUNT + "/trade"
	method := "POST"
	var payload *strings.Reader
	if len(exitAmt) > 0 {
		formatted_float := strconv.FormatFloat(exitAmt[0], 'f', 2, 64)
		payload = strings.NewReader(fmt.Sprintf(`{
			"actionType": "ORDER_TYPE_%s",
			"symbol":"US30.cash",
			"volume": %v
			}`, condition, formatted_float))
	} else {
		formatted_float := strconv.FormatFloat(QUANTITIY, 'f', 2, 64)
		payload = strings.NewReader(fmt.Sprintf(`{
			"actionType": "ORDER_TYPE_%s",
			"symbol":"US30.cash",
			"volume": %v
			}`, condition, formatted_float))
	}
	// client := &http.Client{}
	// req, err := http.NewRequest(method, url, payload)

	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// req.Header.Add("Content-Type", "application/json")
	// req.Header.Add("Accept", "application/json")
	// req.Header.Add("auth-token", AUTH)
	// res, err := client.Do(req)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return
	// }
	// defer res.Body.Close()
	var body Response
	res, body, err := HandleHTTP(body, url, method, payload)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		fmt.Println("Failed to execute trade", res.Status, positionSize)
		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(body)
	} else if res.StatusCode == 200 {
		fmt.Println(body)
		if strings.Contains(body.StringCode, "ERR") && !strings.Contains(body.StringCode, "NO_ERROR") {
			fmt.Println("Failed to execute trade\n", body.Message, res.Status, body, positionSize)
			return
		}
		fmt.Println("Trade executed successfully NEW POSITION SIZE = ", positionSize, res.Status)
	} else {
		fmt.Println("Failed to execute trade")

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(body))
	}
}
func Exit(id string) {
	url := REGION + "/users/current/accounts/" + ACCOUNT + "/trade"
	method := "POST"

	payload := strings.NewReader(`{
"actionType": "POSITION_CLOSE_ID",
"positionId":"` + id + `"
}`)
	var response Response
	res, _, err := HandleHTTP(response, url, method, payload)
	if err != nil {
		log.Fatal(err)
	}

	if res.StatusCode != 200 {
		defer res.Body.Close()

		body, err := io.ReadAll(res.Body)
		if err != nil {
			fmt.Println(err)
			return
		}
		fmt.Println(string(body))
	} else {
		if !strings.Contains(response.Message, "No error returned") {
			fmt.Println("Failed to exit position\n", response.Message)
			return
		}
		fmt.Println("Position successfully exited")
		delete(Positions, id)
		positionSize = 0.00
	}
}

// STRUCT DUMP
type Position struct {
	Id               string    `json:"id"`
	Type             string    `json:"type"`
	Symbol           string    `json:"symbol"`
	Magic            int       `json:"magic"`
	Time             time.Time `json:"time"`
	BrokerTime       string    `json:"brokerTime"`
	OpenPrice        float64   `json:"openPrice"`
	CurrentPrice     float64   `json:"currentPrice"`
	CurrentTickValue float64   `json:"currentTickValue"`
	StopLoss         float64   `json:"stopLoss"`
	Volume           float64   `json:"volume"`
	Swap             float64   `json:"swap"`
	Profit           float64   `json:"profit"`
	UpdateTime       time.Time `json:"updateTime"`
	Commission       float64   `json:"commission"`
	ClientId         string    `json:"clientId"`
	UnrealizedProfit float64   `json:"unrealizedProfit"`
	RealizedProfit   float64   `json:"realizedProfit"`
	Buy              bool      `json:"buy"`
}
type AccountData struct {
	Broker                      string  `json:"broker"`
	Currency                    string  `json:"currency"`
	Server                      string  `json:"server"`
	Balance                     float64 `json:"balance"`
	Equity                      float64 `json:"equity"`
	Margin                      float64 `json:"margin"`
	FreeMargin                  float64 `json:"freeMargin"`
	Leverage                    int     `json:"leverage"`
	MarginLevel                 float64 `json:"marginLevel"`
	Type                        string  `json:"type"`
	Name                        string  `json:"name"`
	Login                       string  `json:"login"`
	Credit                      int     `json:"credit"`
	Platform                    string  `json:"platform"`
	MarginMode                  string  `json:"marginMode"`
	TradeAllowed                bool    `json:"tradeAllowed"`
	InvestorMode                bool    `json:"investorMode"`
	AccountCurrencyExchangeRate int     `json:"accountCurrencyExchangeRate"`
}
type Response struct {
	NumericCode int    `json:"numericCode"`
	StringCode  string `json:"stringCode"`
	Message     string `json:"message"`
	OrderID     string `json:"orderId"`
	PositionID  string `json:"positionId"`
}
