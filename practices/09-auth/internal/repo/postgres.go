package repo

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Open подключается к PostgreSQL. TranslateError превращает нарушение
// уникального индекса в gorm.ErrDuplicatedKey. «Запись не найдена» при логине
// неизвестного email — штатная ситуация, поэтому в лог она не попадает.
func Open(dsn string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(dsn), &gorm.Config{
		TranslateError: true,
		Logger: logger.New(log.New(os.Stderr, "", log.LstdFlags), logger.Config{
			SlowThreshold:             300 * time.Millisecond,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		}),
	})
}
