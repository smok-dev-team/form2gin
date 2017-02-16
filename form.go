package form2gin

import (
	"github.com/gin-gonic/gin"
	f "github.com/smartwalle/form"
	v "github.com/smartwalle/validator"
	"reflect"
	"sync"
)

const k_FORM_ERROR = "_f2g_bind_form_error"
const k_VALIDATE_ERROR = "_f2g_validate_form_error"
const k_BIND_FORM = "_f2g_form"
const k_DEFAULT_KEY = "Default"

type Getter interface {
	Get(key string) (value interface{}, exists bool)
}

type Handler interface{}
type ErrorHandler func(c *gin.Context, err error)

var bindErrorHandlers map[string]ErrorHandler
var validateErrorHandlers map[string]ErrorHandler
var once sync.Once

func init() {
	once.Do(func() {
		bindErrorHandlers = make(map[string]ErrorHandler)
		validateErrorHandlers = make(map[string]ErrorHandler)
	})
}

func RegisterBindErrorHandler(handler ErrorHandler) {
	RegisterBindErrorHandlerWithKey(k_DEFAULT_KEY, handler)
}

func RegisterBindErrorHandlerWithKey(key string, handler ErrorHandler) {
	bindErrorHandlers[key] = handler
}

func RegisterValidateErrorHandler(handler ErrorHandler) {
	RegisterValidateErrorHandlerWithKey(k_DEFAULT_KEY, handler)
}

func RegisterValidateErrorHandlerWithKey(key string, handler ErrorHandler) {
	validateErrorHandlers[key] = handler
}

// ================================================================================
func BindForm(form interface{}) gin.HandlerFunc {
	return BindFormWithKey(k_DEFAULT_KEY, form)
}

func BindFormWithKey(key string, form interface{}) gin.HandlerFunc {
	var formType = reflect.TypeOf(form)
	if formType.Kind() == reflect.Ptr {
		formType = formType.Elem()
	}

	var bindErrorHandler = bindErrorHandlers[key]

	return func(c *gin.Context) {
		var newForm = reflect.New(formType)
		var err = f.BindWithRequest(c.Request, newForm.Interface())
		if err != nil {
			if bindErrorHandler != nil {
				bindErrorHandler(c, err)
			}
			c.Set(k_FORM_ERROR, err)
		}
		c.Set(k_BIND_FORM, newForm.Interface())
		c.Next()
	}
}

func BindAndValidateForm(form interface{}) gin.HandlerFunc {
	return BindAndValidateFormWithKey(k_DEFAULT_KEY, form)
}

func BindAndValidateFormWithKey(key string, form interface{}) gin.HandlerFunc {
	var formType = reflect.TypeOf(form)
	if formType.Kind() == reflect.Ptr {
		formType = formType.Elem()
	}

	var bindErrorHandler = bindErrorHandlers[key]
	var validateErrorHandler = validateErrorHandlers[key]

	return func(c *gin.Context) {
		var newForm = reflect.New(formType)
		var err = f.BindWithRequest(c.Request, newForm.Interface())
		if err != nil {
			if bindErrorHandler != nil {
				bindErrorHandler(c, err)
			}
			c.Set(k_FORM_ERROR, err)
		}
		var val = v.LazyValidate(newForm.Interface())
		if val.OK() == false {
			if validateErrorHandler != nil {
				validateErrorHandler(c, val.Error())
			}
			c.Set(k_VALIDATE_ERROR, val.Error())
		}
		c.Set(k_BIND_FORM, newForm.Interface())
		c.Next()
	}
}

func HandlerFuncWrapper(h Handler) gin.HandlerFunc {
	return func(c *gin.Context) {
		var funValue = reflect.ValueOf(h)
		if funValue.IsValid() {
			var numIn = reflect.TypeOf(h).NumIn()
			var in = make([]reflect.Value, numIn)
			if numIn > 0 {
				in[0] = reflect.ValueOf(c)
			}
			if numIn > 1 {
				var obj, exist = c.Get(k_BIND_FORM)
				if exist {
					in[1] = reflect.ValueOf(obj)
				}
			}
			funValue.Call(in)
		}
		c.Next()
	}
}
