package authapi

import (
	"net/http"

	"github.com/teamkeel/keel/proto"
	"github.com/teamkeel/keel/runtime/common"
)

func RevokeHandler(schema *proto.Schema) common.HandlerFunc {
	return func(r *http.Request) common.Response {
		return common.Response{
			Status: http.StatusNotImplemented,
		}
	}
}
