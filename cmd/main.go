package main

import (
	"log"

	"github.com/Confialink/wallet-files/internal/config"
	"github.com/Confialink/wallet-files/internal/di"
	"github.com/Confialink/wallet-files/internal/routes"
	"github.com/Confialink/wallet-pkg-env_mods"
	"github.com/gin-gonic/gin"
)

var (
	appConfig *config.Config
)

// main: main function
func main() {
	c := di.Container
	appConfig = c.Config()
	ginMode := env_mods.GetMode(appConfig.Env)
	gin.SetMode(ginMode)

	ginRouter := routes.GetRouter()

	log.Printf("Starting API on port: %s", appConfig.Port)

	// Start proto buf server
	go c.PbServer().Start()

	// Start gin server
	ginRouter.Run(":" + appConfig.Port)
}
