package cache

import (
	"context"
	"errors"

	"golang.org/x/sync/singleflight"
)

var (
	ErrShotLimits = errors.New("error shot limits")
	ErrTypeError  = errors.New("wrong type")
)

type (
	SingleFlight struct {
		*Cache
		sfGroup *singleflight.Group
	}
)

// SingleFlightAPI
//
//	 usage:
//	 ```go
//	    sf := cli.SingleFlightAPI()
//		for _, k := range keys {
//		   go ... sf.Get(ctx, k) ...
//		}
//	 ```
func (cli *Cache) SingleFlightAPI() SingleFlight {
	return SingleFlight{
		Cache:   cli,
		sfGroup: new(singleflight.Group),
	}
}

func (sf SingleFlight) Get(ctx context.Context, key string) (string, error) {
	iResp, err, _ := sf.sfGroup.Do(key, func() (interface{}, error) {
		cmd := sf.Client.Get(ctx, key)
		if err := cmd.Err(); err != nil {
			return nil, err
		}
		return cmd.Val(), nil
	})
	if err != nil {
		return "", err
	}
	resp, ok := iResp.(string)
	if !ok {
		return "", ErrTypeError
	}
	return resp, nil
}
