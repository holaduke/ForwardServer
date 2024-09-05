package filter
var (
DefaultDisallowPath     []string          = []string{".woff",".png", ".js", ".jpg", ".jpeg", ".svg", ".mp4", ".css", ".mp3", ".gif", ".bmp", ".json", ".txt", ".m3u8", ".avi", ".ico"}
DefaultDisallowHostname []string = []string {
		"*google*",
		"*baidu*",
		"*cloudflare*",
		"*qq.com*",
		"*gov.cn",
		"*edu.cn*",
		"*firefox",
		"*tencent*",
		"*mozilla*",
		"*tencent*",
		"*apple*",
		"*taobao*",
		"*apple*",
		"*umeng*",
		"*facebook*",
		"*github*",
		"*miui*",
		"*xiaomi*",
		"*alipay*",
		"*icloud*",
		"*amap.com*",
	}
DefaultAllowHostname []string = []string{}
)

