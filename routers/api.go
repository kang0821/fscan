package routers

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	v1 "github.com/shadow1ng/fscan/api/v1"
	"github.com/shadow1ng/fscan/model/response"
	"log"
	"net/http"
	"runtime/debug"
)

func InitApiRouter() *gin.Engine {
	var router = gin.Default()
	router.Use(Recover)
	pprof.Register(router)

	router.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "漏扫引擎V0.1版本")
	})

	groupApi := router.Group("/flaw-scan-engine/api")
	{
		home := groupApi.Group("/v1")
		{
			home.POST("/scan", v1.RouterGroup.ScanApi.StartScan)
		}
	}
	return router

}

func Recover(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			//打印错误堆栈信息
			log.Printf("panic: %v\n", r)
			debug.PrintStack()
			//封装通用json返回
			response.FailWithMessage(errorToString(r), c)
			//终止后续接口调用，不加的话recover到异常后，还会继续执行接口里后续代码
			c.Abort()
		}
	}()
	//加载完 defer recover，继续后续接口调用
	c.Next()
}

// recover错误，转string
func errorToString(r interface{}) string {
	switch v := r.(type) {
	case error:
		return v.Error()
	default:
		return r.(string)
	}
}
