package main

import (
	"fmt"
	"github.com/Kong/go-pdk"
	"github.com/go-redis/redis/v8"
	"strconv"
	"sync"
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
const redisHostRight = "localhost"

//right redis port
const redisPortRight = 6379

//right redis auth
const redisAuthRight = ""

//err redis host
const redisHostErr = "10.10.0.10"

// empty redis host
const redisHostEmpty = ""

//err redis host
const redisPortErr = 16379

type TestConfig struct {
	input                Config
	identifier           string
	unix                 int64
	confExpected         string
	prefixExpected       string
	rateLimitKeyExpected string
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
		confExpected:         "",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
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
			RedisLimitKeyPrefix: "",
			HideClientHeader:    false,
		},
		confExpected:         "Key: 'Config.QPS' Error:Field validation for 'QPS' failed on the 'required' tag",
		prefixExpected:       "kong:customratelimit:",
		identifier:           "",
		unix:                 0,
		rateLimitKeyExpected: "kong:customratelimit::qps:0",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostEmpty,
			RedisPort:           redisPortErr,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest:",
			HideClientHeader:    false,
		},
		confExpected:         "Key: 'Config.RedisHost' Error:Field validation for 'RedisHost' failed on the 'required' tag",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "111",
		unix:                 1111,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:111:qps:1111",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostRight,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		confExpected:         "Key: 'Config.RedisPort' Error:Field validation for 'RedisPort' failed on the 'required' tag",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostRight,
			RedisPort:           0,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		confExpected:         "Key: 'Config.RedisPort' Error:Field validation for 'RedisPort' failed on the 'required' tag",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostRight,
			RedisPort:           65536,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  2,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		confExpected:         "Key: 'Config.RedisPort' Error:Field validation for 'RedisPort' failed on the 'lte' tag",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortErr,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  0,
			RedisDB:             0,
			RedisLimitKeyPrefix: "test",
			HideClientHeader:    false,
		},
		confExpected:         "Key: 'Config.RedisTimeoutSecond' Error:Field validation for 'RedisTimeoutSecond' failed on the 'required' tag",
		prefixExpected:       "test:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "test:kong:customratelimit:username-nick:qps:1600067356",
	},
	{
		input: Config{
			QPS:                 30,
			Log:                 true,
			LimitResourcesJson:  jsonStr,
			RedisHost:           redisHostRight,
			RedisPort:           redisPortErr,
			RedisAuth:           redisAuthRight,
			RedisTimeoutSecond:  -1,
			RedisDB:             0,
			RedisLimitKeyPrefix: "nicktest",
			HideClientHeader:    false,
		},
		confExpected:         "Key: 'Config.RedisTimeoutSecond' Error:Field validation for 'RedisTimeoutSecond' failed on the 'gt' tag",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
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
		confExpected:         "",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
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
		confExpected:         "LimitResourcesJson with incorrect json format,invalid character 't' looking for beginning of object key string",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
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
		confExpected:         "LimitResourcesJson with empty value",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
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
		confExpected:         "LimitResourcesJson with empty value",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
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
		confExpected:         "LimitResourcesJson with empty value",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
	},
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
			MatchCondition:      "test",
		},
		confExpected:         "Key: 'Config.MatchCondition' Error:Field validation for 'MatchCondition' failed on the 'oneof' tag",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
	},
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
			MatchCondition:      "and",
		},
		confExpected:         "",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
	},
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
			MatchCondition:      "or",
		},
		confExpected:         "",
		prefixExpected:       "nicktest:kong:customratelimit:",
		identifier:           "username-nick",
		unix:                 1600067356,
		rateLimitKeyExpected: "nicktest:kong:customratelimit:username-nick:qps:1600067356",
	},
}

func TestCheckConfig(t *testing.T) {
	for _, conf := range configList {
		actual := conf.input.checkConfig()
		if actual != nil && actual.Error() != conf.confExpected {
			t.Errorf("checkConfig return: [%s], confExpected: [%s]", actual.Error(), conf.confExpected)
		}
		if actual == nil && conf.confExpected != "" {
			t.Errorf("checkConfig success, buf has wrong confExpected: [%s]", conf.confExpected)
		}
	}
}

func TestGetPrefix(t *testing.T) {
	for _, conf := range configList {
		actual := conf.input.getPrefix()
		if actual != conf.prefixExpected {
			t.Errorf("getPrefix return: [%s], prefixExpected: [%s]", actual, conf.prefixExpected)
		}
	}
}

func TestGetRateLimitKey(t *testing.T) {
	for _, conf := range configList {
		actual := conf.input.getRateLimitKey(conf.identifier, conf.unix)
		if actual != conf.rateLimitKeyExpected {
			t.Errorf("getRateLimitKey return: [%s], rateLimitKeyExpected: [%s]", actual, conf.rateLimitKeyExpected)
		}
	}
}

func getDefaultConf() *Config {
	return &Config{
		QPS:                 30,
		Log:                 false,
		LimitResourcesJson:  jsonStr,
		RedisHost:           redisHostRight,
		RedisPort:           redisPortRight,
		RedisAuth:           redisAuthRight,
		RedisTimeoutSecond:  2,
		RedisDB:             0,
		RedisLimitKeyPrefix: "nicktest",
		HideClientHeader:    false,
	}
}

func TestGetRemainingAndIncr(t *testing.T) {
	kong := &pdk.PDK{}
	conf := getDefaultConf()
	remaining, stop, _ := conf.getRemainingAndIncr(kong, "username-nick", 1600067356)
	if remaining != 30 && stop != false {
		t.Errorf("getRemainingAndIncr return: [%v %v], rateLimitKeyExpected: [%v %v]", remaining, stop, 30, false)
	}
}

func TestGetRemainingAndIncrConcurrent(t *testing.T) {
	kong := &pdk.PDK{}
	conf := getDefaultConf()
	var wg sync.WaitGroup
	for i := 0; i <= 100; i++ {
		wg.Add(1)
		go func(i int) {
			remaining, stop, _ := conf.getRemainingAndIncr(kong, "username-nick", 1600067356)
			fmt.Println(remaining, stop)
			wg.Done()
		}(i)
	}
	wg.Wait()
}

func TestInSlice(t *testing.T) {
	list := []struct {
		search   string
		slice    []string
		expected bool
	}{
		{
			search: "test1",
			slice: []string{
				"test1",
				"test2",
				"name=test2",
			},
			expected: true,
		},
		{
			search: "name=test2",
			slice: []string{
				"test1",
				"test2",
				"name=test2",
			},
			expected: true,
		},
		{
			search: "jack",
			slice: []string{
				"test1",
				"test2",
				"name=test2",
			},
			expected: false,
		},
		{
			search: "",
			slice: []string{
				"",
				"test1",
				"test2",
				"name=test2",
			},
			expected: true,
		},
	}
	for _, val := range list {
		actual := inSlice(val.search, val.slice)
		if actual != val.expected {
			t.Errorf("inSlice return: [%v], expected: [%v]", actual, val.expected)
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
	redisClient := redis.NewClient(options)
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
		t.Errorf("get redis eval result err,actual:%d, confExpected:%d", i, expected)
	}
	result, err = redisClient.Eval(ctx, luaScript, []string{limitKey}, 1, 1).Result()
	if err != nil {
		t.Errorf("eval2 failed, %s", err.Error())
	}
	expected = 1
	i = result.(int64)
	if i != expected {
		t.Errorf("get redis eval result err,actual:%d, confExpected:%d", i, expected)
	}
}
