package service

import (
	"sync"

	"github.com/google/wire"
)

var StateServiceSet = wire.NewSet(
	ProvideStateService,
)

// StateChangeListener 是状态变化的监听器函数类型
type StateChangeListener func(key string, oldValue, newValue interface{})

// StateService 管理应用程序的状态
type StateService struct {
	mu        sync.RWMutex
	states    map[string]interface{}
	listeners []StateChangeListener
}

// NewStateService 创建并返回一个新的 StateService 实例
func ProvideStateService() *StateService {
	return &StateService{
		states:    make(map[string]interface{}),
		listeners: make([]StateChangeListener, 0),
	}
}

// GetState 获取指定键的状态值
func (s *StateService) GetState(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.states[key]
	return value, exists
}

// SetState 设置指定键的状态值
func (s *StateService) SetState(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	oldValue := s.states[key]
	s.states[key] = value
	s.notifyListeners(key, oldValue, value)
}

// AddListener 添加一个状态变化监听器
func (s *StateService) AddListener(listener StateChangeListener) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.listeners = append(s.listeners, listener)
}

// notifyListeners 通知所有监听器状态变化
func (s *StateService) notifyListeners(key string, oldValue, newValue interface{}) {
	for _, listener := range s.listeners {
		go listener(key, oldValue, newValue)
	}
}

// GetBoolState 获取布尔类型的状态值
func (s *StateService) GetBoolState(key string) bool {
	value, exists := s.GetState(key)
	if !exists {
		return false
	}
	boolValue, ok := value.(bool)
	return ok && boolValue
}

func (s *StateService) GetIntState(key string) int {
	value, exists := s.GetState(key)
	if !exists {
		return 0
	}
	intValue, ok := value.(int)
	if !ok {
		return 0
	}
	return intValue
}

// GetStringState 获取字符串类型的状态值
func (s *StateService) GetStringState(key string) string {
	value, exists := s.GetState(key)
	if !exists {
		return ""
	}
	strValue, ok := value.(string)
	if !ok {
		return ""
	}
	return strValue
}

// 可以根据需要添加更多类型特定的获取方法
