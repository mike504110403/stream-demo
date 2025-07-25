package database

import (
	"fmt"
	"stream-demo/backend/config"
	"stream-demo/backend/database/models"
	"stream-demo/backend/utils"

	"gorm.io/gorm"
)

// MigratePostgreSQL 執行PostgreSQL相關的資料庫遷移
func MigratePostgreSQL(conf *config.Config) error {
	// 獲取主資料庫連接
	db := conf.DB["master"]

	utils.LogInfo("開始PostgreSQL資料庫遷移...")

	// 自動遷移所有模型
	err := db.AutoMigrate(
		&models.User{},
		&models.Video{},
		&models.VideoQuality{}, // 新增 VideoQuality 模型
		&models.Payment{},
		&models.Live{},
		&models.ChatMessage{},
	)

	if err != nil {
		utils.LogError("資料庫遷移失敗:", err)
		return err
	}

	// 創建PostgreSQL特定的索引和優化
	if err := createPostgreSQLIndexes(db); err != nil {
		utils.LogError("創建索引失敗:", err)
		return err
	}

	// 創建PostgreSQL擴展
	if err := createPostgreSQLExtensions(db); err != nil {
		utils.LogError("創建擴展失敗:", err)
		return err
	}

	utils.LogInfo("PostgreSQL資料庫遷移完成")
	return nil
}

// createPostgreSQLIndexes 創建PostgreSQL特定的索引
func createPostgreSQLIndexes(db *gorm.DB) error {
	// 用戶表索引
	indexes := []string{
		"CREATE INDEX IF NOT EXISTS idx_users_username ON users (username)",
		"CREATE INDEX IF NOT EXISTS idx_users_email ON users (email)",
		"CREATE INDEX IF NOT EXISTS idx_users_created_at ON users (created_at DESC)",

		// 影片表索引
		"CREATE INDEX IF NOT EXISTS idx_videos_user_id ON videos (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_videos_status ON videos (status)",
		"CREATE INDEX IF NOT EXISTS idx_videos_created_at ON videos (created_at DESC)",
		"CREATE INDEX IF NOT EXISTS idx_videos_views ON videos (views DESC)",
		"CREATE INDEX IF NOT EXISTS idx_videos_likes ON videos (likes DESC)",
		// PostgreSQL全文搜索索引
		"CREATE INDEX IF NOT EXISTS idx_videos_search ON videos USING gin(to_tsvector('english', title || ' ' || description))",

		// 影片品質表索引
		"CREATE INDEX IF NOT EXISTS idx_video_qualities_video_id ON video_qualities (video_id)",
		"CREATE INDEX IF NOT EXISTS idx_video_qualities_quality ON video_qualities (quality)",

		// 直播表索引
		"CREATE INDEX IF NOT EXISTS idx_lives_user_id ON lives (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_lives_status ON lives (status)",
		"CREATE INDEX IF NOT EXISTS idx_lives_start_time ON lives (start_time DESC)",
		"CREATE INDEX IF NOT EXISTS idx_lives_stream_key ON lives (stream_key)",

		// 支付表索引
		"CREATE INDEX IF NOT EXISTS idx_payments_user_id ON payments (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_status ON payments (status)",
		"CREATE INDEX IF NOT EXISTS idx_payments_transaction_id ON payments (transaction_id)",
		"CREATE INDEX IF NOT EXISTS idx_payments_created_at ON payments (created_at DESC)",

		// 聊天訊息表索引
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_live_id ON chat_messages (live_id)",
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_user_id ON chat_messages (user_id)",
		"CREATE INDEX IF NOT EXISTS idx_chat_messages_created_at ON chat_messages (created_at DESC)",
	}

	for _, index := range indexes {
		if err := db.Exec(index).Error; err != nil {
			utils.LogError("創建索引失敗:", index, err)
			return err
		}
	}

	utils.LogInfo("PostgreSQL索引創建完成")
	return nil
}

// createPostgreSQLExtensions 創建PostgreSQL擴展
func createPostgreSQLExtensions(db *gorm.DB) error {
	extensions := []string{
		"CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"", // UUID支援
		"CREATE EXTENSION IF NOT EXISTS \"pg_trgm\"",   // 三元組相似搜索
		"CREATE EXTENSION IF NOT EXISTS \"btree_gin\"", // GIN索引支援
	}

	for _, ext := range extensions {
		if err := db.Exec(ext).Error; err != nil {
			utils.LogWarn("創建擴展失敗（可能需要超級用戶權限）:", ext, err)
			// 擴展創建失敗不應該中斷遷移，只記錄警告
		}
	}

	utils.LogInfo("PostgreSQL擴展初始化完成")
	return nil
}

// CreateTriggersAndFunctions 創建PostgreSQL觸發器和函數
func CreateTriggersAndFunctions(db *gorm.DB) error {
	// 創建自動更新updated_at的函數
	updateFunction := `
		CREATE OR REPLACE FUNCTION update_updated_at_column()
		RETURNS TRIGGER AS $$
		BEGIN
			NEW.updated_at = CURRENT_TIMESTAMP;
			RETURN NEW;
		END;
		$$ language 'plpgsql';
	`

	if err := db.Exec(updateFunction).Error; err != nil {
		utils.LogError("創建更新函數失敗:", err)
		return err
	}

	// 為每個表創建觸發器
	tables := []string{"users", "videos", "video_qualities", "lives", "payments", "chat_messages"}

	for _, table := range tables {
		trigger := fmt.Sprintf(`
			DROP TRIGGER IF EXISTS update_%s_updated_at ON %s;
			CREATE TRIGGER update_%s_updated_at
				BEFORE UPDATE ON %s
				FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
		`, table, table, table, table)

		if err := db.Exec(trigger).Error; err != nil {
			utils.LogError("創建觸發器失敗:", table, err)
			return err
		}
	}

	utils.LogInfo("PostgreSQL觸發器和函數創建完成")
	return nil
}
