package apikeys

import "github.com/monms/monms/internal/authbootstrap"

const (
	CollectionName = authbootstrap.CollectionAPIKeys
	CollectionUsers = authbootstrap.CollectionUsers
	prefixLen       = 8
	secretLen       = 32
	tokenPrefix     = "monms_"
)
