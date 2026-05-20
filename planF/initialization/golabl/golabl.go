package golabl

import (
	"context"
	_type "planA/type"

	"github.com/gorilla/mux"
)

var (
	Ctx    context.Context
	Config _type.Config
	Router = mux.NewRouter()
)
