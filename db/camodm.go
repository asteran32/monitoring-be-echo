package db

import (
	"app/model"
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// Get all camera list
func (c *Client) GetAllCam() ([]model.Camera, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	var results []model.Camera

	camCol := c.Client.Database(DatabaseName).Collection("camera")
	cursor, err := camCol.Find(ctx, bson.M{})
	if err != nil {
		return results, err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var result model.Camera
		if err := cursor.Decode(&result); err != nil {
			return results, err
		}
		results = append(results, result)
	}

	return results, nil
}

//
func (c *Client) AddNewCam(cam model.Camera) error {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	camCol := c.Client.Database(DatabaseName).Collection("camera")
	err := camCol.FindOne(ctx, bson.M{"name": cam.Name}).Decode(&model.Camera{})
	if err == nil { // existed
		return ErrINVALIDDATA
	}

	_, err = camCol.InsertOne(ctx, bson.D{
		{Key: "name", Value: cam.Name},
		{Key: "rtsp", Value: cam.Rtsp},
		{Key: "codec", Value: cam.Codec},
	})

	if err != nil {
		return err
	}
	return nil
}

//
func (c *Client) GetCamByID(name string) (model.Camera, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	var result model.Camera

	camCol := c.Client.Database(DatabaseName).Collection("camera")
	if err := camCol.FindOne(ctx, bson.M{"name": name}).Decode(&result); err != nil {
		return result, err
	}

	return result, nil
}

//
func (c *Client) DeleteCam(name string) error {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	camCol := c.Client.Database(DatabaseName).Collection("camera")
	result, _ := camCol.DeleteOne(ctx, bson.M{"name": name})
	if result.DeletedCount == 0 {
		return ErrINVALIDDATA
	}

	return nil
}
