package main

import (
	"os"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	//jwt "github.com/blockstatecom/gin-jwt"
	_ "github.com/bridgefinance-net/bridgefinance/api/docs"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr"
	"github.com/bridgefinance-net/bridgefinance/pkg/bfr/helpers/kafkaHelper"
	"github.com/bridgefinance-net/bridgefinance/pkg/http/restServer/bfrApi"
	"github.com/bridgefinance-net/bridgefinance/pkg/http/restServer/kafkaApi"
	models "github.com/bridgefinance-net/bridgefinance/pkg/model"
	"github.com/gin-contrib/cache"
	"github.com/gin-contrib/cache/persistence"
	"github.com/gin-gonic/contrib/static"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	ginSwagger "github.com/swaggo/gin-swagger"
	"github.com/swaggo/gin-swagger/swaggerFiles"
)

// @Title bridgefinance.org API
// @Version 1.0
// @description The world's crowd-driven financial data community has a professional API made for you.
// @description <h2>Decentral and transparent by design</h2>
// @description With our decentral approach to data verification, you can gain a deep insight into current and past pricing, volume and exchange info so you can make the right decisions to stay ahead of the game.
// @description
// @description <h3>Find the right data for your needs</h3>
// @description Show your users the most transparent data on the market with our API. Whether you're building a financial service, a portfolio management tool, a new mebfr offering, or more, we have the most advanced and updated data on the market for your product.
// @description For Oracle usage see [github](https://github.com/bridgefinance-net/bridgefinance/tree/master/documentation/methodology/oracles.md).
// @description
// @description <h3>Backtest your strategies</h3>
// @description Use the most efficient and transparent crypto data to run simulations and backtest your trading or investing strategies. With crowd-aggregated hundreds of exchanges you can be sure that you're getting the right picture every single time.
// @description
// @description <h3>Run Experiments</h3>
// @description Build your own models with our data, to further your interest or just for fun. With our flexible and powerful API, we provide you with a set of data that will help you draw insights and make conclusions.
// @description
// @description <h3>Request your data</h3>
// @description Set a bounty on gitcoin.io or drop us [line](mailto:API@bridgefinance.org).
// @license.name GNU GPLv3
// @Host api.bridgefinance.org
// @BasePath /

type login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

// User demo
type User struct {
	UserName  string
	FirstName string
	LastName  string
}

func GetTradesBlock(c *gin.Context) {
	kafkaApi.Process(c, kafkaHelper.TopicTradesBlock)
}

func GetFiltersBlock(c *gin.Context) {
	kafkaApi.Process(c, kafkaHelper.TopicFiltersBlock)
}

func GetTrades(c *gin.Context) {
	kafkaApi.Process(c, kafkaHelper.TopicTrades)
}

const (
	cachingTimeShort = time.Minute * 2
	cachingTimeLong  = time.Minute * 100
)

var identityKey = "id"

func helloHandler(c *gin.Context) {
	claims := jwt.ExtractClaims(c)
	c.JSON(200, gin.H{
		"userID": claims["id"],
		"text":   "Hello World.",
	})
}

func main() {

	r := gin.Default()
	r.Use(gin.Logger())
	r.Use(gin.Recovery())

	config := bfr.GetConfigApi()

	// the jwt middleware
	authMiddleware, err := jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "party zone",
		Key:         []byte(config.SecretKey), // TOFIX: this secret key should be different from the accepted apikey and secret key downstairs
		MaxRefresh:  time.Hour,
		IdentityKey: identityKey,
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*User); ok {
				return jwt.MapClaims{
					identityKey: v.UserName,
				}
			}
			return jwt.MapClaims{}
		},
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &User{
				UserName: claims["id"].(string),
			}
		},
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginVals login
			if err := c.ShouldBind(&loginVals); err != nil {
				return "", jwt.ErrMissingLoginValues
			}
			userID := loginVals.Username
			password := loginVals.Password

			if userID == config.ApiKey && password == config.SecretKey { // Temporary: only 1 valid key so far.
				return &User{
					UserName:  userID,
					LastName:  "Bo-Yi",
					FirstName: "Wu",
				}, nil
			}

			log.Warning("Authenticator ErrFailedAuthentication")
			return nil, jwt.ErrFailedAuthentication
		},
		Authorizator: func(data interface{}, c *gin.Context) bool {
			if v, ok := data.(*User); ok && v.UserName == config.ApiKey {
				return true
			}
			log.Warning("Authorizator rejected")
			return false
		},
		Unauthorized: func(c *gin.Context, code int, message string) {
			c.JSON(code, gin.H{
				"code":    code,
				"message": message,
			})
		},
		// TokenLookup is a string in the form of "<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>"
		// - "query:<name>"
		// - "cookie:<name>"
		TokenLookup: "header: Authorization, query: token, cookie: jwt",
		// TokenLookup: "query:token",
		// TokenLookup: "cookie:token",

		// TokenHeadName is a string in the header. Default value is "Bearer"
		TokenHeadName: "Bearer",

		// TimeFunc provides the current time. You can override it to use another time value. This is useful for testing or if your server uses a different time zone than your tokens.
		TimeFunc: time.Now,
	})

	r.POST("/login", authMiddleware.LoginHandler)

	auth := r.Group("/auth")
	auth.Use(authMiddleware.MiddlewareFunc())
	{
		auth.GET("/hello", helloHandler)
		auth.GET("/refresh_token", authMiddleware.RefreshHandler)
	}

	kafka := r.Group("/kafka")
	{
		kafka.GET("/tradesBlock", GetTradesBlock)
		kafka.GET("/filtersBlock", GetFiltersBlock)
		kafka.GET("/trades", GetTrades)
	}

	memoryStore := persistence.NewInMemoryStore(time.Second)

	store, err := models.NewDataStore()
	if err != nil {
		log.Errorln("NewDataStore", err)
	}
	bfrApiEnv := &bfrApi.Env{store}

	bfrAuth := r.Group("/v1")
	bfrAuth.Use(authMiddleware.MiddlewareFunc())
	{
		bfrAuth.POST("/supply", bfrApiEnv.PostSupply)
	}

	bfr := r.Group("/v1")
	{
		// Endpoints for cryptocurrencies/exchanges
		bfr.GET("/quotation/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetQuotation))
		bfr.GET("/supply/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetSupply))
		bfr.GET("/supplies/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetSupplies))
		bfr.GET("/symbol/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetSymbolDetails))
		bfr.GET("/symbols", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetAllSymbols))
		bfr.GET("/volume/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetVolume))
		bfr.GET("/coins", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCoins))
		bfr.GET("/pairs", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetPairs))
		bfr.GET("/exchanges", cache.CachePage(memoryStore, cachingTimeLong, bfrApiEnv.GetExchanges))
		bfr.GET("/defiLendingProtocols", cache.CachePage(memoryStore, cachingTimeLong, bfrApiEnv.GetLendingProtocols))
		bfr.GET("/chartPoints/:filter/:exchange/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetChartPoints))
		bfr.GET("/chartPointsAllExchanges/:filter/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetChartPointsAllExchanges))
		bfr.GET("/cviIndex", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCviIndex))
		bfr.GET("/defiLendingRate/:protocol/:asset", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetDefiRate))
		bfr.GET("/defiLendingRate/:protocol/:asset/:time", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetDefiRate))
		bfr.GET("/defiLendingState/:protocol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetDefiState))
		bfr.GET("/defiLendingState/:protocol/:time", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetDefiState))

		// Endpoints for interestrates
		bfr.GET("/interestrates", cache.CachePage(memoryStore, cachingTimeLong, bfrApiEnv.GetRates))
		bfr.GET("/interestrate/:symbol", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetInterestRate))
		bfr.GET("/interestrate/:symbol/:time", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetInterestRate))
		bfr.GET("/compoundedRate/:symbol/:dpy", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCompoundedRate))
		bfr.GET("/compoundedRate/:symbol/:dpy/:time", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCompoundedRate))
		bfr.GET("/compoundedAvg/:symbol/:days/:dpy", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCompoundedAvg))
		bfr.GET("/compoundedAvg/:symbol/:days/:dpy/:time", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCompoundedAvg))
		bfr.GET("/compoundedAvgbfr/:symbol/:days/:dpy", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCompoundedAvgbfr))
		bfr.GET("/compoundedAvgbfr/:symbol/:days/:dpy/:time", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetCompoundedAvgbfr))

		// Endpoints for fiat currencies
		bfr.GET("/fiatQuotations", cache.CachePage(memoryStore, cachingTimeShort, bfrApiEnv.GetFiatQuotations))
	}

	r.Use(static.Serve("/v1/chart", static.LocalFile("/charts", true)))
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// This environment variable is either set in docker-compose or empty
	executionMode := os.Getenv("EXEC_MODE")
	if executionMode == "production" {
		r.Run(":8080")
	} else {
		r.Run(":8081")
	}

}
