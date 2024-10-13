package datastore

import (
	"context"
	"time"

	"github.com/redis/rueidis"
)

func NewRueidis() func() (rueidis.Client, error) {
	return func() (rueidis.Client, error) {
		client, err := rueidis.NewClient(rueidis.ClientOption{
			InitAddress: []string{"localhost:6379"},
		})
		if err != nil {
			return nil, err
		}

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		cmd := client.B().Ping().Build()

		err = client.Do(ctx, cmd).Error()
		if err != nil {
			return nil, err
		}

		return client, nil
	}
}
