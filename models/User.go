package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type User struct {
	gorm.Model
	Username      string `gorm:"unique"`
	Email         string `gorm:"unique"`
	Password      string
	FullName      string    // Họ và tên đầy đủ của học sinh
	DateOfBirth   time.Time // Ngày sinh của học sinh
	Gender        string    // Giới tính của học sinh
	Address       string    // Địa chỉ của học sinh
	ParentName    string    // Tên phụ huynh của học sinh
	ParentContact string    // Số điện thoại liên lạc của phụ huynh
	Class         string    // Lớp học của học sinh
	Section       string    // Khối học của học sinh
	RollNumber    string    // Số thứ tự (Roll Number) của học sinh trong lớp
}
