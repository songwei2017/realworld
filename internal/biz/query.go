package biz

import "gorm.io/gorm"

type ListOptions struct {
	Favorited string
	Tag       string
}

type DbOption func(*gorm.DB) *gorm.DB

func DbOffset(offset int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(int(offset))
	}
}
func DbLimit(size int64) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Limit(int(size))
	}
}

func DbUserId(uid uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Where("user_id", uid)
	}
}

func DbAuthorId(uid uint) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if uid > 0 {
			return db.Where("author_id", uid)
		}
		return db
	}
}

func DbAuthor(username string) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		if len(username) > 0 {
			return db.Where("username", username)
		}
		return db
	}
}
