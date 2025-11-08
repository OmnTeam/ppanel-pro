package responsecode

import (
	"github.com/go-kratos/kratos/v2/errors"
)

// NewSystemNotFoundError 创建系统配置不存在错误
func NewSystemNotFoundError() *errors.Error {
	return errors.New(ErrSystemNotFound, "SYSTEM_NOT_FOUND", CodeMessages[ErrSystemNotFound])
}

// NewDatabaseQueryError 创建数据库查询错误
func NewDatabaseQueryError() error {
	return NewKratosError(ErrDatabaseQuery)
}

// NewDatabaseUpdateError 创建数据库更新错误
func NewDatabaseUpdateError() error {
	return NewKratosError(ErrDatabaseUpdate)
}
