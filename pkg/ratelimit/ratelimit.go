package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Rate struct {
	Limit    int
	Duration time.Duration
}

type RateLimitter struct {
	apiConf   map[string]Rate
	client    *redis.Client
	scriptsha string
}

func NewRateLimitter(conf map[string]Rate) *RateLimitter {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	rt := &RateLimitter{apiConf: conf, client: rdb}
	script := `
local l = redis.call('GET', KEYS[1])
if not l then
	redis.call('SET', KEYS[1], 1, 'NX', 'EX', ARGV[1])
	return '1'
elseif tonumber(l) < tonumber(ARGV[2]) then
	redis.call('INCR', KEYS[1])
	return '1'
else
	return '0'
end
	`
	scriptsha, err := rt.client.ScriptLoad(context.Background(), script).Result()
	if err != nil {
		panic(err)
	}
	rt.scriptsha = scriptsha
	return rt
}

func (rt *RateLimitter) CheckThrottle(api string) (bool, error) {
	ctx := context.Background()
	rate, ok := rt.apiConf[api]
	if !ok {
		return false, fmt.Errorf("API Not configured: %s", api)
	}
	key := "ratelimit:" + api
	return rt.client.EvalSha(ctx, rt.scriptsha, []string{key}, int(rate.Duration.Seconds()), rate.Limit).Bool()
}
