package server

import (
	"github.com/Panyu920/cloud-disk/router"
	"github.com/gin-gonic/gin"
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
	// 注册路由
	s.rigisterRoutes()
	return s.ginEnine.Run(s.addr)
}

// 注册路由
func (s *Server) rigisterRoutes() {
	// 首页
	s.ginEnine.GET("/index", router.HandleIndex)
	// 文件上传
	s.ginEnine.POST("/file", router.HandleUpload)
	// 获取文件元信息
	s.ginEnine.GET("/file/meta", router.GetFileMeta)

}
