package gorm

var migrateModels []interface{}

// RegisterMigrateModel 注册用于数据库迁移的数据库模型
func RegisterMigrateModel(model interface{}) {
	migrateModels = append(migrateModels, &model)
}

// getMigrateModels 获取用于数据库迁移的数据库模型
func getMigrateModels() []interface{} {
	return migrateModels
}
