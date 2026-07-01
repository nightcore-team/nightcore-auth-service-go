package repository

import _ "embed"

//go:embed lua/create_session.lua
var createSessionScript string

//go:embed lua/get_del_session.lua
var getDelSessionScript string

//go:embed lua/delete_all_sessions.lua
var deleteAllSessionsScript string
