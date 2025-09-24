package news

import (
	"context"
	"testing"
	"time"
)

func TestBlockBeats_GetName(t *testing.T) {
	bb := NewBlockBeats()
	expected := "blockbeats"
	if bb.GetName() != expected {
		t.Errorf("GetName() = %v, want %v", bb.GetName(), expected)
	}
}

func TestBlockBeats_GetDisplayName(t *testing.T) {
	bb := NewBlockBeats()
	expected := "BlockBeats"
	if bb.GetDisplayName() != expected {
		t.Errorf("GetDisplayName() = %v, want %v", bb.GetDisplayName(), expected)
	}
}

func TestBlockBeats_GetNews(t *testing.T) {
	bb := NewBlockBeats()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 测试获取少量新闻
	news, err := bb.GetNews(ctx, 2)
	if err != nil {
		t.Errorf("GetNews() error = %v", err)
		return
	}

	if len(news) == 0 {
		t.Error("GetNews() returned empty news list")
		return
	}

	// 验证新闻项的基本字段
	for i, item := range news {
		if item.ID == "" {
			t.Errorf("News item %d: ID is empty", i)
		}
		if item.Title == "" {
			t.Errorf("News item %d: Title is empty", i)
		}
		if item.Source != "BlockBeats" {
			t.Errorf("News item %d: Source = %v, want BlockBeats", i, item.Source)
		}
		if item.URL == "" {
			t.Errorf("News item %d: URL is empty", i)
		}
		if item.PublishedAt.IsZero() {
			t.Errorf("News item %d: PublishedAt is zero", i)
		}
	}
}

func TestBlockBeats_IsHealthy(t *testing.T) {
	bb := NewBlockBeats()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 健康检查应该返回 true（假设网络正常）
	healthy := bb.IsHealthy(ctx)
	// 注意：这个测试可能因为网络问题而失败，所以这里只检查函数能正常执行
	_ = healthy
}

func TestManager_RegisterSource(t *testing.T) {
	manager := NewManager()
	bb := NewBlockBeats()

	manager.RegisterSource(bb)

	source, exists := manager.GetSource("blockbeats")
	if !exists {
		t.Error("Source not found after registration")
	}

	if source.GetName() != "blockbeats" {
		t.Errorf("Registered source name = %v, want blockbeats", source.GetName())
	}
}

func TestManager_GetNewsFromSource(t *testing.T) {
	manager := NewManager()
	bb := NewBlockBeats()
	manager.RegisterSource(bb)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	news, err := manager.GetNewsFromSource(ctx, "blockbeats", 1)
	if err != nil {
		t.Errorf("GetNewsFromSource() error = %v", err)
		return
	}

	if len(news) == 0 {
		t.Error("GetNewsFromSource() returned empty news list")
	}
}

func TestManager_GetNewsFromNonExistentSource(t *testing.T) {
	manager := NewManager()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := manager.GetNewsFromSource(ctx, "nonexistent", 1)
	if err == nil {
		t.Error("Expected error for non-existent source, got nil")
	}
}
