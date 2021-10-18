package db

import (
	"app/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func (c *Client) GetAllServer() ([]model.OpcUAServer, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	var results []model.OpcUAServer

	opcCol := c.Client.Database(DatabaseName).Collection("opcua")
	cursor, err := opcCol.Find(ctx, bson.M{})
	if err != nil {
		return results, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result model.OpcUAServer
		if err := cursor.Decode(&result); err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

func (c *Client) AddNewServer(server model.OpcUAServer) error {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	opcCol := c.Client.Database(DatabaseName).Collection("opcua")
	err := opcCol.FindOne(ctx, bson.M{"name": server.Name}).Decode(&model.OpcUAServer{})
	if err == nil { // existed
		return ErrINVALIDDATA
	}

	_, err = opcCol.InsertOne(ctx, bson.D{
		{Key: "name", Value: server.Name},
		{Key: "endpoint", Value: server.Endpoint},
		{Key: "policy", Value: server.Policy},
		{Key: "mode", Value: server.Mode},
		{Key: "cert", Value: server.Cert},
		{Key: "key", Value: server.Key},
		{Key: "nodeid", Value: server.NodeID},
	})

	if err != nil {
		return err
	}
	return nil
}

//
func (c *Client) GetServerByID(name string) (model.OpcUAServer, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	var result model.OpcUAServer

	opcCol := c.Client.Database(DatabaseName).Collection("opcua")
	if err := opcCol.FindOne(ctx, bson.M{"name": name}).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

//
func (c *Client) DeleteServer(name string) error {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	opcCol := c.Client.Database(DatabaseName).Collection("opcua")
	result, _ := opcCol.DeleteOne(ctx, bson.M{"name": name})
	if result.DeletedCount == 0 {
		return ErrINVALIDDATA
	}

	return nil
}
