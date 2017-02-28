package form2gin

import (
	"github.com/gin-gonic/gin"
	f "github.com/smartwalle/form"
	"reflect"
	"sync"
)

const k_FORM_ERROR = "_f2g_bind_form_error"
const k_BIND_FORM = "_f2g_form"
const k_DEFAULT_KEY = "_f2g_default"

type Handler interface{}
type ErrorHandler func(c *gin.Context, err error)

var bindErrorHandlers map[string]ErrorHandler
var once sync.Once

func init() {
	once.Do(func() {
		bindErrorHandlers = make(map[string]ErrorHandler)
	})
}

func RegisterBindErrorHandler(handler ErrorHandler) {
	RegisterBindErrorHandlerWithKey(k_DEFAULT_KEY, handler)
}

func RegisterBindErrorHandlerWithKey(key string, handler ErrorHandler) {
	bindErrorHandlers[key] = handler
}

// ================================================================================
func MidBindForm(form interface{}) gin.HandlerFunc {
	return MidBindFormWithKey(k_DEFAULT_KEY, form)
}

func MidBindFormWithKey(key string, form interface{}) gin.HandlerFunc {
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
			return
		}
		c.Set(k_BIND_FORM, newForm.Interface())
		c.Next()
	}
}

func HandlerWrapper(h Handler) gin.HandlerFunc {
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

// ================================================================================

func BindForm(c *gin.Context, form interface{}) bool {
	return BindFormWithKey(k_DEFAULT_KEY, c, form)
}

func BindFormWithKey(key string, c *gin.Context, form interface{}) bool {
	var err = f.BindWithRequest(c.Request, form)
	if err != nil {
		var bindErrorHandler = bindErrorHandlers[key]
		if bindErrorHandler != nil {
			bindErrorHandler(c, err)
		}
		return false
	}
	return true
}