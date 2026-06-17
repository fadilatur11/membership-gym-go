package response

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"

	apperror "membership-gym/pkg/errors"
	"membership-gym/pkg/pagination"
)

type Envelope struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
	Errors  any    `json:"errors,omitempty"`
	Meta    any    `json:"meta,omitempty"`
}

func OK(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusOK, Envelope{Success: true, Message: message, Data: data})
}

func Created(ctx *gin.Context, message string, data any) {
	ctx.JSON(http.StatusCreated, Envelope{Success: true, Message: message, Data: data})
}

func NoContent(ctx *gin.Context) {
	ctx.Status(http.StatusNoContent)
}

func Paginated(ctx *gin.Context, data any, meta pagination.Meta) {
	ctx.JSON(http.StatusOK, Envelope{Success: true, Data: data, Meta: meta})
}

func Error(ctx *gin.Context, err error) {
	var appErr *apperror.AppError
	if errors.As(err, &appErr) {
		ctx.JSON(appErr.Code, Envelope{Success: false, Message: appErr.Message, Errors: appErr.Errs})
		return
	}
	ctx.JSON(http.StatusInternalServerError, Envelope{Success: false, Message: "Terjadi kesalahan server", Errors: nil})
}

func ValidationError(ctx *gin.Context, err error) {
	errs := map[string]string{}
	var validationErrs validator.ValidationErrors
	if errors.As(err, &validationErrs) {
		for _, fieldErr := range validationErrs {
			errs[fieldErr.Field()] = fieldErr.Tag()
		}
	} else {
		errs["request"] = err.Error()
	}
	ctx.JSON(http.StatusBadRequest, Envelope{Success: false, Message: "Validasi gagal", Errors: errs})
}
