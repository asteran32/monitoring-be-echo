package db

import (
	"app/model"
	"context"
	"errors"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func (c *Client) UserSignIn(email, password string) (model.User, error) {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	var dbUser model.User
	// check email
	userCol := c.Client.Database(DatabaseName).Collection("user")
	if err := userCol.FindOne(ctx, bson.M{"email": email}).Decode(&dbUser); err != nil {
		return dbUser, err
	}

	if !checkPassword(dbUser.Password, password) {
		return dbUser, ErrINVALIDPASSWORD
	}
	dbUser.Password = ""

	return dbUser, nil

}

func (c *Client) UserSignUp(user model.User) error {
	ctx, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	var dbUser model.User
	// Check email(unique)
	userCol := c.Client.Database(DatabaseName).Collection("user")
	err := userCol.FindOne(ctx, bson.M{"email": user.Email}).Decode(&dbUser)
	if err == nil { // email has existed
		return ErrINVALIDDATA
	}

	if err = hashPassword(&user.Password); err != nil {
		return err
	}

	_, err = userCol.InsertOne(ctx, bson.D{
		{Key: "firstname", Value: user.FirstName},
		{Key: "lastname", Value: user.LastName},
		{Key: "email", Value: user.Email},
		{Key: "password", Value: user.Password},
	})

	if err != nil {
		return err
	}

	return nil
}

// Check the password is correct or not.
// This method will return an error if the hash does not match the provided password string.
func checkPassword(existingHash, incomingPass string) bool {
	return bcrypt.CompareHashAndPassword([]byte(existingHash), []byte(incomingPass)) == nil
}

// Get the hash value of a password.
func hashPassword(s *string) error {
	if s == nil {
		return errors.New("Reference provided for hashing password is nil")
	}
	sBytes := []byte(*s)                                                        // Convert password string to byte slice.
	hashedBytes, err := bcrypt.GenerateFromPassword(sBytes, bcrypt.DefaultCost) // Obtain hashed password.
	if err != nil {
		return err
	}
	*s = string(hashedBytes[:]) // Update password string with the hashed version.
	return nil
}
