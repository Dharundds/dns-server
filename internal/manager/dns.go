package manager

import "sync"

type ContextManager struct {
	context map[string]string
	mu      sync.RWMutex
}

func NewContextManager() *ContextManager {
	return &ContextManager{}
}

func (m *ContextManager) AddRP(domainName string, port string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.context[domainName] = port
}

func (m *ContextManager) RemoveRP(domainName string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.context, domainName)
}

func (m *ContextManager) GetContext() map[string]string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.context
}

func (m *ContextManager) LoadContext(value map[string]string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.context = value
}
