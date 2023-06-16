package controller

import (
	"alert-map-service/app/db"
	"alert-map-service/app/models"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bradfitz/gomemcache/memcache"
	"github.com/gin-gonic/gin"
)

func Clients(c *gin.Context) {
	var Client []models.Client
	var count int64
	// db.DB.Find(&MasterUser) <- for get All Data
	db.DB.Where("is_deleted = ?", false).Find(&Client)
	db.DB.Model(&Client).Where("is_deleted = ?", false).Count(&count)

	if count == 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"Total Data": count, "data": "Empty",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"Total Data": count, "data": Client})
}

func ClientSSE(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")

	clientChannel := make(chan models.Client)
	go func() {
		for {
			var clients []models.Client
			err := db.DB.Order("created_at desc").Limit(2).Find(&clients).Error
			if err != nil {
				log.Println("Error fetching messages:", err)
			}

			for i := len(clients) - 1; i >= 0; i-- {
				clientChannel <- clients[i]
			}

			time.Sleep(2 * time.Second) // Poll every 2 seconds
		}
	}()
	for {
		select {
		case client := <-clientChannel:
			data, err := json.Marshal(client)
			if err != nil {
				log.Println("Error marshaling message:", err)
				continue
			}

			fmt.Fprintf(c.Writer, "data: %s\n\n", string(data))
			c.Writer.Flush()

		case <-c.Writer.CloseNotify():
			// Client closed connection, stop sending events
			return
		}
	}

}

// @Summary Get a client with SSE
// @Description Get a client with SSE
// @Accept text/event-stream
// @Produce text/event-stream
// @Tags Users
// @Param        id    query     string  false  "query by id"  models.Client
// @Success 200 {object} models.Client
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /data-stream [get]
func DataStreamWithMemchaced(c *gin.Context) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	id := c.Query("id")
	clientChannel := make(chan []models.Client)

	// Goroutine to continuously fetch data from the database and send it over SSE
	go func() {
		for {
			var data []models.Client
			cacheKey := id

			// Try to fetch data from Memcached
			item, err := db.MC.Get(cacheKey)
			if err == nil {
				// Data found in Memcached, unmarshal it
				err = json.Unmarshal(item.Value, &data)
				if err != nil {
					log.Println("Error unmarshaling data from Memcached:", err)
				}
			} else if err == memcache.ErrCacheMiss {
				// Data not found in Memcached, fetch it from the database
				if err := db.DB.Find(&data).Error; err != nil {
					log.Println("Error fetching data from database:", err)
				}

				// Marshal and store the data in Memcached for future use
				dataJSON, err := json.Marshal(data)
				if err != nil {
					log.Println("Error marshaling data for Memcached:", err)
				} else {
					err = db.MC.Set(&memcache.Item{Key: cacheKey, Value: dataJSON})
					if err != nil {
						log.Println("Error storing data in Memcached:", err)
					}
				}
			} else {
				log.Println("Error fetching data from Memcached:", err)
			}
			clientChannel <- data

			time.Sleep(2 * time.Second) // Poll every 2 seconds
		}
	}()

	// SSE event stream loop
	for {
		select {
		case data := <-clientChannel:
			dataJSON, err := json.Marshal(data)
			if err != nil {
				log.Println("Error marshaling data:", err)
				continue
			}

			fmt.Fprintf(c.Writer, "data: %s\n\n", string(dataJSON))
			c.Writer.Flush()

		case <-c.Writer.CloseNotify():
			// Client closed the connection, stop sending events
			return
		}
	}

}

type ErrorResponse struct {
	Message string `json:"message"`
}

var clientChannels = make(chan []models.Client)

// @Summary Create a new user
// @Description Create a new user with the provided details
// @Tags Users
// @Accept json
// @Produce json
// @Param user body models.ClientModelAdd true "User object"
// @Success 200 {object} models.Client
// @Failure 400 {object} ErrorResponse
// @Router /client [post]
func AddClient(c *gin.Context) {
	var Client *models.Client

	if err := c.ShouldBindJSON(&Client); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}
	now := time.Now()
	newUser := models.Client{
		Username:  Client.Username,
		Email:     strings.ToLower(Client.Email),
		Fullname:  Client.Fullname,
		CreatedBy: 0,
		CreatedAt: now,
		UpdatedAt: now,
		UpdatedBy: 0,
		IsDeleted: false,
	}
	if err := db.DB.Create(&newUser).Error; err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	id := strconv.FormatUint(newUser.ID, 10)
	// Marshal and store the data in Memcached for future use
	log.Print("TEST ", id)
	dataJSON, err := json.Marshal(newUser)
	if err != nil {
		log.Println("Error marshaling data for Memcached:", err)
	} else {
		err = db.MC.Set(&memcache.Item{Key: id, Value: dataJSON})
		if err != nil {
			log.Println("Error storing data in Memcached:", err)
		}
	}
	c.JSON(http.StatusOK, gin.H{"status": "success", "master_user": newUser})

}

// @Summary Get a client by memchaced key value
// @Description Get a user by their memchaced key value
// @Tags Users
// @Param        id    query     string  false  "query by id"  models.Client
// @Success 200 {object} models.Client
// @Failure 404 {object} ErrorResponse
// @Failure 400 {object} ErrorResponse
// @Router /client [get]
func GetDataByMemchaced(c *gin.Context) {
	var data models.Client
	id := c.Query("id")
	cacheKey := id

	// Try to fetch data from Memcached
	item, err := db.MC.Get(cacheKey)
	if err == nil {
		// Data found in Memcached, unmarshal it
		err = json.Unmarshal(item.Value, &data)
		if err != nil {
			log.Println("Error unmarshaling data from Memcached:", err)
			c.JSON(http.StatusBadRequest, data)
		}
	} else {
		log.Println("Error fetching data from Memcached:", err)
		c.JSON(http.StatusBadRequest, data)
	}
	c.JSON(http.StatusOK, data)
}
