package client

type Action string

const (
	PullAction    Action = "pull"
	PushAction    Action = "push"
	CatalogAction Action = "*"
	DeleteAction  Action = "delete"
)
