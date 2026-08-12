# Lexical 排版 Tech Article 最佳實踐報告
## 在 Meta Lexical JSON 可實現的前提下,讓 Guideline / SOP / Rules 類文章獲得最佳讀者體驗

> 研究日期: 2026-06-11 | 來源數量: 35(官方源占比 74%)| 模式: Standard(5 任務、多代理並行)| AS_OF: 2026-06-11

---

## 摘要 / Executive Summary

本報告回答一個工程上很具體的問題:**在 Meta Lexical 序列化 JSON(SerializedEditorState)能表達的範圍內,一篇技術文章(guideline、SOP、rules 等)應該如何排版,才能取得最佳讀者體驗?**

研究分五條線並行:(1) 盤點 Lexical JSON 的真實表達能力(官方節點、format bitmask、style 字串);(2) 讀者閱讀行為與宏觀結構(NN/g 眼動研究、Google / Microsoft 官方文件風格指南);(3) 行內強調與排版學(粗體/斜體/底線/全大寫);(4) 用色與 highlight 的可及性規範(WCAG 2.2);(5) Emoji 與主流編輯器生態(GitHub / Notion / Confluence / MkDocs / Payload CMS)的語意區塊慣例。

核心結論:

1. **讀者最多只讀一頁約 20–28% 的文字,79% 的人以掃讀為主**[8][9]——因此排版的第一目標不是「美」,而是**可掃描性(scannability)**:答案前置、標題分層、列表化、段落短小。
2. Lexical 的文字格式是 bitmask(bold=1、italic=2、…、highlight=128),顏色與底色則放在自由的 `style` CSS 字串中[1][2]——**技術上「什麼顏色都能做」,正因如此才需要人為紀律**:Google 風格指南明令不要覆寫全域顏色樣式[20]。
3. **強調是稀缺資源**:粗體與斜體互斥、越少用越有效[17];底線只留給連結[20];全大寫會讓閱讀速度下降 10–20%[22]。
4. **顏色永遠不能是唯一的訊息載體**(WCAG 1.4.1),文字對比至少 4.5:1(WCAG 1.4.3)[18][19];語意色(藍=info、綠=tip、黃/橘=warning、紅=danger)是跨平台共識,但必須搭配圖示或文字標籤[19][29][30][33]。
5. **Emoji 是雙面刃**:作為段落/狀態錨點(✅ ❌ ⚠️)可提升掃讀效率[26],但螢幕報讀器會逐字念出 Unicode 名稱,絕不可單獨承載關鍵語意[23];GitLab 甚至在技術文件中全面禁用[25]。
6. Lexical 核心**沒有** callout/admonition、摺疊、多欄、圖片節點——這些都要自訂節點[4][34];在「純官方節點」前提下,最佳替代是 **Quote 區塊 + 粗體語意標籤 + 單一 emoji 圖示**的慣例組合(模式詳見第 6 節)。

第 7 節給出一張可直接執行的 Cheat Sheet,把每個 Lexical 排版元素對應到「何時用、怎麼用、用多少」的規則。

---

## 目錄

1. [Lexical JSON 排版能力盤點](#1-lexical-json-排版能力盤點你手上有哪些工具)
2. [宏觀結構:讀者怎麼讀,文章就怎麼分層](#2-宏觀結構讀者怎麼讀文章就怎麼分層)
3. [行內強調:粗體、斜體、底線、行內 Code、全大寫](#3-行內強調粗體斜體底線行內-code全大寫)
4. [用色與 Highlight](#4-用色與-highlight)
5. [Emoji 使用準則](#5-emoji-使用準則)
6. [語意區塊(Callout)模式:Lexical 限制下的實作](#6-語意區塊callout模式lexical-限制下的實作)
7. [最佳實踐速查表(Cheat Sheet)](#7-最佳實踐速查表cheat-sheet)
8. [核心爭議(Counter-Review)](#8-核心爭議-counter-review)
9. [關鍵發現](#9-關鍵發現--key-findings)
10. [局限性與未來方向](#10-局限性與未來方向)
11. [參考文獻](#11-參考文獻--references)

---

## 1. Lexical JSON 排版能力盤點(你手上有哪些工具)

任何「最佳實踐」都必須先確認工具箱裡有什麼。Lexical 是 Meta 開源的編輯器框架,實際驅動 Facebook、Messenger、WhatsApp、Instagram、Workplace 的文字編輯,也是 Payload CMS 的預設富文字編輯器[35]。

### 1.1 文件骨架

序列化後的文件是一棵以 `root` 為根的節點樹;所有 element 節點帶有 `children`、`format`、`indent`、`direction` 等欄位,文字葉節點額外帶 `format`(bitmask)、`style`(CSS 字串)、`mode`、`detail`[2][3]。

### 1.2 行內格式:format bitmask

`TextNode.format` 是位元遮罩,可任意疊加[1]:

| 常數 | 值 | 視覺效果 |
|---|---|---|
| IS_BOLD | 1 | **粗體** |
| IS_ITALIC | 2 | *斜體* |
| IS_STRIKETHROUGH | 4 | ~~刪除線~~ |
| IS_UNDERLINE | 8 | 底線 |
| IS_CODE | 16 | `行內代碼` |
| IS_SUBSCRIPT / IS_SUPERSCRIPT | 32 / 64 | 下標 / 上標 |
| IS_HIGHLIGHT | 128 | 螢光標記 |
| IS_LOWERCASE / IS_UPPERCASE / IS_CAPITALIZE | 256 / 512 / 1024 | 大小寫變換 |

例如 `"format": 3` = 粗體+斜體;`"format": 129` = 粗體+highlight。

**注意**:element 節點上的 `format` 是「對齊」的整數枚舉(left=1、center=2、right=3、justify=4、start=5、end=6),**不是** bitmask——同一個欄位名在 text 與 element 上語意完全不同,是序列化消費端常見的誤解來源[1]。

### 1.3 顏色:style 字串

文字色與底色**不在 bitmask 裡**,而是寫進 `style` 欄位的原始 CSS 字串(如 `"color: #b91c1c; background-color: #fef9c3;"`),playground 工具列的 font color / background color 就是這樣實作的[2]。這代表 Lexical 對顏色毫無限制——紀律必須由內容規範提供(見第 4 節)。

### 1.4 官方區塊節點

來自官方 `@lexical/*` 套件、可直接序列化的區塊[4][5][6][7]:

- **HeadingNode**(h1–h6,`tag` 屬性)與 **QuoteNode**[5]
- **ListNode / ListItemNode**:`listType` 為 `bullet` | `number` | `check`;checklist 的 `checked` 布林值是一級公民,直接序列化[6]
- **CodeNode**(`language` 屬性)+ **CodeHighlightNode**(語法 token)[4]
- **TableNode / TableRowNode / TableCellNode**:支援 colSpan、rowSpan、表頭狀態、儲存格 backgroundColor、凍結列/欄、隔行條紋[4]
- **LinkNode / AutoLinkNode**、**HashtagNode**、**HorizontalRuleNode**[4]
- **MarkNode**(`ids: string[]`):官方的註解/highlight 錨點機制,一段文字可同時掛多個註解 ID[7]

### 1.5 沒有的東西(必須自訂節點)

Playground 原始碼證實,以下常見元素**都不在任何官方 `@lexical/*` 套件中**,playground 是用自訂節點示範的:圖片、callout/admonition、摺疊區塊(CollapsibleContainerNode 是 playground 專屬,帶 `open` 布林)、多欄 layout、嵌入媒體(YouTube/Figma/Excalidraw)、數學公式[4]。Payload CMS 的官方功能清單同樣不含 callout 與摺疊,留給其 BlocksFeature 在 CMS schema 層解決[34]。

> 本報告其餘章節的所有建議,都以「官方節點 + style 字串」為邊界;需要自訂節點的方案會明確標註。

**置信度:High。** 依據:全部來自 facebook/lexical 原始碼與 lexical.dev 官方文件,屬一手資料。
**反方解釋:** Meta 自家產品(如 WhatsApp)實際啟用的節點子集並未公開[35],生產環境的能力邊界可能比 playground 展示的更窄。

---

## 2. 宏觀結構:讀者怎麼讀,文章就怎麼分層

### 2.1 讀者行為的硬數據

- 一次造訪中,使用者最多讀 28% 的頁面文字,常態約 20%[9]。
- NN/g 對 232 名使用者的眼動研究確立了 F 型掃讀:兩道水平掃視 + 左側垂直掃描;79% 的使用者掃讀新頁面,僅 16% 逐字閱讀[8]。
- 注意力隨段落位置陡降:第 1 段 81% → 第 2 段 71% → 第 3 段 63% → **第 4 段驟降至 32%**[9]。

推論很直接:**重要內容必須在前三段、每行的前兩個詞、頁面左上角出現**[8][13]。

### 2.2 標題層級(Lexical: HeadingNode h1–h6)

- 整篇**只有一個 H1**;層級不可跳號(H1 直接到 H3 是錯的);禁止為了字體大小而濫用標題節點[10]。
- 標題與標題之間必須有過渡文字,不可兩個標題連續出現;1–2 頁的內容通常一層標題就夠[14]。
- 重要關鍵詞放在標題開頭(front-loading),配合 F 型掃讀[13][8]。
- 實務上 tech article 建議只用 **H1(篇名)→ H2(章)→ H3(節)** 三層;H4 以下在 Lexical 雖可表達[5],但已超出多數讀者的掃讀解析度(此為本報告的綜合建議)。

### 2.3 段落(ParagraphNode)

- 一段一個想法;**最多 5–6 句**;允許單句成段[12]。
- Microsoft 給的視覺尺度是 **3–7 行**[13];Google 給的句長上限是 **26 個英文詞**[12]。
- **倒金字塔**:每段最重要的資訊放第一句,絕不把重點藏在段尾[12][16]。

### 2.4 列表(ListNode:bullet / number / check)

- **有順序的步驟用編號列表;無順序的集合用項目符號列表**[11][15]——這是 SOP 與 guideline 最重要的單一規則。
- 每個列表前要有一個完整句子作引言(不可用殘句接列表補完);所有項目保持平行的文法結構[11]。
- 列表至少 2 項、**理想不超過 7 項**;每項要短到讀者一眼能看到 2–3 項[15]。
- **程序超過 12 步就該拆段**(子標題分塊)[13]。
- Lexical 的 checklist(`listType: "check"`)是官方一級功能[6]——rules/checklist 類內容應直接用它,而不是用「☑️ + bullet」模擬。

### 2.5 表格(TableNode)

表格用於「可枚舉的對照事實」(參數表、權限矩陣、版本對照)。Lexical 表格支援表頭列、儲存格底色、凍結列欄[4],足以表達文件級表格;但說明性內容仍應放在表格前後的段落,而非塞進儲存格(綜合 [13][15] 的掃讀原則)。

### 2.6 三種文體的版型模板

| 文體 | 結構骨架 | 關鍵規則 |
|---|---|---|
| **Guideline** | 倒金字塔:結論/原則 → 理由 → 細節[16] | 前三段給出全部核心主張[9] |
| **SOP** | 前置摘要 + 前置條件 → 編號步驟(≤12 步/塊)[13][11] | 倒金字塔**不適用**於步驟本體,步驟依時序排列[16] |
| **Rules** | 規則條目用 checklist 或短編號條文[6][15] | 每條規則一句話;例外與罰則用縮排子項 |

**置信度:High。** 依據:NN/g 眼動數據與 Google/Microsoft 官方風格指南高度一致,數字互相印證。
**反方解釋:** 有研究者主張 F 型掃讀是「內容寫得差」的症狀而非閱讀天性——結構良好、視覺層次強的內容可以誘導更完整的閱讀[8];因此 F-pattern 應視為「未優化時的底線行為」,而非版面設計的硬性處方。

---

## 3. 行內強調:粗體、斜體、底線、行內 Code、全大寫

### 3.1 總原則:強調是稀缺資源

粗體與斜體**互斥使用、越少越好**——「如果所有東西都被強調,就沒有東西被強調」;兩者只適合短語,長段落套用粗/斜體會降低可讀性並使讀者疲勞[17]。

### 3.2 分工(Lexical: format 1 / 2 / 8 / 16)

| 格式 | 用途 | 來源 |
|---|---|---|
| **粗體**(format 1) | UI 元素名稱、run-in 標題、單一關鍵句;不作一般語氣強調 | Google 將粗體嚴格保留給 UI 元素與 run-in 標題[20] |
| *斜體*(format 2) | 首次引入的新術語(僅第一次)、作品標題、word-as-word、極少量語氣強調 | Google:「通常你的文字本身就能承載強調,不必加斜體」[20];Microsoft:斜體強調要 sparingly,新術語只在首次出現時斜體[21] |
| 底線(format 8) | **只保留給連結**,正文一律不用 | Google 明文:「Reserve underlining for link text」[20] |
| `行內 code`(format 16) | 指令、檔名、參數、代碼識別符 | Lexical 一級支援[1];與粗體/斜體職責互不重疊 |
| 刪除線(format 4) | 僅用於明示「已廢止的舊規則」並保留歷史脈絡 | 本報告綜合建議(風格指南未涵蓋) |

中文語境補充(本報告推論):中文無斜體傳統、CJK 字形斜體效果差,建議把「斜體職責」改由「引號 + 行內 code」承擔,斜體僅用於西文術語。

### 3.3 全大寫(format 512 IS_UPPERCASE)

避免在連續文字使用:全大寫使閱讀速度下降 10–20%(Tinker, 1955),2019 年研究確認小寫快 13%,讀寫障礙者額外慢 13–18%,55 歲以上讀者對全大寫條款的誤解率高 29%,且螢幕報讀器可能逐字母拼讀[22]。Google 只允許 placeholder 使用全大寫[20];Microsoft 的例外清單同樣極短(如連接埠名稱)[21]。

**置信度:High**(規則本身);**Medium**(精確頻率上限——見爭議)。依據:Butterick、Google、Microsoft 三方獨立一致;全大寫有量化研究支撐。
**反方解釋:** 對 1–3 個詞的短標籤(如 "WARNING"),全大寫的字形辨識懲罰可能不成立——Tinker 的數據針對連續文字,並未隔離單詞場景[22],因此「callout 標籤用全大寫」仍是可接受慣例(GitHub alerts 即如此[29])。

---

## 4. 用色與 Highlight

### 4.1 兩條 WCAG 硬規則(不可協商)

1. **對比**(SC 1.4.3):一般文字對背景至少 **4.5:1**,大字(18pt+,或 14pt+ 粗體)至少 **3:1**;數值不得四捨五入,4.499:1 即不合格[18]。對比以亮度(luminance)計算,色相不同但亮度相近的組合(如白底黃字)照樣不合格[18]。
2. **顏色不得是唯一訊息載體**(SC 1.4.1):只用紅色標示「禁止事項」、只用顏色區分連結,都不合格;必須搭配圖示、文字標籤或粗細等第二指標[19]。兩色之間若有 ≥3:1 的對比差,可算作一種額外的視覺區分[19]。

### 4.2 文字色(style: color)

Google 風格指南**明文禁止以行內樣式覆寫全域字色**[20]。Lexical 的 `style` 欄位雖然允許任意 CSS[2],最佳實踐是:**正文一律不手動改字色**;唯一例外是經過定義的語意色系統(見 4.4),且必須通過 4.5:1 對比[18]。

### 4.3 Highlight(format 128,或 style background-color)

值得注意的研究空白:**沒有任何主流風格指南(Google/Microsoft/Butterick)直接規範行內螢光標記**——可依循的只有 WCAG 對「文字 × 底色」組合的對比要求[18]。本報告的綜合建議:

- 優先使用 `IS_HIGHLIGHT`(format 128)的單一預設黃色,而非 style 自訂底色——語意上等同 HTML `<mark>`,且可被轉換器(如 Markdown `==…==`)無損對應[1]。
- 每篇文章 highlight 不超過 3–5 處,只標「讀者離開前必須記住的那句話」;依據是強調稀缺性原則[17]與掃讀錨點理論[8]。
- highlight 之上的文字仍須滿足 4.5:1 對比[18]。

### 4.4 語意色系統(跨平台共識)

主流文件平台對「語意 → 顏色」的對應高度收斂[29][30][32][33]:

| 語意 | 顏色 | GitHub Alerts[29][30] | Confluence[33] | MkDocs Material[32] |
|---|---|---|---|---|
| 資訊 / Note | 藍 | `[!NOTE]` 藍 | Info(藍) | note/info(藍) |
| 建議 / Tip | 綠 | `[!TIP]` 綠 | Tip(綠) | tip/success(綠) |
| 重要 / Important | 紫 | `[!IMPORTANT]` 紫 | — | — |
| 注意 / Warning | 黃/橘 | `[!WARNING]` 琥珀 | Note(黃) | warning(橘) |
| 危險 / Caution | 紅 | `[!CAUTION]` 紅 | Warning(紅) | danger/failure(紅) |

這套對應是社群慣例而非單一規範標準;且依 WCAG 1.4.1,**每個語意色都必須搭配圖示或標籤文字**才合格[19]。

**置信度:High**(WCAG 規則);**Medium**(highlight 用量建議,屬推論)。
**反方解釋:** 語意色「紅=危險」依賴文化慣例,並無單一 normative 標準收編所有平台;Notion 官方甚至不對任何顏色賦予語意[31]——顏色語意是「平台各自的約定」,跨平台遷移時不保證讀者解讀一致。

---

## 5. Emoji 使用準則

### 5.1 為什麼要節制:可及性成本

- 螢幕報讀器會逐一念出 Unicode 官方名稱:🙏 被念成「雙手合十」而非「謝謝」,🚩 是「插在柱上的三角旗」——作者意圖經常完全流失[23]。
- 連發 emoji(🔥🔥🔥)會被逐顆念出("fire fire fire"),造成嚴重聽覺疲勞[28];膚色修飾符讓每次播報更冗長[28]。
- **句中** emoji 會把 Unicode 名稱嵌進句子文法,直接破壞理解;修復手段是 `aria-hidden="true"` + 相鄰真實文字[23]——但 Lexical JSON 層沒有 aria 概念,這必須由渲染端處理(本報告推論)。
- 同一顆 emoji 在 Apple(擬真)、Google(扁平)、Microsoft(3D)的字形差異可能改變語氣與含義[27];Microsoft 風格指南要求只用廣為人知的 emoji、發佈前到 Emojipedia 核對各平台外觀[24]。

### 5.2 兩極的業界立場

- **禁用派**:GitLab 官方文件風格指南在技術文件中全面禁止 emoji shortcode,要求改用 SVG 圖示;emoji 只允許出現在 handbook/wiki 情境[25]。
- **節制使用派**:在內部流程文件中,作為**段落錨點或狀態指示**(✅ ❌ ⚠️)的 emoji 可提升掃讀效率;但段落中的裝飾性重複會增加訊息解碼時間與歧義[26]。

### 5.3 本報告的執行規則(綜合)

1. **位置白名單**:只出現在 (a) H2/H3 標題行首作章節錨點、(b) 表格狀態欄(✅/❌/⚠️)、(c) callout 標籤前。**永不出現在句子中間**[23][26]。
2. **一處一顆**,全文 emoji 詞彙表不超過 5–7 種,且每種固定一個意義(本篇內保持一對一映射)。
3. **emoji 永不單獨承載語意**:✅ 旁必須有「通過/允許」字樣,否則報讀器只會念「打勾按鈕」[23][19]。
4. 只用 Unicode 老牌、跨平台歧義低的符號(✅ ❌ ⚠️ 💡 📌 🚫),發佈前核對多平台渲染[24][27]。
5. 面向外部/正式規範文件(rules、合規 SOP):比照 GitLab,**整篇禁用**,語意改由文字標籤承擔[25]。

**置信度:Medium-High。** 依據:可及性機制(報讀行為)有一致的多源證據;「emoji 提升掃讀」一側僅有業界經驗,缺對照實驗。
**反方解釋:** 有實務者主張在內部工具中,✅ 類前綴對明眼使用者的任務完成速度有實質幫助,且 ARIA 緩解後報讀成本可控——此立場目前沒有對照研究支持,但也未被證偽[26]。

---

## 6. 語意區塊(Callout)模式:Lexical 限制下的實作

技術文章最需要的「Note / Tip / Warning」區塊,在 Lexical 官方節點中**不存在**[4][34]。各平台的原生做法——Notion callout(圖示+十色底)[31]、Confluence 四種固定 panel[33]、MkDocs 十二種 admonition 與 `???` 摺疊[32]、GitHub 以 blockquote 擴充的 `> [!NOTE]` 語法[29]——全都依賴自家擴充。可行路徑有三:

### 路徑 A:純官方節點(零自訂,最大可攜性)——推薦預設

模仿 GitHub 的思路(GitHub alerts 本質就是「帶語意標記的 blockquote」[29]):

```
QuoteNode
 └─ ParagraphNode
     ├─ TextNode "⚠️ "                ← 單一圖示 emoji(可選)
     ├─ TextNode "注意:" format:1     ← 粗體語意標籤(必須)
     └─ TextNode "本操作不可回復,執行前先備份。"
```

語意由**粗體文字標籤**承載(滿足 WCAG 1.4.1「不靠顏色」[19]),emoji 僅是輔助錨點(規則見第 5 節)。固定標籤詞彙表比照五級制:備註 / 提示 / 重要 / 注意 / 危險[29][30]。

### 路徑 B:官方節點 + style 底色(中等可攜性)

在路徑 A 之上,對 Quote 內文字加 `style: "background-color: …"`[2],顏色按第 4.4 節語意表取用,並維持 4.5:1 對比[18]。代價:顏色寫死在內容 JSON 裡,換主題/深色模式時會失真——這正是 Notion 把調色盤鎖死在 10 個具名顏色、不開放 hex 的原因[31]。**慎用。**

### 路徑 C:自訂 CalloutNode(最佳體驗,綁定自家渲染器)

照 playground 的 CollapsibleContainerNode 模式自訂節點(序列化僅多帶一個屬性,如 `calloutType: "warning"`)[4],渲染端再決定顏色與圖示;Payload 則建議經 BlocksFeature 在 schema 層定義[34]。語意進 JSON、樣式留渲染端,是長期最乾淨的架構——但 JSON 一旦離開自家渲染器就需要 fallback。

**摺疊內容**同理:官方無 toggle 節點[4];純官方節點下用「H3 + 內容」替代,MkDocs 的 `???` 摺疊 admonition 證明摺疊只是 callout 的變體而非獨立結構[32]。

**置信度:High**(能力邊界);**Medium**(路徑 A 為最佳預設,屬架構判斷)。
**反方解釋:** 若文章只在自家產品內渲染、永不外流,路徑 C 直接勝出,路徑 A 的「可攜性」優勢就不成立。

---

## 7. 最佳實踐速查表(Cheat Sheet)

### 7.1 結構層

| 元素 | Lexical 實現 | 規則 | 上限/警示 |
|---|---|---|---|
| 篇名 | HeadingNode h1 | 全篇唯一[10] | 1 個 |
| 章/節 | h2 / h3,`tag` 屬性[5] | 不跳層;關鍵詞前置[10][13];標題間必有正文[14] | 建議止於 h3 |
| 段落 | ParagraphNode | 一段一想法;重點放第一句[12] | ≤5–6 句 / 3–7 行[12][13] |
| 句子 | — | 直述、主動語態 | <26 詞[12] |
| 步驟 | ListNode `number`[6] | 時序內容專用;引言用完整句[11] | >12 步拆塊[13] |
| 並列項 | ListNode `bullet`[6] | 無序集合;文法平行[11] | 2–7 項[15] |
| 檢查清單 | ListNode `check` + `checked`[6] | rules/驗收條目專用 | 同上 |
| 對照事實 | TableNode(表頭+條紋)[4] | 只放可枚舉事實,解釋放正文 | — |
| 分隔 | HorizontalRuleNode[4] | 大段落主題切換 | 少用,優先用標題 |

### 7.2 行內層

| 元素 | format 值 | 用途 | 上限 |
|---|---|---|---|
| 粗體 | 1 | UI 名稱、run-in 標題、關鍵句[20] | 每段 ≤1 處(綜合[17]) |
| 斜體 | 2 | 新術語首現、作品名[20][21] | 與粗體互斥[17] |
| 底線 | 8 | **不用**(連結專屬)[20] | 0 |
| 行內 code | 16 | 指令/檔名/參數[1] | 不設限,但不替代強調 |
| Highlight | 128 | 全篇必記要點 | 3–5 處;對比 4.5:1[18] |
| 刪除線 | 4 | 已廢止條款存檔 | 罕用 |
| 全大寫 | 512 | 避免;僅 1–3 詞標籤可容忍[22] | ~0 |
| 字色/底色 | style 字串[2] | 僅限語意色系統;勿覆寫正文色[20] | 必配圖示/標籤[19] |

### 7.3 語意區塊五級制(路徑 A 寫法)

| 級別 | 標籤(粗體) | 圖示 | 對應色(僅渲染端) |
|---|---|---|---|
| Note | **備註:** | 📌 | 藍[30][33] |
| Tip | **提示:** | 💡 | 綠[30][33] |
| Important | **重要:** | ⭐ | 紫[30] |
| Warning | **注意:** | ⚠️ | 黃/橘[30][33] |
| Caution | **危險:** | 🚫 | 紅[30][33] |

### 7.4 發佈前檢查清單

- [ ] 前三段是否已交付全文最重要的結論?[9][16]
- [ ] 是否只有一個 H1、層級無跳號、標題間有正文?[10][14]
- [ ] 所有時序步驟都是編號列表、無序集合都是 bullet?[11]
- [ ] 任一列表 >7 項、任一程序 >12 步?拆。[15][13]
- [ ] 粗體/斜體是否互斥且稀少?底線是否為 0?[17][20]
- [ ] 任何顏色語意是否都有圖示/文字標籤備援?對比 ≥4.5:1?[19][18]
- [ ] emoji 是否只在標題首/狀態欄/callout 標籤,且不單獨承載語意?[23][26]
- [ ] highlight 是否 ≤5 處?[17]（綜合建議)

---

## 8. 核心爭議 (Counter-Review)

P6 反向審查發現的主要張力,讀者採用本報告建議時應知悉:

- **爭議 1 — F-pattern 是描述還是處方:** NN/g 數據描述的是「使用者面對未優化內容」的行為;有觀點認為強視覺層次可以打破 F 型、誘導完整閱讀[8]。本報告把 F-pattern 當作「底線假設」而非目標。
- **爭議 2 — Emoji 禁用 vs 節制使用:** GitLab 在技術文件全面禁 emoji[25],與「狀態錨點提升掃讀」的業界經驗[26]直接衝突;後者缺乏對照研究。本報告以「文件正式程度」作為裁決軸:外部規範禁用、內部指南白名單節制使用。
- **爭議 3 — 全大寫禁令的邊界:** 10–20% 減速數據來自連續文字研究,未隔離 1–3 詞短標籤場景[22];GitHub alerts 的 NOTE/WARNING 標籤本身就是全大寫[29]。短標籤豁免是合理推論而非實證結論。
- **爭議 4 — Highlight 規則是推論不是引用:** 沒有任何主流風格指南直接規範行內螢光標記;本報告的「3–5 處」上限是從強調稀缺原則[17]外推的,無直接證據。
- **爭議 5 — 顏色語意的普適性:** 五級語意色在 GitHub/Confluence/MkDocs 收斂[29][32][33],但 Notion 官方拒絕賦予顏色語意[31];不存在單一 normative 標準,跨文化讀者的解讀亦未驗證。

---

## 9. 關鍵發現 / Key Findings

- **發現 1:** 排版的支配性目標是可掃描性——讀者常態只讀 ~20% 文字、第 4 段注意力剩 32%,「答案前置 + 標題分層 + 列表化」比任何視覺修飾都重要。[8][9][12]
- **發現 2:** Lexical JSON 的能力邊界清晰:行內 11 種 bitmask 格式 + 自由 style 字串 + 9 類官方區塊節點;callout/摺疊/圖片一律需要自訂節點。[1][2][4][34]
- **發現 3:** 行內強調的鐵律是稀缺與分工:粗體=UI/關鍵句、斜體=術語首現、底線=0、全大寫≈0;兩大廠風格指南與排版學專著三方一致。[17][20][21][22]
- **發現 4:** 顏色是「渲染端的事」:WCAG 要求顏色不得單獨承載語意且對比 ≥4.5:1;最穩健的架構是語意進 JSON(標籤/節點型別)、顏色留給主題。[18][19][20]
- **發現 5:** 在純官方節點前提下,「Quote + 粗體語意標籤 + 單一 emoji」是 callout 的最佳替代——這正是 GitHub alerts 在 Markdown 上走過的同一條路。[29][19][23]

---

## 10. 局限性與未來方向

### 本研究局限

- **無第一手中文閱讀研究**:NN/g 眼動數據與句長上限(26 詞)皆基於英文;中文掃讀行為、CJK 斜體替代方案部分為本報告推論。
- **Highlight 與 emoji 密度缺乏對照實驗**:相關建議(3–5 處、5–7 種)是從強調稀缺原則外推的專家式判斷。
- **Meta 生產環境 schema 不公開**:Facebook/WhatsApp 實際啟用的 Lexical 節點子集無法驗證[35]。
- **部分來源不可達**:GOV.UK 細則頁與 Readability Guidelines 專頁抓取失敗,已從引用中剔除而非帶病引用。
- 來源以業界實踐(practitioner)為主,經同行評審的學術文獻佔比低。

### 未來方向

1. 以使用者自己的 lexical-cli 渲染管線做 A/B 掃讀測試,驗證 highlight 密度與 emoji 錨點對中文讀者的實效(優先級:高)。
2. 若文章長期只在自家渲染器呈現,評估實作 `CalloutNode`(路徑 C)並設計 Markdown fallback(`> [!NOTE]` 語法天然對應[29])。
3. 追蹤 @lexical/table 凍結列欄等新特性的版本門檻,確認最低相容版本[4]。

---

## 11. 參考文獻 / References

[1] Meta. "LexicalConstants.ts (facebook/lexical source)". Source-Type: official. As Of: 2025-06. https://raw.githubusercontent.com/facebook/lexical/main/packages/lexical/src/LexicalConstants.ts
[2] Meta. "LexicalTextNode.ts (facebook/lexical source)". Source-Type: official. As Of: 2025-06. https://github.com/facebook/lexical/blob/main/packages/lexical/src/nodes/LexicalTextNode.ts
[3] Meta. "Nodes — Lexical Documentation". Source-Type: official. As Of: 2025-06. https://lexical.dev/docs/concepts/nodes
[4] Meta. "PlaygroundNodes.ts (facebook/lexical source)". Source-Type: official. As Of: 2025-06. https://github.com/facebook/lexical/blob/main/packages/lexical-playground/src/nodes/PlaygroundNodes.ts
[5] Meta. "@lexical/rich-text API". Source-Type: official. As Of: 2025-06. https://lexical.dev/docs/api/modules/lexical_rich-text
[6] Meta. "@lexical/list API". Source-Type: official. As Of: 2025-06. https://lexical.dev/docs/api/modules/lexical_list
[7] Meta. "MarkNode.ts (facebook/lexical source)". Source-Type: official. As Of: 2025-06. https://raw.githubusercontent.com/facebook/lexical/main/packages/lexical-mark/src/MarkNode.ts
[8] Nielsen Norman Group. "F-Shaped Pattern For Reading Web Content". Source-Type: official. As Of: 2006. https://www.nngroup.com/articles/f-shaped-pattern-reading-web-content-discovered/
[9] Nielsen Norman Group. "Website Reading: It (Sometimes) Does Happen". Source-Type: official. As Of: 2013. https://www.nngroup.com/articles/website-reading/
[10] Google. "Headings — developer documentation style guide". Source-Type: official. As Of: 2024. https://developers.google.com/style/headings
[11] Google. "Lists — developer documentation style guide". Source-Type: official. As Of: 2024. https://developers.google.com/style/lists
[12] Google. "Paragraph structure — developer documentation style guide". Source-Type: official. As Of: 2024. https://developers.google.com/style/paragraph-structure
[13] Microsoft. "Scannable content — Writing Style Guide". Source-Type: official. As Of: 2023-06. https://learn.microsoft.com/en-us/style-guide/scannable-content/
[14] Microsoft. "Headings — Writing Style Guide". Source-Type: official. As Of: 2018-07. https://learn.microsoft.com/en-us/style-guide/scannable-content/headings
[15] Microsoft. "Lists — Writing Style Guide". Source-Type: official. As Of: 2023-06. https://learn.microsoft.com/en-us/style-guide/scannable-content/lists
[16] Veeam. "Inverted Pyramid — Technical Writing Style Guide". Source-Type: secondary-industry. As Of: 2024. https://helpcenter.veeam.com/docs/styleguide/tw/inverted_pyramid.html
[17] Matthew Butterick. "Bold or italic — Practical Typography". Source-Type: secondary-industry. As Of: 2024. https://practicaltypography.com/bold-or-italic.html
[18] W3C WAI. "Understanding SC 1.4.3: Contrast (Minimum)". Source-Type: official. As Of: 2023-10. https://www.w3.org/WAI/WCAG22/Understanding/contrast-minimum.html
[19] W3C WAI. "Understanding SC 1.4.1: Use of Color". Source-Type: official. As Of: 2023-10. https://www.w3.org/WAI/WCAG22/Understanding/use-of-color.html
[20] Google. "Text-formatting summary — developer documentation style guide". Source-Type: official. As Of: 2025. https://developers.google.com/style/text-formatting
[21] Microsoft. "Formatting common text elements — Style Guide". Source-Type: official. As Of: 2025-03. https://learn.microsoft.com/en-us/style-guide/text-formatting/formatting-common-text-elements
[22] Brickfield Education Labs. "Why ALL CAPS Text Creates Reading Barriers". Source-Type: secondary-industry. As Of: 2026-03. https://brickfield.ie/2026/03/26/why-all-caps-text-creates-reading-barriers/
[23] Pope Tech. "Making Emojis and Icons Screen Reader Accessible". Source-Type: secondary-industry. As Of: 2026-04. https://blog.pope.tech/2026/04/01/making-emojis-and-icons-screen-reader-accessible/
[24] Microsoft. "emoji, emoticons — Style Guide A–Z". Source-Type: official. As Of: 2024. https://learn.microsoft.com/en-us/style-guide/a-z-word-list-term-collections/e/emoji-emoticons
[25] GitLab. "Documentation Style Guide". Source-Type: official. As Of: 2024. https://docs.gitlab.com/development/documentation/styleguide/
[26] Process Street. "Emoji Aren't Just For Fun: Using Emoji in Business Documents". Source-Type: secondary-industry. As Of: 2022. https://www.process.st/emoji-in-business-documents/
[27] Emojifyi. "How Emoji Rendering Works Across Platforms". Source-Type: secondary-industry. As Of: 2024. https://emojifyi.com/stories/emoji-rendering-across-platforms/
[28] Bureau of Internet Accessibility. "Emojis and Web Accessibility: Best Practices". Source-Type: secondary-industry. As Of: 2023. https://www.boia.org/blog/emojis-and-web-accessibility-best-practices
[29] GitHub. "Markdown Alert Syntax — Community Discussion #16925". Source-Type: official. As Of: 2023-12. https://github.com/orgs/community/discussions/16925
[30] ShowMeMyMD. "GitHub Markdown Callouts Guide". Source-Type: secondary-industry. As Of: 2024-06. https://www.showmemymd.com/blog/github-callouts-guide
[31] super.so. "Notion Callout Block — The Complete Guide". Source-Type: secondary-industry. As Of: 2024-01. https://super.so/blog/notion-callout-block-the-complete-guide-2024
[32] squidfunk. "Admonitions — Material for MkDocs". Source-Type: official. As Of: 2025-01. https://squidfunk.github.io/mkdocs-material/reference/admonitions/
[33] Atlassian. "Insert the Info, Tip, Note, and Warning macros — Confluence Cloud". Source-Type: official. As Of: 2024-09. https://support.atlassian.com/confluence-cloud/docs/insert-the-info-tip-note-and-warning-macros/
[34] Payload CMS. "Official Features — Rich Text". Source-Type: official. As Of: 2025-06. https://payloadcms.com/docs/rich-text/official-features
[35] Meta. "facebook/lexical — GitHub README". Source-Type: official. As Of: 2025-06. https://github.com/facebook/lexical
