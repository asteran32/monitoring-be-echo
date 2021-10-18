package db

import (
	"app/model"
	"context"
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	ErrINVALIDPASSWORD = errors.New("Invalid password")
	ErrINVALIDDATA     = errors.New("Already associated")

	DatabaseName = "testApp"
)

type DBInterface interface {
	// user
	UserSignIn(string, string) (model.User, error)
	UserSignUp(model.User) error
	// cams
	GetAllCam() ([]model.Camera, error)
	AddNewCam(model.Camera) error
	GetCamByID(string) (model.Camera, error)
	DeleteCam(string) error
	// server
	GetAllServer() ([]model.OpcUAServer, error)
	AddNewServer(model.OpcUAServer) error
	GetServerByID(string) (model.OpcUAServer, error)
	DeleteServer(string) error
}

type Client struct {
	*mongo.Client
}

func NewClient() (*Client, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		fmt.Print(err)
		return &Client{}, fmt.Errorf("DB Connection err:%v", err)
	}

	return &Client{client}, nil
}
