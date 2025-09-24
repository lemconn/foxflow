package news

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"
)

// Example 展示如何使用新闻模块
func Example() {
	// 创建新闻管理器
	manager := NewManager()

	// 注册 BlockBeats 新闻源
	blockBeats := NewBlockBeats()
	manager.RegisterSource(blockBeats)

	// 创建上下文
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 示例1: 从指定新闻源获取新闻
	fmt.Println("=== 从 BlockBeats 获取 5 条新闻 ===")
	news, err := manager.GetNewsFromSource(ctx, "blockbeats", 5)
	if err != nil {
		log.Printf("获取新闻失败: %v", err)
		return
	}

	for i, item := range news {
		fmt.Printf("新闻 %d:\n", i+1)
		fmt.Printf("  标题: %s\n", item.Title)
		fmt.Printf("  来源: %s\n", item.Source)
		fmt.Printf("  时间: %s\n", item.PublishedAt.Format("2006-01-02 15:04:05"))
		fmt.Printf("  链接: %s\n", item.URL)
		fmt.Printf("  内容: %s\n", truncateString(item.Content, 100))
		fmt.Println("   " + strings.Repeat("-", 50))
	}

	// 示例2: 检查新闻源健康状态
	fmt.Println("=== 检查新闻源健康状态 ===")
	availableSources := manager.GetAvailableSources(ctx)
	fmt.Printf("可用的新闻源: %v\n", availableSources)

	// 示例3: 获取所有新闻源的信息
	fmt.Println("=== 所有已注册的新闻源 ===")
	allSources := manager.GetAllSources()
	for name, source := range allSources {
		fmt.Printf("名称: %s, 展示名: %s\n", name, source.GetDisplayName())
	}
}

// RunExample 运行示例程序（可在测试中调用）
func RunExample() {
	fmt.Println("🚀 新闻模块示例程序")
	fmt.Println("==================")
	Example()
	fmt.Println("\n🎉 示例程序执行完成！")
}

// truncateString 截断字符串到指定长度
func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
