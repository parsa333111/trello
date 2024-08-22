package custom_errors

import "errors"

var ErrAccessDenied = errors.New("access denied")
var ErrDatabaseFailure = errors.New("request failed")
var ErrDuplicateUsername = errors.New("username has already been taken")
var ErrDuplicateEmail = errors.New("account with this email already exists")
var ErrDuplicateWorspaceName = errors.New("workspace name has already been taken")
var ErrDuplicateTaskTitle = errors.New("task with this title already exists")
var ErrDuplicateSubtaskTitle = errors.New("subtask with this title already exists")

var ErrCreateWorkspaceTableFailed = errors.New("failed to create workspace table")
var ErrCreateTaskTableFailed = errors.New("failed to create task table")
var ErrCreateSubtaskTableFailed = errors.New("failed to create subtask table")
var ErrCreateUserTableFailed = errors.New("failed to create user table")
var ErrCreateUserworkspaceroleTableFailed = errors.New("failed to create userworkspacerole table")
var ErrCreateMessageTableFailed = errors.New("failed to create message table")
var ErrCreateWatchTableFailed = errors.New("failed to create watch table")

var ErrTokenFailure = errors.New("failed to generate token")
var ErrInvalidArguments = errors.New("invalid arguments")

var ErrPictureOpenFailure = errors.New("failed to open file")
var ErrPictureLoadFailure = errors.New("failed to load file")
var ErrPictureDecodeFailure = errors.New("failed to decode file")
