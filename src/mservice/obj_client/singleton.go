package obj_client

import (
	"sync"
)

var (
	once	sync.Once
)

func GetInstance() 