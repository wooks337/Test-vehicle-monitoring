package sessionstore

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-redis/redis"
	"time"
)

type ServiceSessionStore struct {
	client *redis.Client
}

type ServiceSessionInfo struct {
	LoginTime *time.Time `json:"LoginTime"`
	ExpireOn  *time.Time `json:"ExpiredOn"`
	UserID    string     `json:"UserID"`
	Email     string     `json:"Email"`
	Name      string     `json:"Name"`
	Locale    string     `json:"Locale"`
	XDeviceID string     `json:"XDeviceID"`
}

func DecodeServiceSessionInfo(sessionInfoStr string) (sessionInfo *ServiceSessionInfo, err error) {
	jsonBytes, _ := base64.StdEncoding.DecodeString(sessionInfoStr)

	// JSON 디코딩
	sessionInfo = &ServiceSessionInfo{}
	err = json.Unmarshal(jsonBytes, sessionInfo)
	if err != nil {
		if _, ok := err.(*time.ParseError); ok {
			fmt.Println("old session decode")

			var sessionInfoMap map[string]interface{}
			json.Unmarshal(jsonBytes, &sessionInfoMap)

			sessionInfo.UserID = sessionInfoMap["UserID"].(string)

		}
	}

	return
}

func EncodeServiceSessionInfo(sessionInfo *ServiceSessionInfo) (sessionInfoStr string, err error) {
	// JSON 인코딩
	jsonBytes, err := json.Marshal(sessionInfo)

	sessionInfoStr = base64.StdEncoding.EncodeToString(jsonBytes)

	return
}

func (ss *ServiceSessionStore) Open(host string, pw string, db int, poolSize int) error {
	ss.client = redis.NewClient(&redis.Options{
		Addr:     host,
		Password: pw,
		DB:       db,
		PoolSize: poolSize,
	})

	_, err := ss.client.Ping().Result()

	return err
}

func (ss *ServiceSessionStore) Close() {
	if ss.client != nil {
		ss.client.Close()
	}
}

func (ss *ServiceSessionStore) GetSessionInfo(userID string, accessToken string) (sessionInfo *ServiceSessionInfo, err error) {
	if ss.client == nil {
		return nil, errors.New("service session store not initialized")
	}

	if accessToken != "" {
		key := userID // + "." + accessToken
		val, err := ss.client.Get(key).Result()

		if err == nil {
			sessionInfo, err = DecodeServiceSessionInfo(val)
		}
	}

	return
}

func (ss *ServiceSessionStore) GetScanSession(userID string) (keys []string, err error) {
	if ss.client == nil {
		return nil, errors.New("service session store not initialized")
	}

	if userID != "" {
		keys, _, err = ss.client.Scan(0, userID+"*", 10000).Result()
	}

	return nil, nil
}

func (ss *ServiceSessionStore) SetSessionInfo(userID string, accessToken string, sessionInfo *ServiceSessionInfo) (err error) {
	if ss.client == nil {
		return errors.New("service session store not initialized")
	}

	sessionInfoStr, err := EncodeServiceSessionInfo(sessionInfo)

	if err == nil {
		expireDuration := 0 * time.Second

		if sessionInfo.ExpireOn != nil {
			diff := sessionInfo.ExpireOn.Sub(time.Now())

			if diff > 0 {
				expireDuration = diff
			}

		} else { // 미설정시 디폴트로
			expireDuration = 60 * 60 * 24 * time.Second // 1 day
		}

		// 이미 만료된 것은 세팅할 필요가 없음 삭제하거나 경고 남길지를 더 검토가 필요함
		if expireDuration > 0 {
			key := userID //+ ":" + accessToken
			err = ss.client.Set(key, sessionInfoStr, expireDuration).Err()
		}

	}

	return err
}

func (ss *ServiceSessionStore) DelSessionInfo(userID, accessToken, key string) error {
	if ss.client == nil {
		return errors.New("service session store not initialized")
	}

	if key == "" {
		key = userID + ":" + accessToken
	}
	return ss.client.Del(key).Err()
}

func Encode(info ServiceSessionInfo) string {
	// Json 인코딩
	jsonBytes, err := json.Marshal(info)
	if err != nil {
		panic(err)
	}

	enc := base64.StdEncoding.EncodeToString(jsonBytes)
	return enc
}
