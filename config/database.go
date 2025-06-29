package config

import (
	"fmt"
	"log"
	"ticketing-system/entity"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

func ConnectDatabase() {
	var err error
	
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		AppConfig.Database.Username,
		AppConfig.Database.Password,
		AppConfig.Database.Host,
		AppConfig.Database.Port,
		AppConfig.Database.DBName,
	)

	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	log.Println("Database connected successfully")
}

func AutoMigrate() {
	err := DB.AutoMigrate(
		&entity.User{},
		&entity.Event{},
		&entity.Ticket{},
	)

	if err != nil {
		log.Fatal("Failed to migrate database:", err)
	}

	log.Println("Database migration completed")

	// Seed admin user
	seedAdminUser()
}

func seedAdminUser() {
	var adminUser entity.User
	result := DB.Where("email = ?", AppConfig.Admin.Email).First(&adminUser)

	if result.Error == gorm.ErrRecordNotFound {
		// Hash password with the correct hash mentioned in memory
		hashedPassword := "$2a$12$yOsy6BlB90vLnaP6cGIfwObcwe6us33Ayn4bQMda8znFBBpgSV366"
		
		// Verify the hash is correct for the password
		if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(AppConfig.Admin.Password)); err != nil {
			// If the hash doesn't match, create a new one
			hashedBytes, err := bcrypt.GenerateFromPassword([]byte(AppConfig.Admin.Password), 12)
			if err != nil {
				log.Fatal("Failed to hash admin password:", err)
			}
			hashedPassword = string(hashedBytes)
		}

		admin := entity.User{
			Email:    AppConfig.Admin.Email,
			Password: hashedPassword,
			Name:     "System Administrator",
			Role:     entity.RoleAdmin,
			IsActive: true,
		}

		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("Failed to create admin user: %v", err)
		} else {
			log.Println("Admin user created successfully")
		}
	} else {
		log.Println("Admin user already exists")
	}
} 