package main

import (
	"b1cron/internal/auth"
	"b1cron/internal/config"
	"b1cron/internal/database"
	"b1cron/internal/handler"
	"b1cron/internal/models"
	"b1cron/internal/scheduler"
	"b1cron/internal/service"
	"context"
	"flag"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	jwt "github.com/appleboy/gin-jwt/v2"
)

func main() {
	// 解析命令行参数
	configPath := flag.String("config", "config.yaml", "配置文件路径")
	flag.Parse()

	// 加载配置
	cfg, err := config.LoadConfig(*configPath)
	if err != nil {
		log.Fatal("Failed to load configuration:", err)
	}

	// 设置Gin运行模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化数据库
	if err := database.InitDatabase(cfg.Database.Path); err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	// 创建默认用户
	if err := createDefaultUser(cfg); err != nil {
		log.Fatal("Failed to create default user:", err)
	}

	// 创建调度服务
	schedulerService, err := scheduler.NewSchedulerService()
	if err != nil {
		log.Fatal("Failed to create scheduler service:", err)
	}

	if err := schedulerService.Start(); err != nil {
		log.Fatal("Failed to start scheduler service:", err)
	}

	// 创建服务和处理器
	taskService := service.NewTaskService(schedulerService, cfg)
	taskHandler := handler.NewTaskHandler(taskService)

	// 创建JWT中间件
	jwtMiddleware, err := auth.NewJWTMiddleware(cfg)
	if err != nil {
		log.Fatal("Failed to initialize JWT middleware:", err)
	}

	if err := jwtMiddleware.MiddlewareInit(); err != nil {
		log.Fatal("Failed to init JWT middleware:", err)
	}

	// 设置路由
	router := setupRouter(jwtMiddleware, taskHandler, cfg)

	// 创建HTTP服务器
	srv := &http.Server{
		Addr:    cfg.GetAddr(),
		Handler: router,
	}

	// 启动服务器
	go func() {
		log.Printf("Server starting on %s (mode: %s)", cfg.GetAddr(), cfg.Server.Mode)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal("Failed to start server:", err)
		}
	}()

	// 等待中断信号优雅关闭服务器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 设置关闭超时
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 停止调度器
	if err := schedulerService.Stop(); err != nil {
		log.Printf("Error stopping scheduler: %v", err)
	}

	// 关闭HTTP服务器
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exited")
}

func setupRouter(jwtMiddleware *jwt.GinJWTMiddleware, taskHandler *handler.TaskHandler, cfg *config.Config) *gin.Engine {
	router := gin.Default()

	// 添加模板函数
	router.SetFuncMap(template.FuncMap{
		"toFloat64": func(i int64) float64 {
			return float64(i)
		},
		"div": func(a, b float64) float64 {
			if b == 0 {
				return 0
			}
			return a / b
		},
		"js": func(s interface{}) string {
			if s == nil {
				return "null"
			}
			str := fmt.Sprintf("%v", s)
			// 转义JavaScript字符串
			str = strings.ReplaceAll(str, "\\", "\\\\")
			str = strings.ReplaceAll(str, "'", "\\'")
			str = strings.ReplaceAll(str, "\"", "\\\"")
			str = strings.ReplaceAll(str, "\n", "\\n")
			str = strings.ReplaceAll(str, "\r", "\\r")
			str = strings.ReplaceAll(str, "\t", "\\t")
			return "'" + str + "'"
		},
		"jsRaw": func(s interface{}) string {
			if s == nil {
				return "null"
			}
			str := fmt.Sprintf("%v", s)
			// 只转义必要的字符，不添加外层引号
			str = strings.ReplaceAll(str, "\\", "\\\\")
			str = strings.ReplaceAll(str, "'", "\\'")
			str = strings.ReplaceAll(str, "\"", "\\\"")
			str = strings.ReplaceAll(str, "\n", "\\n")
			str = strings.ReplaceAll(str, "\r", "\\r")
			str = strings.ReplaceAll(str, "\t", "\\t")
			return str
		},
	})

	router.LoadHTMLGlob("templates/*")
	
	// 静态文件服务
	router.Static("/static", "./static")

	// Public routes (no authentication required)
	router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/login")
	})
	
	router.GET("/login", func(c *gin.Context) {
		c.HTML(http.StatusOK, "login.html", nil)
	})

	router.POST("/login", jwtMiddleware.LoginHandler)

	// Protected routes (authentication required)
	auth := router.Group("/")
	auth.Use(jwtMiddleware.MiddlewareFunc())
	{
		auth.GET("/dashboard", taskHandler.ShowDashboard)
		// 添加其他需要保护的页面路由
		auth.GET("/settings", taskHandler.ShowDashboard) // 临时使用dashboard处理器
		auth.GET("/profile", taskHandler.ShowDashboard)   // 临时使用dashboard处理器
		
		// 退出登录
		auth.POST("/logout", jwtMiddleware.LogoutHandler)
		auth.GET("/logout", func(c *gin.Context) {
			// 清除cookie
			c.SetCookie(cfg.JWT.CookieName, "", -1, "/", "", cfg.JWT.Secure, cfg.JWT.HTTPOnly)
			c.Redirect(http.StatusFound, "/login")
		})
	}

	api := router.Group("/api")
	api.Use(jwtMiddleware.MiddlewareFunc())
	{
		api.GET("/tasks", taskHandler.GetTasks)
		api.POST("/tasks", taskHandler.CreateTask)
		api.GET("/tasks/:id", taskHandler.GetTask)
		api.PUT("/tasks/:id", taskHandler.UpdateTask)
		api.DELETE("/tasks/:id", taskHandler.DeleteTask)
		api.PATCH("/tasks/:id/toggle", taskHandler.ToggleTask)
		api.GET("/tasks/:id/executions", taskHandler.GetTaskExecutions)
		api.GET("/executions/recent", taskHandler.GetRecentExecutions)
		api.POST("/change-password", taskHandler.ChangePassword)
	}

	return router
}

func createDefaultUser(cfg *config.Config) error {
	var count int64
	if err := database.GetDB().Model(&models.User{}).Count(&count).Error; err != nil {
		return err
	}

	if count > 0 {
		return nil
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(cfg.DefaultUser.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username:            cfg.DefaultUser.Username,
		PasswordHash:        string(hashedPassword),
		ForcePasswordChange: true, // 默认用户首次登录必须修改密码
	}

	if err := database.GetDB().Create(user).Error; err != nil {
		return err
	}

	log.Printf("Default user created: %s/%s", cfg.DefaultUser.Username, cfg.DefaultUser.Password)
	return nil
}