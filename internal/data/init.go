package data

// InitDefaultModules 初始化默认数据模块
func InitDefaultModules() *Manager {
	manager := NewManager()

	// 注册所有默认模块
	manager.RegisterModule(NewCandlesModule())
	manager.RegisterModule(NewNewsModule())
	manager.RegisterModule(NewIndicatorsModule())

	return manager
}
