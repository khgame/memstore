package cache

import (
	"context"
	"strings"

	"github.com/redis/go-redis/v9"
)

type Prefix string

func (p Prefix) DialHook(next redis.DialHook) redis.DialHook {
	return next
}

func (p Prefix) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		p.AssembleCMD(cmd)
		return next(ctx, cmd)
	}
}

func (p Prefix) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		for _, cmd := range cmds {
			p.AssembleCMD(cmd)
		}
		return next(ctx, cmds)
	}
}

var _ redis.Hook = Prefix("")

func (p Prefix) String() string {
	return string(p)
}

func (p Prefix) ColonStr() string {
	n := len(p)
	if n == 0 {
		return ""
	}
	str := p.String()
	if str[n-1] == ':' {
		return str
	}
	return str + ":"
}

func (p Prefix) MakeKey(key string) string {
	return p.ColonStr() + key
}

func (p Prefix) AssembleCMD(cmd redis.Cmder) redis.Cmder {
	args := cmd.Args()
	lArgs := len(args)
	if lArgs < 2 {
		return cmd
	}

	switch c := strings.ToLower(cmd.Name()); c {
	case "mset", "msetnx":
		for i := 1; i < lArgs; i += 2 {
			args[i] = p.MakeKey(args[i].(string))
		}
	case "mget", "cmget", "exists", "del":
		for i := 1; i < lArgs; i++ {
			args[i] = p.MakeKey(args[i].(string))
		}
	default:
		args[1] = p.MakeKey(args[1].(string))

	}
	return cmd
}
