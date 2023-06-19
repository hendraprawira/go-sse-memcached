package router

import (
	client "alert-map-service/app/controller/client"
	"alert-map-service/app/db"

	docs "alert-map-service/docs"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

func Routes() *gin.Engine {
	// gin.SetMode(gin.ReleaseMode) // for release mod, uncomment if need it
	r := gin.Default()

	db.ConnectMemcached()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"POST", "OPTIONS", "GET", "PUT", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	// Set a lower memory limit for multipart forms (default is 32 MiB)
	r.MaxMultipartMemory = 8 << 20 // 8 MiB

	apiUri := r.Group("/api")
	docs.SwaggerInfo.BasePath = "/api"

	clientSchema := apiUri.Group("")
	{
		clientSchema.GET("/docs/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
		clientSchema.GET("/clients", client.Clients)
		clientSchema.GET("/clientSSE", client.ClientSSE)
		clientSchema.GET("/data-stream", client.DataStreamWithMemchaced)

		clientSchema.POST("/client", client.AddClient)
		clientSchema.GET("/client", client.GetDataByMemchaced)
	}

	return r
}
