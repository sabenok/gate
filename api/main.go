package api

import (
	"fmt"
	"net/http"
	"os"

	"github.com/Depado/ginprom"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/noah-blockchain/explorer-gate/core"
	"github.com/noah-blockchain/explorer-gate/handlers"
	"github.com/tendermint/tendermint/libs/pubsub"
)

// Run API
func Run(gateService *core.NoahGate, pubsubServer *pubsub.Server) {
	router := SetupRouter(gateService, pubsubServer)
	err := router.Run(fmt.Sprintf("%s:%s", os.Getenv("GATE_API_LINK"), os.Getenv("GATE_API_PORT")))
	if err != nil {
		panic(err)
	}
}

//Setup router
func SetupRouter(gateService *core.NoahGate, pubsubServer *pubsub.Server) *gin.Engine {
	router := gin.Default()
	if os.Getenv("DEBUG_MODE") == "false" {
		gin.SetMode(gin.ReleaseMode)
	}

	p := ginprom.New(
		ginprom.Engine(router),
		ginprom.Subsystem("gin"),
		ginprom.Path("/metrics"),
	)
	router.Use(p.Instrument())
	router.Use(cors.Default())                           // CORS
	router.Use(gin.ErrorLogger())                        // print all errors
	router.Use(gin.Recovery())                           // returns 500 on any code panics
	router.Use(apiMiddleware(gateService, pubsubServer)) // init global context

	router.GET(`/`, handlers.Index)

	v1 := router.Group("/api/v1")
	{
		v1.GET(`/estimate/tx-commission`, handlers.EstimateTxCommission)
		v1.GET(`/estimate/coin-buy`, handlers.EstimateCoinBuy)
		v1.GET(`/estimate/coin-sell`, handlers.EstimateCoinSell)
		v1.GET(`/estimate/coin-sell-all`, handlers.EstimateCoinSellAll)
		v1.GET(`/nonce/:address`, handlers.GetNonce)
		v1.GET(`/min-gas`, handlers.GetMinGas)
		v1.POST(`/transaction/push`, handlers.PushTransaction)
	}
	// Default handler 404
	router.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": 404, "log": "Resource not found."}})
	})
	return router
}

//Add necessary services to global context
func apiMiddleware(gate *core.NoahGate, pubsubServer *pubsub.Server) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("gate", gate)
		c.Set("pubsub", pubsubServer)
		c.Next()
	}
}
