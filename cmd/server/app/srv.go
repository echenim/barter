package server

import (
	"fmt"
	"log"

	hdl "github.com/echenim/barter/internal/handlers"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/labstack/echo/v4"
)

const exchangePrivateKey = "4f3edf983ac636a65a842ce7c78d9aa706d3b113bce9c46f30d7d21715b23b1d"

func StartServer() {
	e := echo.New()
	e.HTTPErrorHandler = httpErrorHandler

	client, err := ethclient.Dial("http://localhost:8545")
	if err != nil {
		log.Fatal(err)
	}

	ex, err := hdl.NewExchange(exchangePrivateKey, client)
	if err != nil {
		log.Fatal(err)
	}

	ex.RegisterUser("829e924fdf021ba3dbbc4225edfece9aca04b929d6e75613329ca6f1d31c0bb4", 8)
	ex.RegisterUser("a453611d9419d0e56f499079478fd72c37b251a94bfde4d19872c44cf65386e3", 7)
	ex.RegisterUser("e485d098507f54e7733a205420dfddbe58db035fa577fc294ebd14db90767a52", 666)

	e.POST("/order", ex.PlaceOrder)

	e.GET("/trades/:market", ex.GetTrades)
	e.GET("/order/:userID", ex.GetOrders)
	e.GET("/book/:market", ex.GetBook)
	e.GET("/book/:market/bid", ex.GetBestBid)
	e.GET("/book/:market/ask", ex.GetBestAsk)

	e.DELETE("/order/:id", ex.CancelOrder)

	e.Start(":3000")
}

func httpErrorHandler(err error, c echo.Context) {
	fmt.Println(err)
}
