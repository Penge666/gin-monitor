package metrics

import (
	"fmt"
	"testing"
)

func TestBloomFilter(t *testing.T) {
	// 初始化布隆过滤器
	bf := NewBloomFilter()

	// 测试添加和查询已存在的元素
	testValues := []string{"192.168.0.1", "10.0.0.2", "172.16.0.3"}
	for _, value := range testValues {
		bf.Add(value)
	}

	for _, value := range testValues {
		if !bf.Contains(value) {
			t.Errorf("Expected value %s to be present in the BloomFilter, but it was not.", value)
		}
	}

	// 测试查询不存在的元素
	nonExistentValues := []string{"192.168.0.4", "10.0.0.3", "172.16.0.4"}
	for _, value := range nonExistentValues {
		if bf.Contains(value) {
			t.Errorf("Expected value %s to not be present in the BloomFilter, but it was.", value)
		}
	}

	// 测试空字符串处理
	if bf.Contains("") {
		t.Error("Expected BloomFilter to return false for empty string, but it did not.")
	}
	// 测试存在的字符串
	if !bf.Contains("192.168.0.1") {
		t.Error("Expected BloomFilter to return false for have ip.")
	}
}
func TestPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered from panic:", r)
			//ctx.AbortWithStatus(http.StatusInternalServerError)
			//return
		}
		fmt.Println("hhh")
	}()

	// 这里触发 panic
	panic("something went wrong!")
}
