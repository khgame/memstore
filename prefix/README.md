# Prefix

Prefix 是一个用于 Redis 键前缀管理的 Go 模块。它提供了一种简单而强大的方式来为 Redis 键添加前缀,有助于组织和管理大型 Redis 数据库中的键。

## 主要特性

- 为 Redis 键添加前缀
- 实现了 `redis.Hook` 接口,可以无缝集成到现有的 Redis 客户端中
- 支持大多数常见的 Redis 命令
- 提供了简单的 API 来创建和管理前缀

## 使用方法

### 基本用法

```go
import "github.com/your-username/your-repo-name/prefix"

// 创建一个前缀
p := prefix.Prefix("myapp:") // or prefix.Prefix("myapp")

// 使用前缀创建键
key := p.MakeKey("user:123")
// 结果: "myapp:user:123"
```

### 与 Redis 客户端集成

```go
import (
    "github.com/redis/go-redis/v9"
    "github.com/your-username/your-repo-name/prefix"
)

// 创建 Redis 客户端
client := redis.NewClient(&redis.Options{
    Addr: "localhost:6379",
})

// 创建前缀并添加到 Redis 客户端
p := prefix.Prefix("myapp:")
client.AddHook(p)

// 现在,所有的 Redis 操作都会自动添加前缀
client.Set(ctx, "user:123", "John Doe", 0)
// 实际上设置的键是 "myapp:user:123"
```

## API 参考

### `type Prefix string`

Prefix 是一个字符串类型,表示要添加到 Redis 键前面的前缀。

#### 方法

- `func (p Prefix) String() string`: 返回前缀的字符串表示。
- `func (p Prefix) ColonStr() string`: 返回带冒号的前缀字符串。
- `func (p Prefix) MakeKey(key string) string`: 使用前缀创建一个新的键。

### Redis Hook 方法

Prefix 类型实现了 `redis.Hook` 接口,包括以下方法:

- `func (p Prefix) DialHook(next redis.DialHook) redis.DialHook`
- `func (p Prefix) ProcessHook(next redis.ProcessHook) redis.ProcessHook`
- `func (p Prefix) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook`

这些方法允许 Prefix 无缝集成到 Redis 客户端中,自动为所有操作添加前缀。

## 注意事项

- 确保在整个应用程序中一致地使用前缀,以避免键冲突。
- 某些 Redis 命令(如 EVAL 和 SCRIPT)可能需要特殊处理,请参考源代码中的 `AssembleCMD` 方法以了解详细信息。