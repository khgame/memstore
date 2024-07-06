# CacheKey

CacheKey is a powerful and flexible cache key generation library for Go, 
offering type-safe, pre-defined schemas for robust cache key management 
in your applications.

## Features

- Type-safe, pre-defined cache key schemas
- Support for both struct and single value parameters
- Placeholder-based cache key schema definition
- Custom field name mapping via `cachekey` tags
- Key formatting and validation utilities
- Partial application support for dynamic key generation
- Thread-safe and efficient key generation

## Installation

To install CacheKey, use the `go get` command:

```bash
go get github.com/khgame/memstore
```

and import with
```go
import "github.com/khgame/memstore/cachekey"
```

## Quick Start

### Using Structs

```go
package main

import (
    "fmt"
    "time"

    "github.com/khgame/memstore/cachekey"
)

type User struct {
    ID   int    `cachekey:"user_id"`
    Name string
}

func main() {
    schema := cachekey.MustNewSchema[User]("user:{user_id}:{name}", 10*time.Minute)
    key, err := schema.Build(User{ID: 123, Name: "Alice"})
    if err != nil {
        panic(err)
    }
    fmt.Println(key) // Output: user:123:Alice
}
```

### Using Single Values

```go
schema := cachekey.MustNewSchema[int64]("user:{id}", 10*time.Minute)
key, err := schema.Build(123)
if err != nil {
    panic(err)
}
fmt.Println(key) // Output: user:123
```

## Benefits of Pre-defined, Type-checked Schemas

CacheKey's approach of using pre-defined, type-checked schemas offers several significant advantages:

1. **Error Prevention at Initialization**:
   Schemas are validated at creation time, catching mismatches between the schema and the associated type immediately. This prevents runtime errors and ensures that all cache keys are structurally correct from the start.

```go
   // This will panic at initialization, preventing future runtime errors
   invalidSchema := cachekey.MustNewSchema[User]("user:{invalid_field}", 10*time.Minute)
```

2. **Type Safety, Avoiding Dynamic Format Pitfalls**:
   The use of generics ensures that only the correct types can be used with each schema, eliminating a whole class of runtime errors.

```go
   userSchema := cachekey.MustNewSchema[User]("user:{user_id}:{name}", 10*time.Minute)
   // This won't compile, catching the error at compile-time
   userSchema.Build(123) // Error: cannot use 123 (type int) as type User
```

3. **Observability Enhancement**:
   Pre-defined schemas can be easily integrated with monitoring, tracing, and logging systems. They provide consistent key structures that can be automatically tagged and tracked.

```go
   func getCachedUser(id int, name string) {
       key := userSchema.MustBuild(User{ID: id, Name: name})
       span := tracer.StartSpan("cache_operation")
	   // Add schema fingerprint for tracking
	   // its absolutlly better than using the dynamic format
       span.SetTag("fingerprint", userSchema.FingerPrint) 
       // ... rest of the function
   }
```

4. **Schema Governance**:
   Having pre-defined schemas allows for centralized management and governance of cache key structures across your application or organization.

   ```go
      // In a central package or configuration file
      var (
          UserCacheSchema    = cachekey.MustNewSchema[User]("user:{user_id}:{name}", 10*time.Minute)
          ProductCacheSchema = cachekey.MustNewSchema[Product]("product:{product_id}", 1*time.Hour)
      )
   ```

5. **Performance Optimization**:
   Pre-defined schemas allow for optimizations that wouldn't be possible with fully dynamic key generation, potentially improving performance in high-throughput scenarios.

## Advanced Usage

### Custom Field Name Mapping

Use the `cachekey` tag to customize field name mapping:

```go
type Product struct {
    ID    int     `cachekey:"product_id"`
    Name  string
    Price float64 `cachekey:"product_price"`
}

schema := cachekey.MustNewSchema[Product]("product:{product_id}:{name}:{product_price}", 10*time.Minute)
key, _ := schema.Build(Product{ID: 456, Name: "Laptop", Price: 999.99})
fmt.Println(key) // Output: product:456:Laptop:999.99
```

### Dynamic Key Generation with ToFormat

While pre-defined schemas are preferred for their safety and performance, CacheKey also supports dynamic key generation when needed:

```go
schema := cachekey.MustNewSchema[User]("user:{user_id}:{name}", 10*time.Minute)
format := schema.ToFormat()

// Generate keys dynamically
key1 := format.Make(123, "Alice")
key2 := format.Make(456, "Bob")

fmt.Println(key1) // Output: user:123:Alice
fmt.Println(key2) // Output: user:456:Bob
```

### Partial Application with FormatPartial

The `Partial()` method allows for efficient creation of partially applied key generators:

```go
schema := cachekey.MustNewSchema[User]("user:{user_id}:{name}", 10*time.Minute)
format := schema.ToFormat()

// Create a partial key generator for a specific user ID
userKeyGen := format.Partial(789)

// Generate keys for different names
key1 := userKeyGen("Charlie")
key2 := userKeyGen("David")

fmt.Println(key1) // Output: user:789:Charlie
fmt.Println(key2) // Output: user:789:David
```

## Best Practices

1. **Define Schemas Centrally**: Create a central location for all your cache key schemas. This improves maintainability and allows for easier schema governance.

```go
   // cachekeys/schemas.go
   package cachekeys

   var (
       UserSchema    = cachekey.MustNewSchema[User]("user:{user_id}:{name}", 10*time.Minute)
       ProductSchema = cachekey.MustNewSchema[Product]("product:{product_id}", 1*time.Hour)
   )
```

2. **Use `cachekey` Tags**: For structs, use `cachekey` tags to explicitly define field mappings. This improves clarity and allows for more flexible key schemas.

3. **Validate Schemas Early**: Use `NewSchema` instead of `MustNewSchema` during development to catch schema errors early. Switch to `MustNewSchema` in production for panic-free operation.

4. **Reuse Schema Instances**: Create schema instances once and reuse them for better performance.

5. **Prefer Pre-defined Schemas**: While dynamic key generation is supported, prefer pre-defined schemas for their type safety and performance benefits.

6. **Integrate with Observability Tools**: Use the consistent structure of your cache keys to enhance logging, monitoring, and tracing in your application.

7. **Consider Performance**: For high-throughput scenarios, benchmark your key generation and consider using single value schemas or pre-compiled formats where possible.

## Error Handling

CacheKey provides detailed error messages to help you identify and resolve issues:

```go
schema, err := cachekey.NewSchema[User]("user:{invalid_field}", 10*time.Minute)
if err != nil {
    fmt.Println("Schema creation error:", err)
    // Output: Schema creation error: invalid schema for given type: fields: [user_id name]
}

schema := cachekey.MustNewSchema[User]("user:{user_id}:{name}", 10*time.Minute)
_, err = schema.Build(User{ID: 123}) // Missing Name field
if err != nil {
    fmt.Println("Key generation error:", err)
    // Output: Key generation error: field name does not exist, paramsMap= map[user_id:123]
}
```

## License

CacheKey is released under the MIT License. See tmhe [LICENSE](../LICENSE) file for more details.