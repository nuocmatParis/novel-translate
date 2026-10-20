# novel-translate-cli

CLI và TUI dịch light novel **Nhật → Việt**, nhận EPUB và xuất EPUB. Mỗi workspace quản lý một bộ truyện, gồm nhiều tập dùng chung quy ước, thuật ngữ và lịch sử xưng hô theo thời điểm của truyện.

Luồng chính: **nhập EPUB → dịch → review/sửa → xử lý chữ trong ảnh → kiểm tra → xuất EPUB**. Hai chương logic mỗi đợt; hệ thống tự chạy tiếp và lưu kết quả theo khối. Một chương dài có thể cần nhiều request API.

Phần lõi có kiểm thử offline và kiểm tra EPUB bằng EPUBCheck. Chất lượng dịch JP–VN, chi phí toàn tập và tính liên tục xuyên tập vẫn cần nghiệm thu bằng model thật. EPUB hợp lệ không đồng nghĩa với bản dịch đúng nghĩa 100%.

## 1. Cài đặt

| Thành phần | Khi nào cần |
| --- | --- |
| Go **1.25.5+** | Build từ mã nguồn; không cần khi dùng binary |
| API key và model phù hợp | Dịch, review và AI đề xuất quy ước |
| Java và EPUBCheck | Xuất EPUB hoàn chỉnh có kiểm chứng định dạng |
| Bốn font Liberation Serif | Tùy chọn, để nhúng font vào EPUB |

Import, xem tiến độ, xuất ảnh và mở TUI không cần API key. Không cần Python, Docker hay Ollama để chạy CLI với API cloud. Các script nghiệm thu bổ sung dùng Python 3; chạy race test cần cgo và C compiler.

Build trên Windows, trong thư mục repo:

```powershell
go build -trimpath -o novel-translate-cli.exe ./cmd/novel-translate-cli
.\novel-translate-cli.exe --help
```

Linux/macOS:

```sh
# Trong thư mục repo; tự build nếu chưa có binary, rồi mở TUI
sh scripts/run.sh

# Hoặc dùng Make để build lại mỗi lần chạy
make run

# Truyền lệnh CLI qua script
sh scripts/run.sh status --workspace ./workspace/bo-truyen
```

Sau khi cập nhật mã nguồn, chạy `make build` hoặc `go build -trimpath -o novel-translate-cli ./cmd/novel-translate-cli` để thay binary cũ. `scripts/run.sh` dùng lại binary đã có.

Cài vào thư mục người dùng trên Linux/macOS:

```sh
sh scripts/install.sh
"$HOME/.local/bin/novel-translate-cli" tui
```

Script build checkout hiện tại vào `~/.local/bin`; biến `NOVEL_TRANSLATE_INSTALL_DIR` đổi thư mục đích. Gọi bằng đường dẫn trên không cần chỉnh `PATH`. API key lưu trong TUI hoặc `.env` nên không cần `export` lại mỗi lần mở.

Workflow CI tạo artifacts Windows/Linux. Workflow đóng gói được kích hoạt thủ công và tạo snapshot cho Windows/Linux/macOS, amd64/arm64; không tự phát hành GitHub Release. Build chéo chưa chứng minh chạy thực tế trên mọi hệ điều hành.

Các ví dụ sau dùng **PowerShell** và binary trong thư mục hiện tại. Trên Linux/macOS, thay `.\novel-translate-cli.exe` bằng `./novel-translate-cli`. Đường dẫn và tên model trong ví dụ cần thay theo máy/dịch vụ của bạn.

## 2. Mở TUI

```powershell
.\novel-translate-cli.exe tui
```

Chạy binary không có đối số trong terminal cũng mở TUI; `--help` xem trợ giúp CLI. Chọn **Mở workspace** hoặc **Tạo workspace mới**. Nếu thư mục hiện tại có `workspace.yaml`, TUI mở workspace đó. Có thể chỉ định ngay:

```powershell
.\novel-translate-cli.exe tui --workspace .\workspace\bo-truyen --volume vol01
```

Alias: `--tui`. Luồng thiết lập thông thường thực hiện ngay trong TUI:

1. Nhấn **n** tạo workspace, hoặc **o** mở bộ truyện đã có. Nhấn **Ctrl+O** trong ô đường dẫn để duyệt thư mục.
2. Nhấn **a** thêm tập và chọn EPUB Nhật. Tập tiếp theo tự gợi ý số thứ tự và tập trước.
3. Nhấn **m** chọn giao thức, tên model và nhập API key. Key được che khi nhập, lưu vào `.env` của workspace; cấu hình model lưu ở `.translation_work/provider.json`.
4. Nhấn **s** sửa quy ước bộ truyện; **c** chỉnh giới hạn requests, timeout, đường dẫn EPUBCheck và font.
5. Chọn **Nhập EPUB**, rồi **Dịch / tiếp tục**. Nhấn **e** đọc/sửa bản dịch theo chương và đoạn; **Ctrl+S** lưu. Chọn **Review bản dịch** để đối chiếu lại phần đã sửa.
6. Xử lý ảnh, kiểm tra và xuất EPUB bằng các mục trong menu.

Không cần tự tạo file cấu hình hoặc đặt biến môi trường cho luồng này. Thiết lập model nhanh dùng **một model mặc định cho cả ba role** và giữ giới hạn budget cũ; không chuyển các override/fallback riêng từ file cũ. Với cấu hình nhiều model, dùng **Chọn file cấu hình API**. Các thao tác revision, research và scene chi tiết vẫn có CLI.

| Phím | Thao tác |
| --- | --- |
| ↑ / ↓ hoặc j / k | Chọn thao tác |
| ← / → | Chọn tập |
| Enter | Thực hiện; ở biểu mẫu: sang ô tiếp theo/lưu ở ô cuối |
| n / o / a | Tạo workspace / mở workspace / thêm tập, tại màn hình chính |
| m / c / s / e | Model và key / cấu hình chạy / sửa quy ước / sửa bản dịch |
| Tab / Shift+Tab | Đổi ô trong biểu mẫu |
| ← / →, Home / End, Delete / Backspace | Di chuyển con trỏ và sửa nội dung ô nhập |
| Ctrl+U | Xóa nội dung ô nhập |
| Ctrl+O | Duyệt file/thư mục trong ô đường dẫn; Space chọn thư mục đang mở |
| Ctrl+N | Đổi lựa chọn giao thức, API hoặc cách xử lý ảnh |
| Ctrl+S | Lưu trong trình sửa bản dịch/quy ước |
| Tab | Trong trình sửa bản dịch: chuyển giữa nguồn Nhật và bản dịch Việt |
| ? | Xem trợ giúp phím tắt tại màn hình chính |
| r | Làm mới tiến độ |
| PgUp / PgDn | Cuộn nhật ký |
| Esc | Hủy hộp thoại; khi đang chạy thì dừng tác vụ |
| Ctrl+C | Dừng tác vụ; khi rảnh thì thoát |
| q | Tại màn hình chính: thoát; nếu đang chạy, dừng và chờ tác vụ kết thúc |

Dịch/review yêu cầu xác nhận `y` trước khi gọi API. Trong trình sửa, Esc hỏi trước khi bỏ nội dung chưa lưu; lưu bị lỗi vẫn giữ nội dung đang nhập. Nếu bản dịch, quy ước hoặc cấu hình workspace đã được cập nhật ở nơi khác, TUI báo xung đột để bạn đối chiếu. Trình sửa hiển thị xuống dòng kiểu LF và tab thành bốn dấu cách; khi lưu phần đã sửa dùng định dạng này. File quá 9999 dòng hoặc có ký tự không hỗ trợ được giữ nguyên và báo lỗi khi mở. Nhật ký trên màn hình giữ tối đa 500 dòng; dữ liệu lâu dài nằm trong workspace.

## 3. Tạo workspace và nhập tập đầu

```powershell
$workspacePath = Join-Path $PWD 'workspace/bo-truyen'
.\novel-translate-cli.exe init $workspacePath --series bo-truyen --title "Tên bộ truyện"
.\novel-translate-cli.exe volume add --workspace $workspacePath --id vol01 --order 1 --source "D:/Sach/Tap01_JP.epub"
.\novel-translate-cli.exe import --workspace $workspacePath --volume vol01
.\novel-translate-cli.exe status --workspace $workspacePath
```

Truyền trực tiếp file EPUB, không cần giải nén. `import` lập kiểm kê nguồn và mapping các khối văn bản, chương, ảnh, thuộc tính và chú thích. Thứ tự đọc lấy từ spine/anchors, không theo tên file. Nội dung chưa hỗ trợ được ghi rõ và có thể chặn xuất hoàn chỉnh.

Thêm tập tiếp theo vào cùng workspace:

```powershell
.\novel-translate-cli.exe volume add --workspace $workspacePath --id vol02 --order 2 --previous vol01 --source "D:/Sach/Tap02_JP.epub"
```

`--previous` nối ngữ cảnh đã được chấp nhận của tập trước. Dữ liệu xưng hô được giới hạn theo thời điểm/cảnh; không lấy trạng thái cuối truyện áp cho hồi tưởng.

Sau các bước import/dịch/export, workspace có dạng:

```text
workspace/bo-truyen/
├── workspace.yaml
├── .env                           # Key cục bộ nếu thiết lập model qua TUI
├── Skills/
│   ├── SKILL.md                    # Quy ước riêng của bộ truyện
│   ├── core/<skill-id>/SKILL.md    # Quy tắc chung đã cài
│   ├── imported/                  # Ghi chú tham khảo đã import
│   └── rules.json                 # Xuất hiện khi có quy tắc được lưu
├── vol01_translation/
│   ├── chapters/                  # Bản thảo Markdown có unit markers
│   └── images/source/             # Ảnh lấy ra bằng image export
├── vol02_translation/
├── Output/                        # EPUB và report sau export
└── .translation_work/             # Snapshots, generations, jobs, evidence, usage
```

File nguồn ngoài workspace chỉ được đọc. Đường dẫn tương đối trong `workspace.yaml`, đường dẫn file của `image assess` và file provider được tính từ workspace, không từ thư mục terminal. `--source` của `volume add` được chuyển thành đường dẫn tuyệt đối khi đăng ký.

## 4. Cấu hình model

Cách nhanh: mở TUI và nhấn **m**, không cần lệnh đặt biến môi trường. Ô API OpenAI chỉ áp dụng cho giao thức `openai`; Gemini/Anthropic dùng giao thức native. Để trống key khi lưu lại sẽ dùng key hiện có. Nếu đổi dịch vụ, nhập key tương ứng của dịch vụ mới.

Nếu muốn cấu hình nâng cao bằng file, chọn một mẫu:

| Dịch vụ / giao thức | File mẫu | Lưu ý |
| --- | --- | --- |
| OpenAI-compatible | [config.example.jsonc](config.example.jsonc) | Điền `base_url`, model; `api` là `chat` hoặc `responses` |
| Gemini native | [config.gemini.example.jsonc](config.gemini.example.jsonc) | `type: gemini`; không thêm `api: chat` |
| Anthropic native | [config.anthropic.example.jsonc](config.anthropic.example.jsonc) | `type: anthropic`; không thêm `api: chat` |

Ví dụ sao chép mẫu OpenAI-compatible vào thư mục cấu hình riêng:

```powershell
New-Item -ItemType Directory -Force .\config | Out-Null
Copy-Item .\config.example.jsonc .\config\provider.jsonc
```

Mở file vừa sao chép, thay **tất cả** placeholder tên model/endpoint theo dịch vụ và quyền truy cập của API key. Đặt `context_window` theo model nếu cần; khi không biết, runtime dùng registry hoặc mặc định bảo thủ. Mẫu chưa điền chưa phải cấu hình chạy được.

Key có thể lưu trong `.env` cạnh `workspace.yaml`, cạnh file provider, hoặc trong thư mục chạy ứng dụng. Tạo từ mẫu rồi mở file để điền key của bạn:

```powershell
Copy-Item .\.env.example .\workspace\bo-truyen\.env
```

Linux/macOS:

```sh
cp .env.example ./workspace/bo-truyen/.env
```

Nội dung `.env` dùng tên biến trùng với `api_key_env` trong file provider:

```dotenv
NOVEL_TRANSLATE_API_KEY=YOUR_API_KEY
```

Ứng dụng tự đọc `.env` mỗi lần cần cấu hình model, không cần chạy `source`, `export` hoặc lệnh PowerShell ở mỗi phiên. Thứ tự ưu tiên cho từng key: **biến môi trường có giá trị → `.env` workspace → `.env` cạnh provider → `.env` thư mục đang chạy**. Giá trị môi trường rỗng cho phép đọc file. Các file này không được nạp vào môi trường chung của tiến trình, nên mở workspace khác không kế thừa key vừa đọc. Nếu terminal đang đặt key cùng tên, key đó vẫn thắng key mới lưu trong TUI; bỏ thiết lập cũ hoặc mở terminal mới để dùng key trong file.

TUI giữ các biến khác khi cập nhật `.env`, nhưng có thể định dạng lại file và bỏ comment. Key được lưu dạng văn bản trong file cục bộ; file do ứng dụng tạo dùng quyền `0600` trên hệ thống POSIX. Repo và Docker build context đã loại `.env`/`.env.*`, trừ mẫu `.env.example` được theo dõi trong Git.

Trong TUI, dùng **Chọn file cấu hình API** để gắn file provider vào workspace. Hoặc sửa `workspace/bo-truyen/workspace.yaml`, thêm/sửa dòng ở cấp ngoài cùng:

```yaml
provider_config: ../../config/provider.jsonc
```

Đường dẫn trên đúng khi dùng cấu trúc ví dụ trong repo. Có thể dùng đường dẫn tuyệt đối. Nếu bỏ `provider_config`, CLI tìm `~/.novel-translate/config.json`. Không lưu key trong sách/skill hoặc commit cấu hình chứa key. File JSON/JSONC hỗ trợ UTF-8 và UTF-8 có BOM; không dùng UTF-16.

Ba role là `translation` (dịch), `translation_review` (đối chiếu) và `translation_research` (AI đề xuất quy ước). Role không khai báo dùng model mặc định; có thể cấu hình model/fallback khác nhau theo mẫu. Research web do research adapter thực hiện, không tự có chỉ vì chọn một model.

Endpoint nội bộ không xác thực có thể dùng `allow_no_api_key: true` và bỏ `api_key_env`. Điều này không thay thế API key của dịch vụ yêu cầu xác thực.

### Giới hạn khi chạy thử

Trong TUI, nhấn **c** để đổi giới hạn requests và timeout. Các tùy chọn chi tiết có thể sửa trong phần `run` đã có của `workspace.yaml`, không tạo thêm phần `run` trùng lặp:

```yaml
run:
  chapters_per_batch: 2
  auto_continue: true
  max_contract_corrections: 3
  max_repair_rounds: 3
  max_transport_retries: 3
  request_timeout_seconds: 180
  max_input_bytes: 16000
  max_output_tokens: 12000
  max_requests: 20
```

`max_requests` tính cộng dồn trên **toàn workspace**, gồm retries/fallbacks; `0` là không giới hạn. Khi tăng giới hạn rồi `resume`, kết quả còn hợp lệ được dùng lại. Hai chương mỗi đợt không có nghĩa mỗi lần gọi API chứa hai chương.

`budget.book_usd` nằm trong file provider; `0` tắt giới hạn USD. Khi bật giới hạn mà giá request trước không xác định, runtime dừng thay vì đoán. Chi phí còn phụ thuộc token ngữ cảnh/quy ước, review, sửa lỗi, suy luận và retries; cần đo trên mẫu thật, không có giá cố định cho một tập.

## 5. Quy ước, thuật ngữ và xưng hô

Nhấn **s** trong TUI để sửa `Skills/SKILL.md` của bộ truyện và **Ctrl+S** để lưu, hoặc mở file bằng trình soạn thảo của bạn. File mặc định có priority 100; quy ước riêng cần nêu rõ phạm vi và ưu tiên nguồn Nhật. Ví dụ:

```markdown
---
id: bo-truyen
priority: 100
---
# Quy ước bộ truyện
- Giữ hậu tố kính ngữ theo nguyên văn.
- Không tự thêm hành động, cảm xúc hoặc thông tin không có trong nguồn.
- Tên nhân vật và thuật ngữ theo bảng đã xác nhận của bộ truyện.
- Xưng hô phụ thuộc người nói → người nghe và thời điểm của cảnh.
```

Tránh đưa tình tiết tương lai vào chỉ dẫn chung. Quy tắc chung, quy ước riêng và dữ liệu tham khảo có scope/version; xung đột quy tắc đã chốt cần được giải quyết rõ ràng.

```powershell
.\novel-translate-cli.exe skill list --workspace $workspacePath
.\novel-translate-cli.exe skill import --workspace $workspacePath --source "D:/GhiChu/quy-uoc.md"
.\novel-translate-cli.exe skill import --workspace $workspacePath --source "D:/GhiChu/tap26-27.md" --from-volume 26 --until-volume 28
```

`until-volume` là cận trên **không bao gồm**. Import giữ văn bản, hash và nguồn; không biến mọi ghi chú thành quy tắc đã xác nhận. Đoạn chưa rõ phạm vi được giữ riêng. AI tạo quy ước dùng bản dịch đã được chấp nhận; nếu thiếu cơ sở, mục mới vẫn là candidate/conflict.

```powershell
.\novel-translate-cli.exe skill research --workspace $workspacePath --volume vol01 --term "Thuật ngữ có trong nguồn"
.\novel-translate-cli.exe skill generate --workspace $workspacePath --volume vol01
.\novel-translate-cli.exe skill resolve --workspace $workspacePath --id RULE_ID --decision confirm --reason "Căn cứ xác nhận"
```

Research cần khai báo `research_sources` trong YAML: `api` hỗ trợ `fetch` (URL cụ thể), `mediawiki` (site dùng API MediaWiki) hoặc `json` (endpoint tìm kiếm trả danh sách URL). Mỗi nguồn có `id`, `base_url`, `edition`, `kind`. Chỉ fetch/redirect trong origin đã cấu hình; lưu URL, hash nội dung và thời điểm. Nguồn official/fandom là bằng chứng tham khảo, không thay nguyên tác và không được tạo URL giả. Dùng `retranslate plan` sau khi đổi quy ước ảnh hưởng bản dịch cũ.

## 6. Dịch, xem tiến độ và sửa tay

```powershell
.\novel-translate-cli.exe translate --workspace $workspacePath --volume vol01
.\novel-translate-cli.exe status --workspace $workspacePath --json
.\novel-translate-cli.exe resume --workspace $workspacePath --volume vol01
```

Luồng tự dịch → review → sửa trong giới hạn rồi tiếp tục; dừng nếu còn vấn đề chặn, xung đột hoặc chạm giới hạn. Ctrl+C yêu cầu dừng. `resume` dùng lại kết quả hợp lệ đã lưu; request đã gửi nhưng chưa lưu kết quả có thể phải gọi lại và tính phí lại.

Trong TUI, nhấn **e**, chọn chương rồi đoạn cần sửa. Nguồn Nhật hiển thị phía trên; Tab chuyển vùng để cuộn nguồn hoặc nhập bản dịch. **Ctrl+S** lưu revision và cập nhật Markdown cùng lúc, không cần chạy `sync`. Inline tokens `⟦…⟧` phải giữ nguyên; bản sửa làm mất token hoặc đổi nguồn sẽ bị từ chối. Bản sửa và ngữ cảnh phía sau cần được review lại.

Nếu sửa bằng trình soạn thảo ngoài, chỉ đổi văn bản giữa `<!-- unit:ID -->` và `<!-- /unit -->` trong `vol01_translation/chapters/*.md`. Giữ nguyên markers, thứ tự và inline tokens `⟦…⟧`. Sau đó:

```powershell
.\novel-translate-cli.exe sync --workspace $workspacePath --volume vol01
.\novel-translate-cli.exe review --workspace $workspacePath --volume vol01 --chapter CHAPTER_ID
```

`CHAPTER_ID` lấy từ `status --json`. Bản sửa tay được lưu thành revision riêng; cần review hoặc quyết định giữ có lý do. Khi đổi nguồn, quy ước hoặc ngữ cảnh:

```powershell
.\novel-translate-cli.exe retranslate plan --workspace $workspacePath --volume vol01
.\novel-translate-cli.exe retranslate apply --workspace $workspacePath --plan PLAN_ID
```

Đọc plan trước khi apply. `PLAN_ID` do lệnh plan trả về. File EPUB nguồn thay đổi phải qua `source plan` / `source apply`. `restore --generation GENERATION_ID` khôi phục một generation đã lưu; không ghi đè bản sửa tay chưa sync. `scene set/clear` dành cho mốc thời gian/hồi tưởng, cần source quote và lý do; xem `--help` để biết flags.

## 7. Xử lý chữ trong ảnh

Xem danh sách và lấy ảnh ra workspace để kiểm tra/edit:

```powershell
.\novel-translate-cli.exe image list --workspace $workspacePath --volume vol01
.\novel-translate-cli.exe image export --workspace $workspacePath --volume vol01
```

`image export` trả JSON gồm occurrence ID và đường dẫn ảnh; mặc định ghi vào `vol01_translation/images/source/`. Có thể đổi bằng `--directory` tương đối với workspace. Lệnh không ghi đè file đã được bạn sửa; nên lưu ảnh edit ở thư mục riêng.

Mỗi lần xuất hiện của ảnh cần một quyết định có lý do:

| Disposition | Ý nghĩa |
| --- | --- |
| `no-text` | Đã xem và xác nhận không có chữ cần dịch |
| `keep-original` | Chủ động giữ chữ/ảnh gốc, ví dụ logo/chữ ký; không được coi là đã dịch phần chữ đó |
| `replace` | Dùng ảnh Việt đã tự edit, giữ định dạng tài nguyên gốc |
| `transcript` | Giữ ảnh và thêm bản dịch chữ sau khối nội dung chứa ảnh |

```powershell
.\novel-translate-cli.exe image assess --workspace $workspacePath --volume vol01 --id IMAGE_ID --disposition no-text --reason "Đã xem: tranh không có chữ"
.\novel-translate-cli.exe image assess --workspace $workspacePath --volume vol01 --id IMAGE_ID --disposition replace --file "images/edited/anh-viet.jpg" --reason "Đã đối chiếu ảnh edit với nguồn"
.\novel-translate-cli.exe image assess --workspace $workspacePath --volume vol01 --id IMAGE_ID --disposition transcript --transcript-file "images/anh-viet.txt" --reason "Đã đối chiếu bản dịch chữ trong ảnh"
```

Thay `IMAGE_ID` bằng occurrence ID thật, và chọn một quyết định phù hợp cho mỗi ảnh. File transcript phải là UTF-8; có thể dùng `--transcript "Bản dịch ngắn"` thay file. TUI cũng có các thao tác xuất ảnh và lưu quyết định.

Ảnh thay thế được kiểm tra hash khi verify/export; sửa file sau khi xác nhận thì phải assess lại. Các vị trí dùng chung tài nguyên phải thống nhất dùng ảnh gốc hoặc cùng bản thay thế. Chú thích ảnh độc lập có thể khác nhau theo vị trí. Readback kiểm tra phần chú thích, thứ tự ảnh và bytes tài nguyên.

Chưa có OCR, tự dịch hay tự xóa/chèn chữ vào tranh. Bạn đọc/dịch/kiểm tra và edit ảnh bằng công cụ riêng, rồi đưa kết quả vào workspace.

## 8. Kiểm tra và xuất EPUB

Tải và giải nén **toàn bộ** [EPUBCheck distribution](https://github.com/w3c/epubcheck/releases), giữ JAR cùng thư mục `lib`. Kiểm tra Java và checker:

```powershell
java -version
java -jar "D:/Tools/epubcheck/epubcheck.jar" --version
```

Nhấn **c** trong TUI để chọn JAR/executable của EPUBCheck và thư mục font. Hoặc sửa phần `output` hiện có trong `workspace.yaml`:

```yaml
output:
  directory: Output
  template: sample-vi
  epubcheck: D:/Tools/epubcheck/epubcheck.jar
  # Tùy chọn:
  # font_directory: D:/Tools/fonts/liberation
```

`epubcheck` cũng nhận đường dẫn executable. Thư mục font nếu khai báo phải có `LiberationSerif-Regular.ttf`, `LiberationSerif-Bold.ttf`, `LiberationSerif-Italic.ttf`, `LiberationSerif-BoldItalic.ttf`; nếu bỏ qua thì không nhúng font. Hiện chỉ có template `sample-vi`, giữ cấu trúc nguồn và áp kiểu đọc ngang/LTR.

```powershell
.\novel-translate-cli.exe verify --workspace $workspacePath --volume vol01
.\novel-translate-cli.exe export --workspace $workspacePath --volume vol01 --preview
.\novel-translate-cli.exe export --workspace $workspacePath --volume vol01
```

- `verify`: kiểm tra độ phủ, revision/ngữ cảnh còn hợp lệ và quyết định ảnh; chưa dựng EPUB hay chạy EPUBCheck.
- `--preview`: xuất bản xem trước có nhãn; có thể còn thiếu dịch hoặc checker chưa pass. Đọc report đi kèm, không coi là bản hoàn chỉnh.
- Export hoàn chỉnh: yêu cầu bản dịch đã đối chiếu còn hợp lệ, không còn vấn đề chặn và mọi ảnh có quyết định hợp lệ; sau đó dựng file tạm, mở lại đối chiếu rồi chạy EPUBCheck thật trước khi đưa vào `Output/`.

Đầu ra mặc định: `Output/vol01_VI.epub` hoặc `Output/vol01_VI.preview.epub`, kèm `.report.json`. Có thể ghi đè bản xuất cùng tên sau khi kiểm tra đạt. Report phân biệt nghĩa/độ phủ, cấu trúc và định dạng. Vẫn cần mở trên reader bạn dùng để kiểm tra trình bày.

Xuất một phạm vi chương dùng `--from CHAPTER_ID --to CHAPTER_ID`, bao gồm cả hai đầu; có thể kết hợp `--preview`. Chương có thể gồm nhiều XHTML/trang ảnh. Phạm vi hoàn chỉnh không có nghĩa toàn tập đã hoàn chỉnh.

## 9. Kiểm thử và đóng gói

```powershell
go vet ./...
go test -buildvcs=false -count=1 ./...
git diff --check
```

Suite mặc định dùng fixtures tổng hợp, model giả và HTTP server cục bộ. Khi `NOVEL_TRANSLATE_LIVE_PROVIDER` chưa đặt, test model thật được skip.

Linux/macOS có thể dùng `make test` và `make vet`. Các test giao diện kiểm tra form, lưu key và `.env`, sửa bản dịch/quy ước, bảo vệ xung đột và mở TUI bằng input giả; vẫn cần thử trực tiếp trên terminal bạn dùng.

Chạy thêm kiểm tra EPUB hoàn chỉnh bằng Java/checker thật, với model giả:

```powershell
$env:NOVEL_TRANSLATE_TEST_EPUBCHECK = 'D:/Tools/epubcheck/epubcheck.jar'
go test -buildvcs=false -count=1 -run 'TestOfflineCLIWorkflowAndPreview|TestImageCLIEditRoundTrip' -v ./internal/translation/cli
```

Smoke model thật **có thể tính phí**, tối đa 6 requests; chỉ kiểm tra kết nối/hợp đồng phản hồi, chưa nghiệm thu chất lượng truyện:

```powershell
$env:NOVEL_TRANSLATE_LIVE_PROVIDER = (Resolve-Path .\config\provider.jsonc).Path
go test -count=1 -run '^TestLiveConfiguredProviderSmoke$' -v ./internal/translation/model
Remove-Item Env:NOVEL_TRANSLATE_LIVE_PROVIDER
```

Race test trên môi trường có cgo/C compiler:

```sh
go test -race -buildvcs=false -count=1 ./internal/translation/state ./internal/translation/engine ./internal/translation/continuity ./internal/translation/revision ./internal/translation/model ./internal/translation/cli
```

CI khai báo test Windows/Linux, race trên Linux và một job kiểm chứng export bằng EPUBCheck thật. Workflow Docker/snapshot chạy thủ công. Có workflow không đồng nghĩa đã chạy thành công: xem logs của lần chạy cụ thể.

Kiểm chứng ngày **20/10/2026**: toàn suite offline, export với EPUBCheck thật, vet/format và build Windows/Linux/macOS amd64/arm64 đều đạt trên checkout này. Đã mở binary Windows trực tiếp, tạo workspace, lưu model/key giả có che nội dung và mở trình sửa quy ước trong TUI; không gọi API trả phí. Script khởi chạy/syntax được thử qua Git Bash Windows. Chạy TUI/installer thực tế trên Linux/macOS, race test, Docker và nghiệm thu model thật vẫn cần môi trường tương ứng.

Docker là tùy chọn: chuẩn bị `config/provider.json`, `tools/epubcheck/` và `workspace/`, rồi dùng `docker compose run --rm --build novel-translate tui`. Workspace được bind-mount nên `.env` bên trong từng workspace cũng được đọc tự động; Compose có thể chuyển `NOVEL_TRANSLATE_API_KEY` từ shell và giá trị không rỗng này được ưu tiên. Workspace trong container dùng đường dẫn Linux: đặt `provider_config: /root/.novel-translate/config.json` (hoặc bỏ trường này để dùng fallback) và checker `/tools/epubcheck/epubcheck.jar`; không dùng đường dẫn `D:/...` trong đó. Nếu thiết lập model bằng TUI, file provider và `.env` được lưu ngay trong workspace đó.

## 10. Giới hạn và xử lý lỗi

| Exit code | Ý nghĩa |
| --- | --- |
| `0` | Lệnh thành công trong phạm vi được yêu cầu |
| `1` | Lỗi cấu hình, IO hoặc request |
| `2` | Còn vấn đề chưa giải quyết; xem report/diagnostics |

- Phạm vi hiện tại là EPUB JP→VN; chưa có PDF hoặc luồng sáng tác.
- EPUB mã hóa, script, chữ SVG/CSS sinh động và cấu trúc chưa được hỗ trợ có thể chặn xuất hoàn chỉnh. Không tự bỏ qua để tuyên bố đủ nội dung.
- Trình bày giữ XHTML/spine và vị trí ảnh nguồn; không sao chép chính xác phân trang của bản mẫu hay bảo đảm giống nhau trên mọi reader.
- `keep-original` là ngoại lệ có quyết định của người dùng. Nó không chứng minh toàn bộ chữ Nhật đã được dịch.
- Chi phí chưa rõ được ghi `unknown`; một request thất bại hoặc bị hủy vẫn có thể phát sinh phí phía dịch vụ.
- Nghiệm thu model thật, đọc thử bản Việt trên reader, chạy cả tập và xuyên tập cần thực hiện trước khi dịch hàng loạt.

Lỗi thường gặp: kiểm tra `provider_config` và biến môi trường nếu báo thiếu key; dùng `resume` sau lỗi mạng; giải quyết xung đột theo `retranslate plan`; assess lại ảnh khi hash đổi; cấu hình EPUBCheck và đọc diagnostics khi final export bị chặn. Nếu Windows báo `An Application Control policy has blocked this file`, kiểm tra môi trường/chính sách thực thi và dùng CI hoặc môi trường Linux phù hợp để kiểm chứng; không đổi key để xử lý lỗi này.

## 11. Sao lưu và giấy phép

Dừng tác vụ ghi trước khi sao lưu. Sao lưu **toàn bộ workspace**, gồm YAML, skills, bản thảo, file ảnh thay thế và `.translation_work/`. Nếu ảnh hoặc provider config nằm ngoài workspace, sao lưu chúng riêng. Chỉ giữ Markdown mà bỏ snapshots/generations sẽ không đủ để resume/restore đúng trạng thái.

Module path `github.com/voocel/ainovel-cli` được giữ từ dự án gốc để bảo toàn imports/attribution; entry point của sản phẩm dịch là `cmd/novel-translate-cli`.

Phân phối theo [MIT License](LICENSE).
