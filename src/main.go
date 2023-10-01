package main

import (
    "fmt"
    "net/http"
    "github.com/gin-gonic/gin"
    "github.com/jinzhu/gorm"
    _ "github.com/jinzhu/gorm/dialects/sqlite"
)

var (
    db  *gorm.DB
    err error
)

type User struct {
    gorm.Model
    ID       uint   `json:"id"`
    Username string `json:"username" gorm:"not null"`
    Email    string `json:"email" gorm:"unique;not null"`
    Password string `json:"password" gorm:"not null;minlength:6"`
    Photos   []Photo
}

type Photo struct {
    gorm.Model
    ID        uint   `json:"id"`
    Title     string `json:"title"`
    Caption   string `json:"caption"`
    PhotoUrl  string `json:"photoUrl"`
    UserID    uint   `json:"userId"`
    User      User
}

func main() {
    // Inisialisasi database
    db, err = gorm.Open("sqlite3", "test.db")
    if err != nil {
        fmt.Println(err)
    }
    defer db.Close()
    db.AutoMigrate(&User{}, &Photo{})

    r := gin.Default()

    // Endpoint untuk registrasi user baru (sign up)
    r.POST("/users/register", func(c *gin.Context) {
        var newUser User
        if err := c.ShouldBindJSON(&newUser); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        db.Create(&newUser)

        c.JSON(http.StatusOK, gin.H{"message": "User berhasil didaftarkan"})
    })

    // Endpoint untuk login
    r.POST("/users/login", func(c *gin.Context) {
        var user User
        email := c.PostForm("email")
        password := c.PostForm("password")

        if err := db.Where("email = ? AND password = ?", email, password).First(&user).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Login gagal"})
            return
        }

        c.JSON(http.StatusOK, gin.H{"message": "Login berhasil"})
    })

    // Middleware untuk autentikasi user
    r.Use(func(c *gin.Context) {
        // Dapatkan user dari database berdasarkan ID atau token sesi
        // Implementasi autentikasi sesuai kebutuhan aplikasi Anda
        userID := c.Param("userId")

        var user User
        if err := db.First(&user, userID).Error; err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Anda harus login terlebih dahulu"})
            c.Abort()
            return
        }

        c.Set("user", user)
        c.Next()
    })

    // Endpoint untuk mengupdate user
    r.PUT("/users/:userId", func(c *gin.Context) {
        user := c.MustGet("user").(User)
        var updatedUser User

        if err := c.ShouldBindJSON(&updatedUser); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Pastikan hanya pemilik akun yang dapat mengupdate
        if user.ID != updatedUser.ID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk mengupdate pengguna lain"})
            return
        }

        db.Save(&updatedUser)

        c.JSON(http.StatusOK, gin.H{"message": "User berhasil diupdate"})
    })

    // Endpoint untuk menghapus user
    r.DELETE("/users/:userId", func(c *gin.Context) {
        user := c.MustGet("user").(User)
        userID := c.Param("userId")

        // Pastikan hanya pemilik akun yang dapat menghapus
        if fmt.Sprintf("%d", user.ID) != userID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk menghapus pengguna lain"})
            return
        }

        db.Where("id = ?", userID).Delete(&User{})

        c.JSON(http.StatusOK, gin.H{"message": "User berhasil dihapus"})
    })

    // Endpoint untuk menambahkan foto
    r.POST("/photos", func(c *gin.Context) {
        user := c.MustGet("user").(User)
        var newPhoto Photo

        if err := c.ShouldBindJSON(&newPhoto); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Pastikan bahwa foto yang ditambahkan adalah milik pengguna yang sedang login
        if newPhoto.UserID != user.ID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk menambahkan foto untuk pengguna lain"})
            return
        }

        db.Create(&newPhoto)

        c.JSON(http.StatusOK, gin.H{"message": "Foto berhasil ditambahkan"})
    })

    // Endpoint untuk mendapatkan daftar foto
    r.GET("/photos", func(c *gin.Context) {
        var photos []Photo

        db.Find(&photos)

        c.JSON(http.StatusOK, photos)
    })

    // Endpoint untuk mengupdate foto
    r.PUT("/photos/:photoId", func(c *gin.Context) {
        user := c.MustGet("user").(User)
        photoID := c.Param("photoId")
        var updatedPhoto Photo

        if err := c.ShouldBindJSON(&updatedPhoto); err != nil {
            c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
            return
        }

        // Dapatkan foto dari database
        var photo Photo
        if err := db.First(&photo, photoID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Foto tidak ditemukan"})
            return
        }

        // Pastikan hanya pemilik foto yang dapat mengupdate
        if photo.UserID != user.ID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk mengupdate foto milik pengguna lain"})
            return
        }

        // Update foto
        photo.Title = updatedPhoto.Title
        photo.Caption = updatedPhoto.Caption
        photo.PhotoUrl = updatedPhoto.PhotoUrl

        db.Save(&photo)

        c.JSON(http.StatusOK, gin.H{"message": "Foto berhasil diupdate"})
    })

    // Endpoint untuk menghapus foto
    r.DELETE("/photos/:photoId", func(c *gin.Context) {
        user := c.MustGet("user").(User)
        photoID := c.Param("photoId")

        // Dapatkan foto dari database
        var photo Photo
        if err := db.First(&photo, photoID).Error; err != nil {
            c.JSON(http.StatusNotFound, gin.H{"error": "Foto tidak ditemukan"})
            return
        }

        // Pastikan hanya pemilik foto yang dapat menghapus
        if photo.UserID != user.ID {
            c.JSON(http.StatusForbidden, gin.H{"error": "Anda tidak memiliki izin untuk menghapus foto milik pengguna lain"})
            return
        }

        db.Delete(&photo)

        c.JSON(http.StatusOK, gin.H{"message": "Foto berhasil dihapus"})
    })

    r.Run(":8080")
}
