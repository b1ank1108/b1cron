package auth

import (
	"b1cron/internal/config"
	"b1cron/internal/database"
	"b1cron/internal/models"
	"net/http"
	"strings"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

func NewJWTMiddleware(cfg *config.Config) (*jwt.GinJWTMiddleware, error) {
	return jwt.New(&jwt.GinJWTMiddleware{
		Realm:       "B1Cron",
		Key:         []byte(cfg.JWT.Secret),
		Timeout:     time.Hour * time.Duration(cfg.JWT.ExpireHours),
		MaxRefresh:  time.Hour * time.Duration(cfg.JWT.RefreshExpireHours),
		IdentityKey: "username",
		
		PayloadFunc: func(data interface{}) jwt.MapClaims {
			if v, ok := data.(*models.User); ok {
				return jwt.MapClaims{
					"username":              v.Username,
					"user_id":               v.ID,
					"force_password_change": v.ForcePasswordChange,
				}
			}
			return jwt.MapClaims{}
		},
		
		IdentityHandler: func(c *gin.Context) interface{} {
			claims := jwt.ExtractClaims(c)
			return &models.User{
				Username: claims["username"].(string),
			}
		},
		
		Authenticator: func(c *gin.Context) (interface{}, error) {
			var loginReq LoginRequest
			if err := c.ShouldBindJSON(&loginReq); err != nil {
				return "", jwt.ErrMissingLoginValues
			}

			var user models.User
			if err := database.GetDB().Where("username = ?", loginReq.Username).First(&user).Error; err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(loginReq.Password)); err != nil {
				return nil, jwt.ErrFailedAuthentication
			}

			// 将用户数据存储到context中，供LoginResponse使用
			c.Set("user_data", &user)
			
			return &user, nil
		},
		
		Authorizator: func(data interface{}, c *gin.Context) bool {
			return true
		},
		
		Unauthorized: func(c *gin.Context, code int, message string) {
			// 如果是API请求，返回JSON错误
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(code, gin.H{
					"code":    code,
					"message": message,
				})
			} else {
				// 如果是页面请求，重定向到登录页面
				c.Redirect(http.StatusFound, "/login")
			}
		},
		
		TokenLookup: "cookie:" + cfg.JWT.CookieName,
		TokenHeadName: "Bearer",
		
		TimeFunc: time.Now,
		
		SendCookie:     true,
		SecureCookie:   cfg.JWT.Secure,
		CookieHTTPOnly: cfg.JWT.HTTPOnly,
		CookieDomain:   "",
		CookieName:     cfg.JWT.CookieName,
		CookieSameSite: http.SameSiteDefaultMode,
		
		LoginResponse: func(c *gin.Context, code int, token string, expire time.Time) {
			// 获取用户信息以检查是否需要强制修改密码
			user, exists := c.Get("user_data")
			response := gin.H{
				"code":   http.StatusOK,
				"token":  token,
				"expire": expire.Format(time.RFC3339),
			}
			
			if exists && user != nil {
				if u, ok := user.(*models.User); ok && u.ForcePasswordChange {
					response["force_password_change"] = true
					response["message"] = "首次登录，请修改密码"
				}
			}
			
			c.JSON(http.StatusOK, response)
		},
		
		LogoutResponse: func(c *gin.Context, code int) {
			// 检查是否是API请求
			if strings.HasPrefix(c.Request.URL.Path, "/api/") {
				c.JSON(http.StatusOK, gin.H{
					"code":    http.StatusOK,
					"message": "Successfully logged out",
				})
			} else {
				// 如果是页面请求，重定向到登录页面
				c.Redirect(http.StatusFound, "/login")
			}
		},
	})
}