package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	iamservice "github.com/rei0721/go-scaffold/internal/modules/iam/service"
	"github.com/rei0721/go-scaffold/internal/ports"
	apperrors "github.com/rei0721/go-scaffold/types/errors"
	"github.com/rei0721/go-scaffold/types/result"
)

const PrincipalKey = "principal"

type Authenticator interface {
	AuthenticateToken(context.Context, string) (iamservice.Principal, error)
}

type Authorizer interface {
	Authorize(context.Context, iamservice.Principal, string, string) (bool, error)
}

func Auth(authenticator Authenticator) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		if authenticator == nil {
			abort(c, http.StatusUnauthorized, apperrors.ErrUnauthorized, "authentication unavailable")
			return
		}
		token := bearerToken(c.GetHeader("Authorization"))
		if token == "" {
			abort(c, http.StatusUnauthorized, apperrors.ErrUnauthorized, "missing bearer token")
			return
		}
		principal, err := authenticator.AuthenticateToken(c.RequestContext(), token)
		if err != nil {
			abort(c, http.StatusUnauthorized, apperrors.ErrUnauthorized, err.Error())
			return
		}
		c.Set(PrincipalKey, principal)
		c.Next()
	}
}

func RequirePermission(authorizer Authorizer, obj, act string, next ports.HTTPHandlerFunc) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		if authorizer == nil {
			abort(c, http.StatusForbidden, apperrors.ErrPermissionDenied, "authorization unavailable")
			return
		}
		principal, ok := GetPrincipal(c)
		if !ok {
			abort(c, http.StatusUnauthorized, apperrors.ErrUnauthorized, "missing principal")
			return
		}
		allowed, err := authorizer.Authorize(c.RequestContext(), principal, obj, act)
		if err != nil || !allowed {
			abort(c, http.StatusForbidden, apperrors.ErrPermissionDenied, "permission denied")
			return
		}
		next(c)
	}
}

func RequireOrgParam(param string, next ports.HTTPHandlerFunc) ports.HTTPHandlerFunc {
	return func(c ports.HTTPContext) {
		principal, ok := GetPrincipal(c)
		if !ok {
			abort(c, http.StatusUnauthorized, apperrors.ErrUnauthorized, "missing principal")
			return
		}
		orgID, err := strconv.ParseInt(c.Param(param), 10, 64)
		if err != nil || orgID != principal.OrgID {
			abort(c, http.StatusForbidden, apperrors.ErrPermissionDenied, "permission denied")
			return
		}
		next(c)
	}
}

func GetPrincipal(c ports.HTTPContext) (iamservice.Principal, bool) {
	value, ok := c.Get(PrincipalKey)
	if !ok {
		return iamservice.Principal{}, false
	}
	principal, ok := value.(iamservice.Principal)
	return principal, ok
}

func bearerToken(header string) string {
	parts := strings.Fields(header)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return parts[1]
}

func abort(c ports.HTTPContext, status int, code int, message string) {
	c.AbortWithStatusJSON(status, result.ErrorWithTrace(code, message, GetTraceID(c)))
}
