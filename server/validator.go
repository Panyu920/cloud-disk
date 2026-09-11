package server

import (
	"github.com/Panyu920/cloud-disk/utils"
	"github.com/go-playground/validator/v10"
)

// phoneValidator 验证手机号是否符合格式
var phoneValidator = func(fl validator.FieldLevel) bool {
	phone := fl.Field().String()
	return utils.IsValidPhone(phone)
}
