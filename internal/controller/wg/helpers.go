package wg

import (
	"github.com/HappyLadySauce/NexusPointWG/cmd/app/middleware"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/code"
	"github.com/HappyLadySauce/NexusPointWG/internal/pkg/spec"
	"github.com/HappyLadySauce/errors"
	"github.com/gin-gonic/gin"
)

func requesterRole(c *gin.Context) (string, error) {
	roleAny, ok := c.Get(middleware.UserRoleKey)
	if !ok {
		return "", errors.WithCode(code.ErrTokenInvalid, "missing auth context")
	}
	role, _ := roleAny.(string)
	if role == "" {
		return "", errors.WithCode(code.ErrTokenInvalid, "invalid role in auth context")
	}
	return role, nil
}

func requesterUserID(c *gin.Context) (string, error) {
	idAny, ok := c.Get(middleware.UserIDKey)
	if !ok {
		return "", errors.WithCode(code.ErrTokenInvalid, "missing auth context")
	}
	id, _ := idAny.(string)
	if id == "" {
		return "", errors.WithCode(code.ErrTokenInvalid, "invalid user id in auth context")
	}
	return id, nil
}

func enforce(c *gin.Context, obj string, act spec.Action) error {
	role, err := requesterRole(c)
	if err != nil {
		return err
	}
	allowed, err := spec.Enforce(role, obj, act)
	if err != nil {
		return errors.WithCode(code.ErrUnknown, "authorization engine error")
	}
	if !allowed {
		return errors.WithCode(code.ErrPermissionDenied, "%s", code.Message(code.ErrPermissionDenied))
	}
	return nil
}
