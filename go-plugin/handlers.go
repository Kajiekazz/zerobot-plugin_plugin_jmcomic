package jmcomic

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/FloatTech/zerobot/common/message"
	zlog "github.com/FloatTech/zerobot/common/log"
	"github.com/FloatTech/zerobot/handlers"
	zero "github.com/FloatTech/zerobot/core"
)

const (
	cmdPrefix = "jm" // 硬编码命令前缀
)

// handleGenericCommand 是一个通用的命令分发器，用于 "jm ..."
// 它会检查 "jm" 后面的第一个词是什么，然后分发到具体的处理器
// 或者如果第一个词看起来像漫画ID，则尝试直接下载
func handleGenericCommand(ctx *zero.Ctx) {
	// 获取 "jm " 后面的所有文本
	// ZeroBot 的 ctx.Event.RawMessage 或 ctx.Event.Message.ExtractPlainText()
	// 或者如果命令注册时使用了 RegexCmd(true)，则 ctx.State["matched"]
	// 假设注册方式为 engine.OnCommand("jm", zero.RegexCmd(true))
	// 那么 ctx.State["matched"] 将包含 "jm" 后面的内容 (不含 "jm" 本身)
	
	// 更可靠的方式是依赖于 `ctx.Event.GetArgs()` 如果命令注册时指定了参数解析
	// 或者自己解析 `ctx.Event.GetMessage()` (获取原始消息链)
	
	fullArgString := ""
	// 尝试从 ctx.State["matched"] 获取 (如果使用 RegexCmd(true) 且命令是 "jm")
	if matched, ok := ctx.State["matched"].(string); ok {
		fullArgString = strings.TrimSpace(matched)
	} else {
		// 后备：尝试从原始消息中提取 "jm "之后的内容
		rawMsg := ctx.Event.GetMessage().ExtractPlainText()
		if strings.HasPrefix(strings.ToLower(rawMsg), cmdPrefix+" ") {
			fullArgString = strings.TrimSpace(rawMsg[len(cmdPrefix)+1:])
		} else if strings.ToLower(rawMsg) == cmdPrefix { // 用户只发送了 "jm"
			handleHelp(ctx) // 显示帮助
			return
		} else {
			// 如果消息不是以 "jm " 开头，理论上不应该进入这个处理器
			// 但为了健壮性，可以记录一下
			zlog.Warnf("[%s] Generic handler called with unexpected message: %s", pluginName, rawMsg)
			return
		}
	}

	if fullArgString == "" { // 用户只发送了 "jm" (已被上面的分支处理，这里是双重检查)
		handleHelp(ctx)
		return
	}

	parts := strings.Fields(fullArgString)
	if len(parts) == 0 { // 理论上不会发生，因为 fullArgString 非空
		handleHelp(ctx)
		return
	}

	command := strings.ToLower(parts[0])
	args := parts[1:]

	switch command {
	case "help":
		handleHelp(ctx)
	case "search":
		handleSearchComic(ctx, args)
	case "detail":
		handleComicDetail(ctx, args)
	case "download": // "jm download <albumID> <chapterIDs...>"
		handleDownloadChapters(ctx, args)
	default:
		// 如果第一个参数不是已知的子命令，我们检查它是否可能是漫画ID
		// 这是一个简化的ID检查，实际JM ID可能有特定格式 (如纯数字，或带前缀)
		// 我们假设漫画ID是数字或者包含数字，并且后面可能跟着章节ID
		// 例如: "jm 12345 67890" -> albumID=12345, chapterIDs=[67890]
		//       "jm JM12345 ch1 ch2" -> albumID=JM12345, chapterIDs=[ch1, ch2]
		if len(parts) >= 2 && isPotentialAlbumID(parts[0]) { // parts[0]是潜在albumID, parts[1:]是潜在chapterIDs
			// 将 parts[0] 作为 albumID，parts[1:] 作为 chapterIDs 调用下载处理器
			// 注意：这里我们复用 handleDownloadChapters，但它期望的 args 是 [albumID, chapterID1, chapterID2...]
			// 所以我们需要重新构造一下参数
			downloadArgs := make([]string, 0, len(parts))
			downloadArgs = append(downloadArgs, parts[0]) // albumID
			downloadArgs = append(downloadArgs, parts[1:]...) // chapterIDs
			handleDownloadChapters(ctx, downloadArgs) // 直接调用下载
		} else {
			// 如果不匹配任何已知命令，也不是直接下载格式，则显示帮助
			ctx.SendChain(message.Text(fmt.Sprintf("未知命令 '%s' 或格式错误。\n发送 '%s help' 查看帮助。", command, cmdPrefix)))
		}
	}
}

// isPotentialAlbumID 简单检查字符串是否可能是漫画ID (例如包含数字)
// 你可能需要根据实际的JM漫画ID格式来改进这个函数
var albumIDRegex = regexp.MustCompile(`(jm)?\d+`) // 匹配 jm12345 或 12345 这样的格式
func isPotentialAlbumID(s string) bool {
	return albumIDRegex.MatchString(strings.ToLower(s))
	// 或者更简单的：
	// for _, r := range s {
	// 	if unicode.IsDigit(r) {
	// 		return true
	// 	}
	// }
	// return false
}


// handleHelp 显示帮助信息
func handleHelp(ctx *zero.Ctx) {
	helpMsg := fmt.Sprintf("%s 插件帮助 (JMComic):\n"+
		"1. %s help - 显示此帮助信息\n"+
		"2. %s search <关键词> - 搜索漫画\n"+
		"3. %s detail <漫画ID> - 获取漫画详情\n"+
		"4. %s download <漫画ID> <章节ID1> [章节ID2...] - 下载指定章节\n"+
		"   或直接: %s <漫画ID> <章节ID1> [章节ID2...] - 快速下载",
		strings.ToTitle(pluginName), cmdPrefix, cmdPrefix, cmdPrefix, cmdPrefix, cmdPrefix)
	ctx.SendChain(message.Text(helpMsg))
}

// handleSearchComic 处理搜索漫画命令
// args 是 "search" 后面的参数列表
func handleSearchComic(ctx *zero.Ctx, args []string) {
	if len(args) == 0 {
		ctx.SendChain(message.Text("请输入搜索关键词！例如: " + cmdPrefix + " search <关键词>"))
		return
	}
	keyword := strings.Join(args, " ")

	reqCtx, cancel := context.WithTimeout(context.Background(), cfg.timeoutDuration)
	defer cancel()

	ctx.SendChain(message.Text(fmt.Sprintf("正在搜索漫画: %s ...", keyword)))
	results, err := SearchComic(reqCtx, keyword)
	if err != nil {
		zlog.Errorf("[%s Handler] 搜索 '%s' 失败: %v", pluginName, keyword, err)
		errMsg := fmt.Sprintf("搜索失败: %v", err)
		if len(errMsg) > 100 {
			errMsg = errMsg[:100] + "..."
		}
		ctx.SendChain(message.Text(errMsg))
		return
	}

	if len(results) == 0 {
		ctx.SendChain(message.Text(fmt.Sprintf("未找到与 '%s' 相关的漫画。", keyword)))
		return
	}

	var msgChain message.Chain
	msgChain = msgChain.Add(message.Text(fmt.Sprintf("找到 %d 个结果:\n", len(results))))

	for i, comic := range results {
		if i >= cfg.MaxSearchResultsDisplay {
			msgChain = msgChain.Add(message.Text(fmt.Sprintf("...等共 %d 个结果。\n", len(results))))
			break
		}
		comicInfo := fmt.Sprintf("%d. %s (ID: %s)\n   作者: %s\n", i+1, comic.Title, comic.ID, comic.Author)
		msgChain = msgChain.Add(message.Text(comicInfo))
		// 封面图发送逻辑 (可选)
	}
	msgChain = msgChain.Add(message.Text(fmt.Sprintf("\n使用 %s detail <漫画ID> 查看详情和章节。", cmdPrefix)))
	ctx.SendChain(msgChain)
}

// handleComicDetail 处理获取漫画详情命令
// args 是 "detail" 后面的参数列表 (期望只有一个：漫画ID)
func handleComicDetail(ctx *zero.Ctx, args []string) {
	if len(args) == 0 {
		ctx.SendChain(message.Text("请输入漫画ID！例如: " + cmdPrefix + " detail <漫画ID>"))
		return
	}
	albumID := args[0]

	reqCtx, cancel := context.WithTimeout(context.Background(), cfg.timeoutDuration)
	defer cancel()

	ctx.SendChain(message.Text(fmt.Sprintf("正在获取漫画 %s 的详情...", albumID)))
	detail, err := GetComicDetail(reqCtx, albumID)
	if err != nil {
		zlog.Errorf("[%s Handler] 获取详情 '%s' 失败: %v", pluginName, albumID, err)
		errMsg := fmt.Sprintf("获取详情失败: %v", err)
		if len(errMsg) > 100 {
			errMsg = errMsg[:100] + "..."
		}
		ctx.SendChain(message.Text(errMsg))
		return
	}

	var msgChain message.Chain
	titleInfo := fmt.Sprintf("漫画: %s (ID: %s)\n作者: %s\n标签: %s\n", detail.Title, detail.ID, detail.Author, detail.Tags)
	msgChain = msgChain.Add(message.Text(titleInfo))
	
	desc := detail.Description
	if len(desc) > 200 {
		desc = desc[:200] + "..."
	}
	msgChain = msgChain.Add(message.Text(fmt.Sprintf("简介: %s\n", desc)))
	// 封面图 (可选)

	msgChain = msgChain.Add(message.Text("\n章节列表 (部分):\n"))
	for i, chapter := range detail.Chapters {
		if i >= cfg.MaxChaptersDisplay {
			msgChain = msgChain.Add(message.Text(fmt.Sprintf("...等共 %d 个章节。\n", len(detail.Chapters))))
			break
		}
		chapterInfo := fmt.Sprintf("%d. %s (章节ID: %s, 页数: %d)\n", i+1, chapter.Title, chapter.ID, chapter.PageCount)
		msgChain = msgChain.Add(message.Text(chapterInfo))
	}
	msgChain = msgChain.Add(message.Text(fmt.Sprintf("\n使用 %s download %s <章节ID1> ... 或直接 %s %s <章节ID1> ... 下载。", cmdPrefix, albumID, cmdPrefix, albumID)))
	ctx.SendChain(msgChain)
}

// handleDownloadChapters 处理下载章节命令
// args 是 "download" 后面的参数，或者直接是 "jm" 后面的参数 (albumID, chapterIDs...)
func handleDownloadChapters(ctx *zero.Ctx, args []string) {
	if len(args) < 2 { // 至少需要 albumID 和一个 chapterID
		ctx.SendChain(message.Text(fmt.Sprintf("参数不足！格式: %s [download] <漫画ID> <章节ID1> [章节ID2...]", cmdPrefix)))
		return
	}

	albumID := args[0]
	chapterIDs := args[1:]

	if len(chapterIDs) == 0 { // 理论上已被上面的检查覆盖
		ctx.SendChain(message.Text("请输入至少一个章节ID进行下载！"))
		return
	}

	downloadTimeout := cfg.timeoutDuration * time.Duration(len(chapterIDs))
	if downloadTimeout > 3*time.Minute {
		downloadTimeout = 3 * time.Minute
	}
	if downloadTimeout < cfg.timeoutDuration {
		downloadTimeout = cfg.timeoutDuration
	}

	reqCtx, cancel := context.WithTimeout(context.Background(), downloadTimeout)
	defer cancel()

	ctx.SendChain(message.Text(fmt.Sprintf("正在为漫画 %s 提交章节 %v 的下载请求...", albumID, chapterIDs)))

	apiMsg, pathHint, err := DownloadChapters(reqCtx, albumID, chapterIDs)
	if err != nil {
		zlog.Errorf("[%s Handler] 下载漫画 '%s' 章节 %v 失败: %v. API消息: %s", pluginName, albumID, chapterIDs, err, apiMsg)
		errMsg := fmt.Sprintf("下载请求失败: %v", err)
		if apiMsg != "" {
			errMsg = fmt.Sprintf("下载请求失败: %s", apiMsg)
		}
		if len(errMsg) > 150 {
			errMsg = errMsg[:150] + "..."
		}
		ctx.SendChain(message.Text(errMsg))
		return
	}

	responseMsg := fmt.Sprintf("下载请求已提交: %s", apiMsg)
	if pathHint != "" {
		responseMsg += fmt.Sprintf("\n提示: 文件可能保存在API服务器的 %s 目录中。", pathHint)
	}
	responseMsg += "\n请注意：下载在API服务器端进行，完成后文件不会直接发送给您，需从服务器获取。"
	ctx.SendChain(message.Text(responseMsg))
}

// MustRegisterHandlers 注册命令处理器到指定的引擎
// 这个函数由 jmcomic.go 中的 init -> OnLoad 调用
func MustRegisterHandlers(engine *zero.Engine) {
	// 我们注册一个顶级的 "jm" 命令，它会捕获所有 "jm ..." 的消息
	// 然后在 handleGenericCommand 中进行二次分发
	// 使用 RegexCmd(true) 使 ctx.State["matched"] 包含 "jm" 后面的所有内容
	// 或者不使用 RegexCmd，然后在 handleGenericCommand 中解析原始消息
	// 为了简单和灵活，我们让 handleGenericCommand 自己解析参数
	engine.OnCommand(cmdPrefix, zero.OnlyToMe(false)).SetBlock(true).Handle(handlers.NewCtxCmd(handleGenericCommand))
	
	zlog.Infof("[%s] 主命令 '%s' 已注册，将进行二次分发。", pluginName, cmdPrefix)
}
