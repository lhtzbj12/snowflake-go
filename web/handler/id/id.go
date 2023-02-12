package id

import (
	"net/http"
	"sfgo/common/convutil"
	"sfgo/common/netutil"
	"sfgo/common/tools"
	"sfgo/common/valiutil"
	"sfgo/core"
	"sfgo/core/snowflake"
	"sfgo/web/vo"
	"strconv"

	"github.com/gin-gonic/gin"
)

const maxCount = 10000

var idGenerator core.IdGenerator

var port = tools.GetEnv("SERVER_PORT", "8074")

var appName = tools.GetEnv("DISCOVERY_MICROSRV_NAME", "id-generator")

func init() {
	idGenerator = snowflake.NewIdGenerator(netutil.GetFirstNonLoopbackIP(), port, appName)
	idGenerator.Init()
}

// GetOne 获取1个id
func GetOne(ctx *gin.Context) {
	id, _ := idGenerator.GetId()
	resp := vo.SuccessRespBase(strconv.FormatInt(id, 10))
	ctx.JSON(http.StatusOK, resp)
}

// GetBatch 获取 count 个id
func GetBatch(ctx *gin.Context) {
	paramCount := ctx.DefaultQuery("count", "1")
	if !valiutil.IsNumber(paramCount) {
		ctx.JSON(http.StatusOK, vo.ParamInvalidRespBase("count"))
		return
	}
	count, _ := strconv.Atoi(paramCount)
	if count > maxCount {
		count = maxCount
	}
	ids, err := idGenerator.GetIds(count)
	if err != nil {
		ctx.JSON(http.StatusOK, vo.BusinessFailedRespBase(err.Error()))
		return
	}
	idsStr := convutil.SliceInt2Str(ids)
	resp := vo.SuccessRespBase(idsStr)
	ctx.JSON(http.StatusOK, resp)
}
