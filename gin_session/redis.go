package gin_session

import (
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"
	uuid "github.com/satori/go.uuid"

	"github.com/go-redis/redis"
)

// SessionData的构造函数
func NewRedisSessinoData(id string, client *redis.Client) SessionData {
	return &RedisSD{
		ID:     id,
		Data:   make(map[string]interface{}, 8),
		client: client,
	}
}

// Redis版本的SessionData数据结构
type RedisSD struct {
	ID      string
	Data    map[string]interface{}
	rwLock  sync.RWMutex  // 读写锁，用来保护Data数据
	expired int           // 过期时间
	client  *redis.Client // redis连接池
}

func (r *RedisSD) Get(key string) (value interface{}, err error) {
	// 获取锁
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	value, ok := r.Data[key]
	if !ok {
		err = fmt.Errorf("invalid key")
		return
	}
	return
}

func (r *RedisSD) Set(key string, value interface{}) {
	// 获取写锁
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	r.Data[key] = value
}

func (r *RedisSD) Del(key string) {
	// 删除对应key的键值对
	r.rwLock.Lock()
	defer r.rwLock.Unlock()
	delete(r.Data, key)
}

func (r *RedisSD) Save() {
	// 将最新的SessionData保存到redis中
	value, err := json.Marshal(r.Data)
	if err != nil {
		fmt.Printf("redis序列化SessionData失败 err = %v\n", err)
		return
	}
	// 入库
	r.client.Set(r.ID, value, time.Duration(r.expired)*time.Second)
}

func (r *RedisSD)GetID() string {
	return r.ID
}

func NewRedisMgr() Mgr {
	// 返回一个对象实例
	return &RedisMgr{
		Session: make(map[string]SessionData, 1024),
	}
}

type RedisMgr struct {
	Session map[string]SessionData
	rwLock  sync.RWMutex
	client  *redis.Client // redis连接池
}

func (r *RedisMgr) Init(addr string, option ...string) {
	// 初始化连接池
	var (
		passwd string
		db     string
	)
	if len(option) == 1 {
		passwd = option[0]
	} else if len(option) == 2 {
		passwd = option[0]
		db = option[1]
	}

	// 转换db数据类型，输入为string，需要转化为int
	dbValue, err := strconv.Atoi(db)
	if err != nil {
		dbValue = 0
	}
	r.client = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: passwd,
		DB:       dbValue,
	})

	_, err = r.client.Ping().Result()
	if err != nil {
		panic(err)
	}
}

// 加载数据库中的数据
func (r *RedisMgr) LoadFromRedis(sessionID string) (err error) {
	// 1. 根据sessionID从redis中拿到数据
	value, err := r.client.Get(sessionID).Result()
	if err != nil {
		// redis中无sessionID对应的数据
		err = fmt.Errorf("数据库中无sessionID对应数据")
		return
	}
	// 2. 反序列化为r.session
	err = json.Unmarshal([]byte(value), &r.Session)
	if err != nil {
		// 反序列化失败
		fmt.Println("反序列化失败")
		return
	}
	return	
}

// GetSessionData 根据传进来的SessionID找到对应的Session
func (r *RedisMgr) GetSessionData(sessionID string) (sd SessionData, err error) {
	// 从redis中拿数据
	if r.Session == nil {
		err = r.LoadFromRedis(sessionID)
		if err != nil {
			return nil, err
		}
	}

	// 已经从数据库中拿到数据
	r.rwLock.RLock()
	defer r.rwLock.RUnlock()
	sd, ok := r.Session[sessionID]
	if !ok {
		err = fmt.Errorf("无效的sessionID")
		return
	}
	return
}

// 创建一个Session记录
func (r *RedisMgr) CreateSession() (sd SessionData) {
	// 1. 构造一个sessionID
	uuidObj := uuid.NewV4()

	// 2. 创建一个SessionData
	sd = NewRedisSessinoData(uuidObj.String(), r.client)

	// 3. 创建对应关系
	r.Session[sd.GetID()] = sd

	return 
}
