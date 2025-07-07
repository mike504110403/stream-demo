package database

import (
	"fmt"
	"log"
	"stream-demo/backend/database/models"

	"gorm.io/gorm"
)

// AutoMigrate 執行自動遷移
// 確保資料庫結構與模型定義一致
func AutoMigrate(db *gorm.DB) error {
	log.Println("🔄 開始執行資料庫自動遷移...")

	// 定義所有需要遷移的模型
	// 遷移順序很重要：先遷移被依賴的模型，再遷移依賴其他模型的
	modelsToMigrate := []interface{}{
		&models.User{},        // 基礎模型，被其他模型依賴
		&models.Video{},       // 依賴 User
		&models.Live{},        // 依賴 User
		&models.Payment{},     // 依賴 User
		&models.ChatMessage{}, // 依賴 User 和 Live
	}

	// 執行自動遷移
	// GORM 會根據模型的 tag 定義自動創建：
	// - 表結構
	// - 索引（單欄位和複合索引）
	// - 外鍵約束
	// - 預設值
	if err := db.AutoMigrate(modelsToMigrate...); err != nil {
		return fmt.Errorf("❌ 資料庫自動遷移失敗: %v", err)
	}

	log.Println("✅ 資料庫自動遷移完成！")
	return nil
}

// MigrationInfo 顯示遷移信息
func MigrationInfo(db *gorm.DB) error {
	log.Println("📊 資料庫結構信息:")

	tables := []string{"users", "videos", "lives", "payments", "chat_messages"}

	for _, table := range tables {
		var count int64
		if err := db.Table(table).Count(&count).Error; err != nil {
			log.Printf("❌ 查詢表 %s 記錄數失敗: %v", table, err)
			continue
		}
		log.Printf("📋 表 %-15s: %d 條記錄", table, count)
	}

	return nil
}

// DropAllTables 清空所有表（危險操作，僅用於開發環境）
func DropAllTables(db *gorm.DB) error {
	log.Println("⚠️  警告：準備清空所有資料表")

	// 按照外鍵依賴順序刪除
	tables := []string{
		"chat_messages", // 依賴 lives 和 users
		"payments",      // 依賴 users
		"videos",        // 依賴 users
		"lives",         // 依賴 users
		"users",         // 基礎表
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s", table)).Error; err != nil {
			return fmt.Errorf("刪除表 %s 失敗: %v", table, err)
		}
		log.Printf("🗑️  已刪除表: %s", table)
	}

	log.Println("✅ 所有表已清空")
	return nil
}

// ValidateSchema 驗證資料庫結構
func ValidateSchema(db *gorm.DB) error {
	log.Println("🔍 驗證資料庫結構...")

	// 檢查必要的表是否存在
	requiredTables := map[string]interface{}{
		"users":         &models.User{},
		"videos":        &models.Video{},
		"lives":         &models.Live{},
		"payments":      &models.Payment{},
		"chat_messages": &models.ChatMessage{},
	}

	for tableName, model := range requiredTables {
		if !db.Migrator().HasTable(model) {
			return fmt.Errorf("❌ 缺少必要的表: %s", tableName)
		}
		log.Printf("✅ 表 %s 存在", tableName)
	}

	// 檢查關鍵欄位是否存在
	criticalColumns := map[string][]string{
		"users":         {"username", "email", "password"},
		"videos":        {"title", "user_id", "video_url", "status"},
		"lives":         {"title", "user_id", "status", "stream_key"},
		"payments":      {"user_id", "amount", "status", "transaction_id"},
		"chat_messages": {"live_id", "user_id", "content"},
	}

	for tableName, columns := range criticalColumns {
		model := requiredTables[tableName]
		for _, columnName := range columns {
			if !db.Migrator().HasColumn(model, columnName) {
				log.Printf("⚠️  表 %s 缺少欄位: %s", tableName, columnName)
			} else {
				log.Printf("✅ 欄位 %s.%s 存在", tableName, columnName)
			}
		}
	}

	log.Println("✅ 資料庫結構驗證完成")
	return nil
}
