package utils // import "admin-console/internal/utils"

import (
	"admin-console/internal/config"
	"fmt"
	"os"
	"sync"

	"github.com/bwmarrin/snowflake"
)

var (
	snowflakeNode *snowflake.Node
	once          sync.Once
)

func initNode() {
	node, err := snowflake.NewNode(config.Opts.NodeId())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	snowflakeNode = node
}

func GenerateID() int64 {
	once.Do(initNode)
	return snowflakeNode.Generate().Int64()
}
