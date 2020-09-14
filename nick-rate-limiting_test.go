package main

import (
	"github.com/go-redis/redis/v8"
	"strconv"
	"testing"
	"time"
)

//正确的json
const jsonStr = `
[{
"type": "header,cookie,get",
"key": "keyName",
"value": "value1,value2,value3"
}, {
"type": "cookie",
"key": "orderId",
"value": "order1,order2,order3"
}]
`
const wrongJson = `
[{
type: "header,cookie,get",
"key": "keyName",
"value": "value1,value2,value3"
}, {
"type": "cookie",
"key": "orderId",
"value": "order1,order2,order3"
}]
`

const wrongJsonNoHeader = `
[{
"type": "",
"key": "keyName",
"value": "value1,value2,value3"
}, {
"type": "cookie",
"key": "orderId",
"value": "order1,order2,order3"
}]
`

const wrongJsonNoValue = `
[{
"type": "header,cookie,get",
"key": "keyName",
"value": "value1,value2,value3"
}, {
"type": "cookie",
"key": "orderId",
"value": ""
}]
`

const wrongJsonNokey = `
[{
"type": "header,cookie,get",
"key": "",
"value": "value1,value2,value3"
}, {
"type": "cookie",
"key": "orderId",
"value": "order1,order2,order3"
}]
`

//right redis host
const redisHostRight = "10.5.24.223"

//right redis port
const redisPortRight = 6379

//right redis auth
const redisAuthRight = "123"

//err redis host
const redisHostErr = "10.10.0.10"

// empty redis host
const redisHostEmpty = ""

//err redis host
const redisPortErr = 16379

type TestConfig struct {
	input    Config
	expected string
}

var configList = []TestConfig{
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "",
	},
	{
		input: Config{
			QPS:                 0,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "QPS must great than 0",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostEmpty,
			RedisPort:           redisPortErr,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  0,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "redis config required",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostErr,
			RedisPort:           redisPortErr,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  0,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "RedisTimeoutSecond must great than 0",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  "",
			RedisHost:           redisHostRight,
			RedisPort:           redisPortRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "LimitResourcesJson required",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  wrongJson,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "LimitResourcesJson with incorrect json format,invalid character 't' looking for beginning of object key string",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  wrongJsonNoHeader,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "LimitResourcesJson with empty value",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  wrongJsonNokey,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "LimitResourcesJson with empty value",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  wrongJsonNoValue,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		expected: "LimitResourcesJson with empty value",
	},
}

func TestCheckConfig(t *testing.T) {
	for _, conf := range configList {
		actual := conf.input.checkConfig()
		if actual != nil && actual.Error() != conf.expected {
			t.Errorf("checkConfig return: [%s], expected: [%s]", actual.Error(), conf.expected)
		}
	}
}

func TestRedisEval(t *testing.T) {
	options := &redis.Options{
		Addr:        redisHostRight + ":" + strconv.Itoa(redisPortRight),
		Password:    redisAuthRight,
		DB:          0,
		DialTimeout: time.Duration(2) * time.Second,
	}
	redisClient = redis.NewClient(options)
	limitKey := "kong:customratelimit:service:0ff4659d-f65a-453f-bc68-7aa49bcf3a80:route:ca5ae44b-bd8f-4b4b-948f-9ba67e35085f:second:1599811581"

	luaScript := `
		local key, value, expiration = KEYS[1], tonumber(ARGV[1]), ARGV[2]
		local newVal = redis.call("incrby", key, value)
		if newVal == value then
			redis.call("expire", key, expiration)
		end
		return newVal - 1
`
	result, err := redisClient.Eval(ctx, luaScript, []string{limitKey}, 1, 1).Result()
	if err != nil {
		t.Errorf("eval1 failed, %s", err.Error())
	}
	var expected int64
	expected = 0
	i := result.(int64)
	if i != expected {
		t.Errorf("get redis eval result err,actual:%d, expected:%d", i, expected)
	}
	result, err = redisClient.Eval(ctx, luaScript, []string{limitKey}, 1, 1).Result()
	if err != nil {
		t.Errorf("eval2 failed, %s", err.Error())
	}
	expected = 1
	i = result.(int64)
	if i != expected {
		t.Errorf("get redis eval result err,actual:%d, expected:%d", i, expected)
	}
}
