package gin_session

import (
	"fmt"
	"sync"

	uuid "github.com/satori/go.uuid"
)

// 内存版本的session服务
// memory的sessiondata
type MemSD struct {
	ID     string
	Data   map[string]interface{}
	rwLock sync.RWMutex // 读写锁，用来锁定Data
}

// Get 根据key获取值
func (m *MemSD) Get(key string) (value interface{}, err error) {
	// 获取读锁
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	value, ok := m.Data[key]
	if !ok {
		err = fmt.Errorf("invalid key")
		return
	}
	return
}

// Set 设置值
func (m *MemSD) Set(key string, value interface{}) {
	// 获取写锁
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	m.Data[key] = value
}

// Del 删除Key对应的键值对
func (m *MemSD) Del(key string) {
	m.rwLock.Lock()
	defer m.rwLock.Unlock()
	delete(m.Data, key)
}

// Save方法，被动设置，因为要照顾Redis版的接口
func (m *MemSD) Save() {
}

// GetID 为了能够拿到接口的ID数据
func (m *MemSD) GetID() string {
	return m.ID
}

// 全局的Session管理器
type MemoryMgr struct {
	Session map[string]SessionData // 存储所有session的
	rwLock  sync.RWMutex           // 读写锁，用于读多写少的情况，读锁共享，写锁互斥
}

// 内存版本session管理器的构造函数
func NewMemory() Mgr {
	return &MemoryMgr{
		Session: make(map[string]SessionData, 1024),
	}
}

func (m *MemoryMgr) Init(addr string, option ...string) {
}

// GetSessionData 根据传进来的SessionID找到对应的Session
func (m *MemoryMgr) GetSessionData(sessionId string) (sd SessionData, err error) {
	// 获取读锁
	m.rwLock.RLock()
	defer m.rwLock.RUnlock()
	sd, ok := m.Session[sessionId]
	if !ok {
		err = fmt.Errorf("无效的sessionId")
		return
	}
	return
}

func (m *MemoryMgr) CreateSession() (sd SessionData) {
	// 1. 构造一个sessionId
	uuidObj := uuid.NewV4()
	// 2. 创建一个sessionData
	sd = NewMemorySessionData(uuidObj.String())
	// 3. 创建对应的关系
	m.Session[sd.GetID()] = sd
	return
}

// 构造函数，用来构造session
func NewMemorySessionData(id string) SessionData {
	return &MemSD{
		ID: id,
		Data: make(map[string]interface{}, 8),
	}
}



