//go:build integration

package integration

// Shutdown terminates shared testcontainers started during integration tests.
func Shutdown() {
	if mysqlCleanup != nil {
		mysqlCleanup()
		mysqlCleanup = nil
	}
	if redisCleanup != nil {
		redisCleanup()
		redisCleanup = nil
	}
}
