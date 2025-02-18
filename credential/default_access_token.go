package credential

import (
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/kuro-liang/wechat-go/cache"
	"github.com/kuro-liang/wechat-go/util"
)

const (
	// AccessTokenURL 获取access_token的接口
	accessTokenURL = "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=%s&secret=%s"
	// StableAccessTokenURL 获取稳定版接口调用凭据接口
	stableAccessTokenURL = "https://api.weixin.qq.com/cgi-bin/stable_token?grant_type=client_credential&appid=%s&secret=%s"
	// AccessTokenURL 企业微信获取access_token的接口
	workAccessTokenURL = "https://qyapi.weixin.qq.com/cgi-bin/gettoken?corpid=%s&corpsecret=%s"
	// CacheKeyOfficialAccountPrefix 微信公众号cache key前缀
	CacheKeyOfficialAccountPrefix = "gowechat_officialaccount_"
	// CacheKeyMiniProgramPrefix 小程序cache key前缀
	CacheKeyMiniProgramPrefix = "gowechat_miniprogram_"
	// CacheKeyWorkPrefix 企业微信cache key前缀
	CacheKeyWorkPrefix = "gowechat_work_"
)

// DefaultAccessToken 默认AccessToken 获取
type DefaultAccessToken struct {
	appID                string
	appSecret            string
	cacheKeyPrefix       string
	useStableAccessToken bool // 是否使用稳定的access_token
	cache                cache.Cache
	accessTokenLock      *sync.Mutex
}

// NewDefaultAccessToken new DefaultAccessToken
func NewDefaultAccessToken(appID, appSecret, cacheKeyPrefix string, cache cache.Cache, useStableAccessToken bool) AccessTokenHandle {
	if cache == nil {
		panic("cache is ineed")
	}
	return &DefaultAccessToken{
		appID:                appID,
		appSecret:            appSecret,
		cache:                cache,
		cacheKeyPrefix:       cacheKeyPrefix,
		useStableAccessToken: useStableAccessToken,
		accessTokenLock:      new(sync.Mutex),
	}
}

// StableAccessTokenRequest struct
type StableAccessTokenRequest struct {
	AppID        string `json:"appid"`
	AppSecret    string `json:"secret"`
	GrantType    string `json:"grant_type"`
	ForceRefresh bool   `json:"force_refresh"`
}

// ResAccessToken struct
type ResAccessToken struct {
	util.CommonError

	AccessToken string `json:"access_token"`
	ExpiresIn   int64  `json:"expires_in"`
}

// GetAccessToken 获取access_token,先从cache中获取，没有则从服务端获取
func (ak *DefaultAccessToken) GetAccessToken() (accessToken string, err error) {
	// 先从cache中取
	accessTokenCacheKey := fmt.Sprintf("%s_access_token_%s", ak.cacheKeyPrefix, ak.appID)
	if val := ak.cache.Get(accessTokenCacheKey); val != nil {
		return val.(string), nil
	}

	// 加上lock，是为了防止在并发获取token时，cache刚好失效，导致从微信服务器上获取到不同token
	ak.accessTokenLock.Lock()
	defer ak.accessTokenLock.Unlock()

	// 双检，防止重复从微信服务器获取
	if val := ak.cache.Get(accessTokenCacheKey); val != nil {
		return val.(string), nil
	}

	// cache失效，从微信服务器获取
	var resAccessToken ResAccessToken
	if ak.useStableAccessToken {
		resAccessToken, err = PostTokenFromServer(ak.appID, ak.appSecret)
	} else {
		resAccessToken, err = GetTokenFromServer(fmt.Sprintf(accessTokenURL, ak.appID, ak.appSecret))
	}

	if err != nil {
		return
	}

	expires := resAccessToken.ExpiresIn - 1500
	err = ak.cache.Set(accessTokenCacheKey, resAccessToken.AccessToken, time.Duration(expires)*time.Second)
	if err != nil {
		return
	}
	accessToken = resAccessToken.AccessToken
	return
}

// WorkAccessToken 企业微信AccessToken 获取
type WorkAccessToken struct {
	CorpID          string
	CorpSecret      string
	cacheKeyPrefix  string
	cache           cache.Cache
	accessTokenLock *sync.Mutex
}

// NewWorkAccessToken new WorkAccessToken
func NewWorkAccessToken(corpID, corpSecret, cacheKeyPrefix string, cache cache.Cache) AccessTokenHandle {
	if cache == nil {
		panic("cache the not exist")
	}
	return &WorkAccessToken{
		CorpID:          corpID,
		CorpSecret:      corpSecret,
		cache:           cache,
		cacheKeyPrefix:  cacheKeyPrefix,
		accessTokenLock: new(sync.Mutex),
	}
}

// GetAccessToken 企业微信获取access_token,先从cache中获取，没有则从服务端获取
func (ak *WorkAccessToken) GetAccessToken() (accessToken string, err error) {
	// 加上lock，是为了防止在并发获取token时，cache刚好失效，导致从微信服务器上获取到不同token
	ak.accessTokenLock.Lock()
	defer ak.accessTokenLock.Unlock()
	accessTokenCacheKey := fmt.Sprintf("%s_access_token_%s", ak.cacheKeyPrefix, ak.CorpID)
	val := ak.cache.Get(accessTokenCacheKey)
	if val != nil {
		accessToken = val.(string)
		return
	}

	// cache失效，从微信服务器获取
	var resAccessToken ResAccessToken
	resAccessToken, err = GetTokenFromServer(fmt.Sprintf(workAccessTokenURL, ak.CorpID, ak.CorpSecret))
	if err != nil {
		return
	}

	expires := resAccessToken.ExpiresIn - 1500
	err = ak.cache.Set(accessTokenCacheKey, resAccessToken.AccessToken, time.Duration(expires)*time.Second)
	if err != nil {
		return
	}
	accessToken = resAccessToken.AccessToken
	return
}

// GetTokenFromServer 强制从微信服务器获取token
func GetTokenFromServer(url string) (resAccessToken ResAccessToken, err error) {
	var body []byte
	body, err = util.HTTPGet(url)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resAccessToken)
	if err != nil {
		return
	}
	if resAccessToken.ErrCode != 0 {
		err = fmt.Errorf("get access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrCode, resAccessToken.ErrMsg)
		return
	}
	return
}

// PostTokenFromServer 强制从微信服务器获取token
func PostTokenFromServer(appId string, appSecret string) (resAccessToken ResAccessToken, err error) {
	var reqAccessTokenReq = StableAccessTokenRequest{
		AppID:        appId,
		AppSecret:    appSecret,
		GrantType:    "client_credential",
		ForceRefresh: true,
	}
	var body []byte
	body, err = util.PostJSON(stableAccessTokenURL, reqAccessTokenReq)
	if err != nil {
		return
	}
	err = json.Unmarshal(body, &resAccessToken)
	if err != nil {
		return
	}
	if resAccessToken.ErrCode != 0 {
		err = fmt.Errorf("get stable_access_token error : errcode=%v , errormsg=%v", resAccessToken.ErrCode, resAccessToken.ErrMsg)
		return
	}
	return
}
