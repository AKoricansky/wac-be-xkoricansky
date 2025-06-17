package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/AKoricansky/wac-be-xkoricansky/api"
	"github.com/AKoricansky/wac-be-xkoricansky/internal/ambulance_counseling_wl"
	"github.com/AKoricansky/wac-be-xkoricansky/internal/db_service"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	log.Printf("Server started")
	port := os.Getenv("AMBULANCE_COUNSELING_API_PORT")
	if port == "" {
		port = "8080"
	}

	engine := gin.New()
	engine.Use(gin.Recovery())
	corsMiddleware := cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "PUT", "POST", "DELETE", "PATCH"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{""},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	})
	engine.Use(corsMiddleware)

	userDbService := db_service.NewMongoService[ambulance_counseling_wl.User](db_service.MongoServiceConfig{
		DbName:     "ambulance-counseling",
		Collection: "users",
	})

	questionDbService := db_service.NewMongoService[ambulance_counseling_wl.Question](db_service.MongoServiceConfig{
		DbName:     "ambulance-counseling",
		Collection: "questions",
	})

	replyDbService := db_service.NewMongoService[ambulance_counseling_wl.Reply](db_service.MongoServiceConfig{
		DbName:     "ambulance-counseling",
		Collection: "replies",
	})

	ctx := context.Background()
	defer func() {
		if err := userDbService.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from user database: %v", err)
		}
		if err := questionDbService.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from question database: %v", err)
		}
		if err := replyDbService.Disconnect(ctx); err != nil {
			log.Printf("Error disconnecting from reply database: %v", err)
		}
	}()

	handleFunctions := &ambulance_counseling_wl.ApiHandleFunctions{
		AmbulanceCounselingAPI:     ambulance_counseling_wl.NewAmbulanceCounselingApi(questionDbService, replyDbService),
		AmbulanceCounselingAuthAPI: ambulance_counseling_wl.NewAmbulanceCounselingAuthApi(userDbService),
	}
	ambulance_counseling_wl.NewRouterWithGinEngine(engine, *handleFunctions)

	engine.GET("/openapi", api.HandleOpenApi)
	engine.Run(":" + port)
}
