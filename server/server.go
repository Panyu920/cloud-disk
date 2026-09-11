package server

import (
	"log"

	"github.com/Panyu920/cloud-disk/router"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type Server struct {
	ginEnine *gin.Engine
	addr     string
}

func New(addr string) *Server {
	gin.SetMode(gin.ReleaseMode)

	return &Server{
		ginEnine: gin.Default(),
		addr:     addr,
	}
}

// 启动服务
func (s *Server) Start() error {
	// 设置文件上传大小限制为 10MB
	s.ginEnine.MaxMultipartMemory = 10 << 20 // 10MB
	// 注册字段验证器
	registerValidators()

	// 注册路由
	s.rigisterRoutes()
	PrintRoutes()
	return s.ginEnine.Run(s.addr)
}

// 注册路由
func (s *Server) rigisterRoutes() {
	// 首页
	s.ginEnine.GET("/index", router.HandleIndex)
	// 创建用户
	s.ginEnine.POST("/user", router.CreateUserHandler)
	// 登录
	s.ginEnine.POST("/user/login", router.LoginHandler)

	// 验证路由
	authRouter := s.ginEnine.Group("/").Use(authMiddleware())
	// 文件上传
	authRouter.POST("/file", router.HandleUpload)
	// 获取文件元信息
	authRouter.GET("/file/meta", router.GetFileMeta)
	// 处理文件下载请求
	authRouter.GET("/file", router.HandleDownload)
	// 更新文件元信息
	authRouter.PUT("/file/meta", router.UpdateFileMeta)
	// 删除文件
	authRouter.DELETE("/file", router.DeleteFile)

}

func PrintRoutes() {

	log.Println("GET /index  首页")
	log.Println("POST /file  文件上传")
	log.Println("GET /file/meta  获取文件元信息")
	log.Println("GET /file  处理文件下载请求")
	log.Println("PUT /file/meta  更新文件元信息")
	log.Println("DELETE /file  删除文件")
	log.Println("POST /user  创建用户")
	log.Println("POST /user/login  登录")
}

func registerValidators() {
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		v.RegisterValidation("phone", phoneValidator)
	} else {
		log.Fatalf("validator.Engine() is not *validator.Validate")
	}
}
