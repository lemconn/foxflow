package news

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// BlockBeatsResponse BlockBeats API 响应结构
type BlockBeatsResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Page  int                  `json:"page"`
		Limit int                  `json:"limit"`
		List  []BlockBeatsNewsItem `json:"list"`
	} `json:"data"`
}

// BlockBeatsNewsItem BlockBeats 新闻项结构
type BlockBeatsNewsItem struct {
	ID               int      `json:"id"`
	ArticleID        int      `json:"article_id"`
	ContentID        int      `json:"content_id"`
	Type             int      `json:"type"`
	IsShowHome       int      `json:"is_show_home"`
	IsDetective      int      `json:"is_detective"`
	IsTop            int      `json:"is_top"`
	IsOriginal       int      `json:"is_original"`
	SpecialID        int      `json:"special_id"`
	TopicID          int      `json:"topic_id"`
	Ios              int      `json:"ios"`
	IsFirst          int      `json:"is_first"`
	IsHot            int      `json:"is_hot"`
	AddTime          int64    `json:"add_time"`
	ImgURL           string   `json:"img_url"`
	CImgURL          string   `json:"c_img_url"`
	URL              string   `json:"url"`
	CryptoToken      string   `json:"crypto_token"`
	Title            string   `json:"title"`
	Lang             string   `json:"lang"`
	PID              int      `json:"p_id"`
	Abstract         string   `json:"abstract"`
	Content          string   `json:"content"`
	IsHyper          int      `json:"is_hyper"`
	CommentCount     int      `json:"comment_count"`
	TagList          []string `json:"tag_list"`
	CollectionStatus int      `json:"collection_status"`
}

// BlockBeats BlockBeats 新闻源实现
type BlockBeats struct {
	baseURL    string
	httpClient *http.Client
}

// NewBlockBeats 创建新的 BlockBeats 新闻源实例
func NewBlockBeats() *BlockBeats {
	return &BlockBeats{
		baseURL: "https://api.blockbeats.cn/v2/newsflash/list",
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GetName 获取新闻源名称
func (b *BlockBeats) GetName() string {
	return "blockbeats"
}

// GetDisplayName 获取新闻源展示名称
func (b *BlockBeats) GetDisplayName() string {
	return "BlockBeats"
}

// GetNews 获取新闻列表
func (b *BlockBeats) GetNews(ctx context.Context, count int) ([]NewsItem, error) {
	var allNews []NewsItem
	page := 1
	limit := 50 // BlockBeats 每页最大数量
	endTime := time.Now().Unix()

	for len(allNews) < count {
		// 计算当前页需要获取的数量
		remaining := count - len(allNews)
		currentLimit := limit
		if remaining < limit {
			currentLimit = remaining
		}

		// 构建请求URL
		url := fmt.Sprintf("%s?page=%d&limit=%d&ios=1&end_time=%d&detective=-2",
			b.baseURL, page, currentLimit, endTime)

		// 发送请求
		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("创建请求失败: %w", err)
		}

		// 设置请求头
		b.setHeaders(req)

		// 执行请求
		resp, err := b.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("请求失败: %w", err)
		}
		defer resp.Body.Close()

		// 读取响应
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("读取响应失败: %w", err)
		}

		// 解析响应
		var blockBeatsResp BlockBeatsResponse
		if err := json.Unmarshal(body, &blockBeatsResp); err != nil {
			return nil, fmt.Errorf("解析响应失败: %w", err)
		}

		// 检查响应状态
		if blockBeatsResp.Code != 0 {
			return nil, fmt.Errorf("API 返回错误: %s", blockBeatsResp.Msg)
		}

		// 转换新闻数据
		for _, item := range blockBeatsResp.Data.List {
			newsItem := b.convertToNewsItem(item)
			allNews = append(allNews, newsItem)

			// 更新 endTime 为当前新闻的时间戳
			endTime = item.AddTime
		}

		// 如果没有更多数据，退出循环
		if len(blockBeatsResp.Data.List) == 0 {
			break
		}

		page++
	}

	// 如果获取的新闻数量超过请求的数量，截取
	if len(allNews) > count {
		allNews = allNews[:count]
	}

	return allNews, nil
}

// IsHealthy 检查新闻源是否健康可用
func (b *BlockBeats) IsHealthy(ctx context.Context) bool {
	// 尝试获取少量新闻来检查服务是否可用
	_, err := b.GetNews(ctx, 1)
	return err == nil
}

// setHeaders 设置请求头
func (b *BlockBeats) setHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json, text/plain, */*")
	req.Header.Set("Accept-Language", "zh-CN,zh;q=0.9,en;q=0.8,zh-TW;q=0.7,ja;q=0.6")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Origin", "https://www.theblockbeats.info")
	req.Header.Set("Referer", "https://www.theblockbeats.info/")
	req.Header.Set("lang", "cn")
	req.Header.Set("Sec-Fetch-Dest", "empty")
	req.Header.Set("Sec-Fetch-Mode", "cors")
	req.Header.Set("Sec-Fetch-Site", "cross-site")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/140.0.0.0 Safari/537.36")
	req.Header.Set("sec-ch-ua", `"Chromium";v="140", "Not=A?Brand";v="24", "Google Chrome";v="140"`)
	req.Header.Set("sec-ch-ua-mobile", "?0")
	req.Header.Set("sec-ch-ua-platform", `"macOS"`)
	req.Header.Set("token", "")
}

// convertToNewsItem 将 BlockBeats 新闻项转换为统一格式
func (b *BlockBeats) convertToNewsItem(item BlockBeatsNewsItem) NewsItem {
	// 构建新闻链接
	url := item.URL
	if url == "" {
		url = fmt.Sprintf("https://www.theblockbeats.info/flash/%d", item.ArticleID)
	}

	// 清理 HTML 标签
	content := b.stripHTML(item.Content)

	// 转换时间戳
	publishedAt := time.Unix(item.AddTime, 0)

	return NewsItem{
		ID:          strconv.Itoa(item.ID),
		Title:       item.Title,
		Content:     content,
		URL:         url,
		Source:      b.GetDisplayName(),
		PublishedAt: publishedAt,
		Tags:        item.TagList,
		ImageURL:    item.ImgURL,
	}
}

// stripHTML 移除 HTML 标签
func (b *BlockBeats) stripHTML(html string) string {
	// 简单的 HTML 标签移除
	content := html
	content = strings.ReplaceAll(content, "<p>", "")
	content = strings.ReplaceAll(content, "</p>", "\n")
	content = strings.ReplaceAll(content, "<br>", "\n")
	content = strings.ReplaceAll(content, "<br/>", "\n")
	content = strings.ReplaceAll(content, "<br />", "\n")

	// 移除其他常见的 HTML 标签
	content = strings.ReplaceAll(content, "<strong>", "")
	content = strings.ReplaceAll(content, "</strong>", "")
	content = strings.ReplaceAll(content, "<em>", "")
	content = strings.ReplaceAll(content, "</em>", "")
	content = strings.ReplaceAll(content, "<b>", "")
	content = strings.ReplaceAll(content, "</b>", "")
	content = strings.ReplaceAll(content, "<i>", "")
	content = strings.ReplaceAll(content, "</i>", "")

	// 清理多余的换行符
	content = strings.TrimSpace(content)
	content = strings.ReplaceAll(content, "\n\n", "\n")

	return content
}
