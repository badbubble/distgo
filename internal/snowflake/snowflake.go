package snowflake

import (
	"time"

	sf "github.com/bwmarrin/snowflake"
)

var node *sf.Node

func Init() (err error) {
	var st time.Time
	st, err = time.Parse("2006-01-02", time.Now().Format("2006-01-02"))
	if err != nil {
		return
	}
	sf.Epoch = st.UnixNano() / 1000000
	node, err = sf.NewNode(1)
	return
}

func GenID() string {
	return node.Generate().String()
}
