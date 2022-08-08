package mongo

import (
	"context"
	"gitlab.com/g6834/team26/auth/internal/domain/errors"
	"gitlab.com/g6834/team26/auth/internal/domain/models"
	"go.mongodb.org/mongo-driver/bson"
)

//var _ ports.UserStorage = (*Database)(nil)

func (db *Database) GetUserByLogin(ctx context.Context, login string) (*models.User, error) {
	var user models.User

	res := db.UserCollection().FindOne(ctx, bson.M{"login": login})
	if err := res.Decode(&user); err != nil {
		return nil, errors.ErrUserNotFound
	}

	return &user, nil
}

func (db *Database) Create(ctx context.Context, login, password string) (err error) {
	_, err = db.UserCollection().InsertOne(ctx, bson.M{
		"login":    login,
		"password": password,
	})
	if err != nil {
		return err
	}

	return nil
}

/*func (db *Database) FindOne(ctx context.Context, id string) (*models.User, error) {
	return nil, nil
}*/
