package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/dgrijalva/jwt-go"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/krunal4amity/tronicscorp/config"
	"github.com/krunal4amity/tronicscorp/handlers"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"github.com/labstack/gommon/random"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	//CorrelationID is a request id unique to the request being made
	CorrelationID = "X-Correlation-ID"
)

var (
	c        *mongo.Client
	db       *mongo.Database
	prodCol  *mongo.Collection
	usersCol *mongo.Collection
	cfg      config.Properties
)

func init() {
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("Configuration cannot be read : %v", err)
	}
	ctx := context.Background()
	connectURI := fmt.Sprintf("mongodb://%s:%s", cfg.DBHost, cfg.DBPort)
	c, err := mongo.Connect(ctx, options.Client().ApplyURI(connectURI))
	if err != nil {
		log.Fatalf("Unable to connect to database : %v", err)
	}
	db = c.Database(cfg.DBName)
	prodCol = db.Collection(cfg.ProductCollection)
	usersCol = db.Collection(cfg.UsersCollection)

	isUserIndexUnique := true
	indexModel := mongo.IndexModel{
		Keys: bson.M{"username": 1},
		Options: &options.IndexOptions{
			Unique: &isUserIndexUnique,
		},
	}
	_, err = usersCol.Indexes().CreateOne(ctx, indexModel)
	if err != nil {
		log.Fatalf("Unable to create an index : %+v", err)
	}
}

func addCorrelationID(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// generate correlation id
		id := c.Request().Header.Get(CorrelationID)
		var newID string
		if id == "" {
			//generate a random number
			newID = random.String(12)
		} else {
			newID = id
		}
		c.Request().Header.Set(CorrelationID, newID)
		c.Response().Header().Set(CorrelationID, newID)
		return next(c)
	}
}

func adminMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		hToken := c.Request().Header.Get("x-auth-token") // Bearer slfjl2jrlk23j
		token := strings.Split(hToken, " ")[1]
		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(*jwt.Token) (interface{}, error) {
			return []byte(cfg.JwtTokenSecret), nil
		})
		if err != nil {
			return echo.NewHTTPError(http.StatusInternalServerError, "Unable to parse token")
		}
		if !claims["authorized"].(bool) {
			return echo.NewHTTPError(http.StatusForbidden, "Not authorized")
		}
		return next(c)
	}
}

func main() {
	e := echo.New()
	e.Logger.SetLevel(log.DEBUG)
	e.Pre(middleware.RemoveTrailingSlash())
	e.Pre(addCorrelationID)
	jwtMiddleware := middleware.JWTWithConfig(middleware.JWTConfig{
		SigningKey:  []byte(cfg.JwtTokenSecret),
		TokenLookup: "header:x-auth-token",
	})
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `${time_rfc3339_nano} ${remote_ip} ${header:X-Correlation-ID} ${host} ${method} ${uri} ${user_agent} ` +
			`${status} ${error} ${latency_human}` + "\n",
	}))
	h := &handlers.ProductHandler{Col: prodCol}
	uh := &handlers.UsersHandler{Col: usersCol}
	e.GET("/products/:id", h.GetProduct)
	e.DELETE("/products/:id", h.DeleteProduct, jwtMiddleware, adminMiddleware)
	e.PUT("/products/:id", h.UpdateProduct, middleware.BodyLimit("1M"), jwtMiddleware)
	e.POST("/products", h.CreateProducts, middleware.BodyLimit("1M"), jwtMiddleware)
	e.GET("/products", h.GetProducts)

	e.POST("/users", uh.CreateUser)
	e.POST("/auth", uh.AuthnUser)
	e.Logger.Infof("Listening on %s:%s", cfg.Host, cfg.Port)
	e.Logger.Fatal(e.Start(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)))
}
