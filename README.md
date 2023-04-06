# memstore
This repository provides an implementation of an in-memory storage that supports IsDirty/Save methods, as well as a standardized encapsulation of CacheKey. Based on this storage model, the repository also implements a set of "resource management" interfaces for managing resources efficiently.

## feature 

### In-Memory Storage
This repository provides a lightweight in-memory storage implementation that supports key-value storage and provides IsDirty/Save methods to easily determine if the cache has expired and manually save data. This storage method can effectively improve the performance of the application in cases where memory is limited.

### CacheKey Encapsulation
CacheKey is a commonly used concept, and this repository provides a standardized encapsulation of CacheKey to facilitate the management and maintenance of CacheKey, avoiding data errors and performance degradation caused by mixed-up CacheKeys.

### Resource Management
Based on the in-memory storage and CacheKey encapsulation, this repository implements a set of "resource management" interfaces that provide common operations such as loading, saving, and deleting resources, and also supports multiple data types for storage and management, making it easy to develop and maintain applications.

We welcome everyone to use and contribute to the code!