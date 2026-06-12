# 設計文件:Lexical 文章評分、Expose 與 Auto-fix

- 日期:2026-06-12
- 狀態:已與使用者逐節確認定案;經三視角對抗式審查(一致性/可實作性/與報告吻合度)修訂
- 依據:`docs/lexical-tech-article-best-practices.md`(最佳實踐報告)、`docs/lexical-writing-guide.html`(撰寫指南)

## 1. 背景與目標

lexical-cli 目前是單向轉換器(Lexical JSON → Markdown/text/HTML)+ 以 JSON tree 為中心的 TUI。本設計將產品重塑為**雙向工作流**:

- **Markdown 是撰寫格式,Lexical JSON 是儲存/交付格式。**
- 工具負責三件事:評分(lint)、反向轉換(expose)、自動修復(fix)。

```
撰寫流:  user 寫 Markdown ──expose──▶ 符合最佳實踐的 Lexical JSON
載入流:  Lexical JSON ──▶ 評分 + 精煉成有損 Markdown 預覽
修復流:  Lexical JSON ──lint──▶ FixProfile 渲染 MD ──expose──▶ 合規 JSON(auto-fix)
```

### 非目標

- 不做語意層規則(倒金字塔、front-loading、emoji 是否單獨承載語意、「列表引言必須是完整句子」的語意判定——靜態無法判定;引言的**結構**檢查見 `list-intro`)。
- 不做規則設定檔(YAML/JSON rule config);v1 門檻與權重寫死,`lint --rules` 可查。
- 不做 CI 把關旗標(`--fail-on` / `--min-score`);lint 完成一律 exit 0,輸入無效維持 exit 1。
- TUI 不做 Markdown 編輯;撰寫仍在使用者自己的編輯器。
- TUI findings 不做「跳到預覽對應位置」互動。

### 成功標準(可量測)

- 對任一合法 Lexical JSON 輸出 0–100 總分 + 五類子分數 + findings 清單(text 與 `--json` 兩種格式)。
- 撰寫指南範例文件(合規):總分 **≥95 且 0 筆 error 級 finding**;故意違規 fixture 觸發對應規則(逐規則正反例)。
- `fix` 後重新 lint:**after 分數 ≥ before 分數**,且 before 中所有「已修復清單」內的 findings 不再出現;損耗逐項申報(見 §6.2 的封閉清單)。
- 全程不依賴 Node.js;新增依賴僅 `github.com/yuin/goldmark`。

## 2. 執行順序(連續交付,每階段是下一階段的依賴)

1. **Lint 引擎**(`internal/lint`)+ `lint` 子命令
2. **Expose**(`internal/mdimport` + `internal/lexical` serializer + `internal/markdown` highlight 輸出)+ `expose` 子命令
3. **Auto-fix**(`internal/fix`)+ `fix` 子命令
4. **TUI 改版**(儀表板式,整合前三者)

## 3. CLI 介面

```
lexical-cli lint   [--json] [--rules] input.json        # 評分與建議
lexical-cli expose [-o out.json] input.md               # Markdown → Lexical JSON
lexical-cli fix    [-o out.json] [--strict] input.json  # auto-fix by best practice
lexical-cli tui    input.{json,md}                      # 儀表板 TUI
```

- 全部支援 stdin(比照現有主命令)。
- `lint --json`:輸出完整 Report 結構(machine-readable)。
- `lint --rules`:列出所有規則、severity、分類、門檻、是否 fixable,並標註門檻來源(報告引用 vs 本 spec 操作化推定),然後退出。
- Exit codes 沿用現有語意:`0` 正常完成(含有 findings 的 lint)、`1` 輸入無效/IO 失敗/strict 違規。

**CLI 重構備註**(現況:`internal/cli/flags.go` 只認 `tui` 一個子命令;`run.go` 在 dispatch 前無條件讀入並 parse Lexical JSON):

- dispatch 必須移到 readInput/Parse **之前**,由各子命令自行決定輸入格式(`expose` 與 `tui` 要吃 `.md`)。
- flags 改為 per-subcommand 作用域;特別注意**現有 `--strict`(轉換時遇 unsupported node 即失敗)與 `fix --strict`(損耗即失敗)語意不同**,分屬不同子命令、各自文件化。

## 4. 階段 1:Lint 引擎

### 4.1 套件結構

```
internal/lint/
  lint.go      // Run(doc *lexical.Document) Report
  rule.go      // Rule interface, Severity, Category
  finding.go   // Finding{RuleID, Severity, Category, Path, Message, Suggestion, Fixable}
  score.go     // findings → CategoryScores + Total
  walk.go      // DFS 走訪 + path 字串(沿用 --list-unsupported 的 path 格式)
  emoji.go     // rune-range emoji 偵測(不引外部依賴)
  cssstyle.go  // style 字串解析 + WCAG 相對亮度/對比計算(hex、rgb())
  rules/
    heading.go paragraph.go list.go inline.go color.go emoji_rules.go
    (各檔對應 *_test.go)
```

`Rule` interface:`ID() string`、`Category() Category`、`Check(ctx *Context) []Finding`。`Context` 提供已建好的走訪索引(段落清單、標題序列、全文 emoji 統計等),避免每條規則重複走樹。引擎不依賴 CLI/TUI,與 `internal/markdown` 同層。

### 4.2 名詞定義(normative,所有規則共用)

- **Run**:同一父節點下,**連續兄弟 TextNode 且皆帶有目標 format bit** 的最長序列。不帶該 bit 的任何節點(含純空白 text node、LinkNode、linebreak)都會中斷 run;LinkNode 邊界一律中斷 run(即使其子節點帶同 bit)。長度以 **rune** 計。適用於 `bold-overuse`、`highlight-budget`、`long-emphasis`。
- **段落容器**(`bold-overuse` 的計數範圍):ParagraphNode、ListItemNode、QuoteNode、TableCellNode 各自獨立計數;HeadingNode 豁免(標題本身已是強調)。
- **句界**(`paragraph-length` / `sentence-length`):全形 `。!?` 一律計為句界;半形 `.!?` 僅在**後接空白或文末**時計為句界(排除 `e.g.`、`v1.2`、檔名、網域的誤判)。省略號(U+2026,或連續 ≥2 個半形 `.`)不是句界。帶 code format bit(16)的 text node 不參與句子切分。
- **巢狀列表**:任意深度的每個 ListNode **獨立**套用列表規則;「包裝項」(children 僅含巢狀 ListNode 的 ListItemNode)**不計入**該層項目數(對 `list-length` 與 `list-single-item` 皆然)。
- **heading-skip 方向性**:僅當標題比**前一個標題深超過一層**(如 H2 → H4)時觸發;任何幅度的回升(H4 → H2)合規。文件第一個標題不做 skip 檢查(缺 H1 由 `single-h1` 負責)。

### 4.3 規則集 v1

severity 扣分:**error −15、warn −5、info −1**(扣在所屬分類,分類分數下限 0)。

Fixable 欄:✓ = 階段 3 的 FixProfile 主動修復;✓ˢ = 經 Markdown round-trip 的副作用自然消除(style 不過 Markdown);✗ = 內容問題,需人工。

**文章結構(權重 0.30)**

| Rule ID | Severity | Fixable | 邏輯 |
|---|---|---|---|
| `single-h1` | error | ✗ | H1 數量 ≠ 1(0 個報一筆;每多一個 H1 報一筆) |
| `heading-skip` | error | ✗ | 見 §4.2 方向性定義 |
| `consecutive-headings` | warn | ✗ | 兩個標題之間沒有任何正文節點 |
| `heading-depth` | info | ✗ | 出現 H4–H6 |
| `paragraph-length` | warn | ✗ | 段落 >6 句(取報告「5–6 句」的寬鬆端,rationale 記入 `--rules`) |
| `sentence-length` | info | ✗ | 單句 >26 個拉丁詞(空白分隔、含拉丁字母的 token);CJK 句豁免 |

**列表(權重 0.20;v1 僅列表規則,表格規則列 v1.1,類名暫不含表格)**

| Rule ID | Severity | Fixable | 邏輯 |
|---|---|---|---|
| `list-length` | warn | ✗ | bullet ListNode 有效項 >7(見 §4.2 巢狀定義) |
| `list-single-item` | info | ✗ | ListNode 有效項 = 1 |
| `procedure-steps` | warn | ✗ | number ListNode 有效項 >12 |
| `list-intro` | info | ✗ | ListNode 的前一個兄弟節點不是段落/標題(結構檢查;引言是否為「完整句子」屬語意,不檢查) |

**行內強調(權重 0.25)**

| Rule ID | Severity | Fixable | 邏輯 |
|---|---|---|---|
| `bold-overuse` | warn | ✗ | 同一段落容器內 bold run >1 |
| `bold-italic-combo` | warn | ✓ | format 同時含 bold(1)+italic(2);fix 保 bold 去 italic |
| `underline-usage` | error | ✓ | LinkNode 之外的文字含 underline(8) |
| `highlight-budget` | warn | ✗ | 全篇 highlight(128)run >5 |
| `all-caps` | warn | ✗ / ✓ | (a) 連續 ≥4 個全大寫拉丁詞(僅 [A-Z],1–3 詞標籤豁免)→ ✗;(b) format 含 IS_UPPERCASE(512)且該節點 ≥4 詞 → ✓(fix 去 bit) |
| `long-emphasis` | info | ✗ | 單一 bold 或 italic run >80 rune(**本 spec 操作化推定**,報告僅言「短語」無數值) |

**用色(權重 0.15)**

| Rule ID | Severity | Fixable | 邏輯 |
|---|---|---|---|
| `inline-color` | warn | ✓ˢ | text/element `style` 含 `color:` |
| `inline-background` | warn | ✓ˢ | `style` 含 `background-color:` |
| `low-contrast` | error | ✓ˢ | 字色對底色(無底色時假設 `#ffffff`)對比 <4.5:1;**統一用 4.5:1,不解析 font-size 做大字 3:1 豁免(刻意簡化,記入 `--rules`)**;僅支援 hex/rgb(),無法解析的值跳過不報 |

**Emoji(權重 0.10)**

| Rule ID | Severity | Fixable | 邏輯 |
|---|---|---|---|
| `emoji-in-sentence` | warn | ✗ | emoji 出現在白名單位置以外。白名單:**H2/H3** 標題文字開頭(H1 與 H4–H6 不豁免)、Quote 的第一個 inline 子節點位置(callout 圖示)、表格儲存格內(**任意位置——對報告「狀態欄」的靜態可判定性妥協,記入 `--rules`**) |
| `emoji-run` | warn | ✓ | 連續 ≥2 顆 emoji(中間無其他字元);fix 折疊為第一顆 |
| `emoji-variety` | info | ✗ | 全篇 distinct emoji >7 種 |

### 4.4 評分模型

- 分類分數 = `max(0, 100 − Σ該類扣分)`。
- 總分 = `round(Σ 分類分數 × 權重)`;權重:結構 0.30、列表 0.20、行內 0.25、用色 0.15、emoji 0.10。
- 文件不含某類元素時該類自然 100,不設 N/A 狀態。
- findings 依文件順序輸出;text 輸出 = 總分 + 五條分數 bar + findings;`--json` 輸出 Report 全結構。

### 4.5 CJK 與邊界情況

- `all-caps`(文字變體)與 `sentence-length` 僅適用拉丁文字;CJK 無大小寫、詞界不同,一律豁免。
- emoji 偵測用 Unicode rune-range 表(Emoticons、Misc Symbols & Pictographs、Supplemental Symbols、Transport;ZWJ 序列以首 rune 判定;variation selector 與 skin-tone modifier 視為同一顆 emoji 的一部分)。偵測器獨立成 `emoji.go` + 專屬測試表。
- 空文件:報 `single-h1`,其餘類 100。
- 未知節點型別:lint 跳過(`--list-unsupported` 已負責申報)。

## 5. 階段 2:Expose(Markdown → Lexical JSON)

### 5.1 前置工作項(本階段交付物,缺一不可)

1. **Lexical JSON serializer**(`internal/lexical`):目前 `Node` 只有 `UnmarshalJSON`,**repo 不存在任何序列化能力**。需實作 MarshalJSON/serializer 輸出標準 EditorState JSON:正確的小寫鍵名、多型 `format` 欄位(text=數字 bitmask、element=對齊字串)、`version` 欄位、依節點型別 omit-empty、未知節點以 `Raw` 原樣 passthrough。expose 輸出、fix 輸出、TUI 存 JSON 全部依賴此項。
2. **Node 欄位補齊**:TableCellNode 的 `headerState`(現有 struct 無此欄,無法表達「首列 header」)。
3. **Renderer 反向能力**(`internal/markdown`):`applyFormat` 目前**靜默丟棄** highlight(128)/subscript(32)/superscript(64)。需:highlight 改輸出 `==text==`(與 mdimport 的自訂 inline 規則互為反函數,§8 收斂測試的前提);sub/superscript 維持丟棄但**改為產生轉換 warning**(損耗申報)。

### 5.2 解析器與映射

- 引入 `github.com/yuin/goldmark`(CommonMark + GFM extensions:table、strikethrough、tasklist)。
- 新套件 `internal/mdimport`:goldmark AST → `lexical.Node` 樹 → 正規化 → serializer 輸出。

| Markdown | Lexical |
|---|---|
| `# h1`–`###### h6` | HeadingNode(tag h1–h6;不合規處由 lint 報,expose 不擅改) |
| 段落 | ParagraphNode |
| `-` / `1.` 列表 | ListNode bullet / number(含巢狀) |
| GFM task list `- [x]` | ListNode `check` + ListItemNode.Checked |
| `> ` blockquote | QuoteNode;**正準形狀:QuoteNode 直接持有 inline 子節點,blockquote 內多段落以 linebreak node 分隔後攤平**(與現有 renderer 的 quote 處理一致) |
| `> [!NOTE]`…`> [!CAUTION]`(五級 alert) | QuoteNode + 前置「圖示 emoji + 粗體標籤」inline 節點:📌 **Note:** / 💡 **Tip:** / ⭐ **Important:** / ⚠️ **Warning:** / 🚫 **Caution:**(標籤採 GitHub alert 英文名以保 OSS 一致;報告 §7.3 的中文標籤為 zh 對應,不在 v1 產生) |
| fenced code(含語言) | CodeNode(Language) |
| GFM table | TableNode/TableRowNode/TableCellNode(首列 headerState) |
| `---` | HorizontalRuleNode |
| `[text](url)` | LinkNode |
| `![alt](src)` | ImageNode(Src、AltText;caption/width 等欄位不在 Markdown 表達範圍,屬 §6.2 申報損耗) |
| `**`、`*`、`~~`、`` ` `` | format bitmask bold/italic/strikethrough/code |
| `==text==`(自訂 inline 規則) | format highlight(128) |
| 行內 HTML(`<u>` 等) | 不支援:原樣保留為純文字,並產生轉換 warning |

### 5.3 正規化原則

expose 只做**不改變內容語意**的正規化:

1. 相鄰同 format/style 的 text node 合併。
2. GFM task list → 原生 `check` listType。
3. Alert 語法 → 標準 callout 結構(上表)。
4. 輸出的 JSON 不含任何 `style` 字串(Markdown 無此概念,天然合規)。

會改變內容的修正(刪 emoji、拆段落)一律不做,由 lint 報告、由使用者或 fix 處理。

## 6. 階段 3:Auto-fix

### 6.1 管線

```
parse → lint(before) → markdown render with FixProfile → expose → lint(after)
```

`FixProfile` 是 `internal/markdown` renderer 的選項集,由 before-findings 驅動:

| 目標 | Fix 行為 |
|---|---|
| `underline-usage` | 渲染時丟棄 underline format;**LinkNode 內的合規 underline 同樣被丟棄(Markdown 無此表達),列入申報損耗** |
| `bold-italic-combo` | 保留 bold、丟棄 italic |
| `all-caps`(512 變體) | 丟棄 IS_UPPERCASE bit |
| `emoji-run` | 連續 emoji 折疊為第一顆 |
| style 字串(`inline-color`/`inline-background`/`low-contrast` 及未觸發規則的其他屬性) | 不過 Markdown,全部消除(✓ˢ) |

實作備註:`applyFormat`/`escapeText` 目前是無 receiver 的自由函式,接不到 options;需改為 converter method(或傳入 profile 參數)才能掛上述行為。

### 6.2 報告語意與損耗申報

- **已修復清單** = before 有、after 無的 findings(以前後 lint diff 計算,不依賴 Fixable 旗標——✓ˢ 類因此正確歸入已修復)。
- **仍需手動清單** = before 有、after 仍在的 findings(段落過長、標題跳層等內容問題)。
- **申報損耗(封閉清單)**:fix 報告逐項列出以下類別,**FixProfile 的目標行為(上表)不算損耗**:
  1. 未知節點(Raw)被丟棄
  2. subscript/superscript format bits(Markdown 無表達)
  3. element 層 alignment 與 indent
  4. LinkNode 的 rel/target(title 經 `[text](url "title")` 保留)
  5. TableCell 的 colSpan/rowSpan/backgroundColor;首列以外的 headerState
  6. ImageNode 除 Src/AltText 外的欄位
  7. LinkNode 內被丟棄的合規 underline
  8. 未觸發任何 finding 的 style 屬性(如 font-size)
- `--strict`:**申報損耗清單非空即 exit 1、不寫輸出**(FixProfile 目標行為不觸發 strict)。fix 永不原地覆寫輸入檔。
- 輸出:修復後 JSON(`-o` 或 stdout)+ 前後分數對比 + 上述三張清單。

## 7. 階段 4:TUI 改版(儀表板式)

```
┌─ article.json ──────────────────────┐
│ Score 78  結構92 列表85 行內55 …    │   ← 頂部分數列(固定)
├─────────────────────────────────────┤
│ (Markdown 預覽,可捲動)             │   ← 主畫面
├─ Findings (8) ──────────────────────┤
│ error root[7]  heading-skip H2→H4   │   ← 可收合抽屜
└─ q quit · l findings · f fix · s save ┘
```

- 輸入 `.md` 或 `.json` 都走同一條管線(md 先 expose),統一呈現評分 + 預覽。
- **這是版面重寫,不是沿用**:現有 warnings 是「按 `w` 整面替換 preview」,不是抽屜;score 列、同屏 findings 抽屜皆為新佈局。互動鍵位風格延續現有(`l` 切抽屜、`?` help)。
- `f` 在記憶體執行 auto-fix 並即時更新分數(顯示 before→after)。**模型需重構為可變:現有 model 欄位在 newModel 一次算定,需改為持有 Document 並可重跑 lint/render 管線。**
- `s` 儲存/匯出:**新增格式選擇(md / json)**;存 json 依賴 §5.1 serializer。現有 save 只寫 preview 顯示行,需擴充。
- 移除 JSON tree 主面板;tree 檢視由既有 `--dump-tree` 承擔。

## 8. 測試策略

- **規則單元測試**:每條規則正反例(最小 doc 構造),含 §4.2 名詞定義的邊界(run 中斷、巢狀列表包裝項、句界誤判詞)。
- **score_test**:扣分、權重、邊界(0 分下限、空文件)。
- **Golden tests**:`testdata/lint/` 滿分文件(由撰寫指南轉出,驗收 ≥95 / 0 error)與違規 fixture,鎖 text 與 JSON 輸出;`testdata/mdimport/` md→json golden;fix 前後 round-trip golden。
- **性質測試**:`expose(render(expose(md)))` 收斂(第二次 round-trip 不再變化;前提:§5.1 renderer highlight 輸出);fix 後「已修復清單」中的 rule 在 after-lint 必為 0;fix after 分數 ≥ before。
- **Serializer round-trip**:`parse(serialize(parse(json)))` 對所有 KnownType fixture 結構相等。
- TUI 比照現有 `model_test.go`。

## 9. 風險與緩解

| 風險 | 緩解 |
|---|---|
| goldmark AST 與 Lexical 結構不對齊(巢狀列表、loose/tight list、blockquote 多段落) | mdimport golden tests 覆蓋巢狀/邊界;quote 正準形狀已定(§5.2);以現有 renderer 輸出為 round-trip 基準 |
| serializer 多型欄位(format)與 omit-empty 出錯 | serializer round-trip 性質測試(§8) |
| emoji rune-range 表誤判(ZWJ、旗幟、變體) | 偵測器獨立 + 專屬測試表;誤判僅影響 emoji 類(權重 0.10) |
| CSS 顏色解析覆蓋不全 | 僅支援 hex/rgb(),其餘明確跳過不報(避免 false positive) |
| fix 的 round-trip 損耗超出預期 | §6.2 封閉損耗清單 + `--strict`;fix 永不原地覆寫 |
| 規則門檻引發爭論 | 門檻引用自最佳實踐報告;**屬本 spec 操作化推定者(80-rune、表格儲存格白名單放寬、>6 句寬鬆端、低對比統一 4.5:1)在 `--rules` 中明確標註** |

## 10. 訊息語言

Rule ID 為英文 kebab-case;message/suggestion 一律英文(與 OSS repo 一致);expose 產生的 callout 標籤亦為英文(§5.2);中文規則說明留在 docs。
