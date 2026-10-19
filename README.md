# novel-translate-cli

> **Công cụ tự động dịch Light Novel từ tiếng Nhật sang tiếng Việt dành cho định dạng EPUB.**
> Giữ nguyên 100% hình ảnh, mục lục, chú thích; nhớ xưng hô nhân vật xuyên suốt nhiều tập; tự động soi lỗi và xuất sách chuẩn đẹp cho mọi thiết bị đọc (Kindle, Kobo, iPad, điện thoại).

---

## Tính năng nổi bật

- **Bảo toàn sách gốc**: Giữ nguyên vị trí tranh minh họa, tranh đôi, mục lục và các liên kết chú thích hai chiều.
- **Nhớ cách xưng hô nhân vật**: Không bị tình trạng đầu truyện xưng "tôi", cuối truyện xưng "anh/em"; không để lộ trước tình tiết của các tập tương lai.
- **AI Reviewer tự động soi lỗi**: Tự động phát hiện khi câu dịch bị bỏ sót ý, bịa thêm chi tiết hoặc sai đại từ xưng hô, sau đó tự động sửa lại.
- **Xử lý linh hoạt chữ trong ảnh**: Hỗ trợ giữ nguyên ảnh, thay bằng bìa tiếng Việt đã edit, hoặc tự chèn bảng dịch chú thích bên dưới tranh.
- **Cho phép bạn tự sửa tay**: Tự do mở file chữ ra sửa câu cú theo ý thích; AI sẽ ghi nhớ và không bao giờ ghi đè lên câu bạn đã sửa.
- **An toàn tuyệt đối**: Mất mạng, sập nguồn giữa chừng chỉ cần bấm tiếp tục (`resume`), không phải dịch lại từ đầu và không bị tốn tiền API trùng lặp.
- **Theo dõi chi phí token**: Lệnh `status` hiển thị số request đã gửi, chi phí USD đã biết và số request chưa rõ chi phí, giúp bạn kiểm soát ngân sách khi dịch trọn một tập sách.

---

## Cách hoạt động (Dành cho người mới)

```text
[File EPUB Nhật gốc]
       |  (1. Phân tích sách thành từng câu văn & tranh ảnh)
       v
[AI Dịch Thuật] --> [AI Biên Tập Soi Lỗi & Tự Sửa]
       |  (2. Bạn có thể mở file đọc thử & sửa tay nếu thích)
       v
[Xử lý tranh ảnh: Bìa sách, tranh có chữ]
       |  (3. Tự động đóng gói & chuyển chữ sang đọc ngang LTR)
       v
[File EPUB Tiếng Việt Hoàn Chỉnh] (Mở đọc ngay trên Kindle, Kobo, Apple Books...)
```

---

## Hướng dẫn cài đặt

Bạn chỉ cần máy tính chạy **Windows, macOS hoặc Linux**.

### Cách 1: Dùng file chạy sẵn (Đơn giản nhất cho Windows)
1. Tải file `novel-translate-cli.exe` về máy tính của bạn.
2. Mở cửa sổ dòng lệnh (**PowerShell** trên Windows hoặc **Terminal** trên Mac/Linux) tại thư mục chứa file và gõ:
   ```powershell
   .\novel-translate-cli.exe --help
   ```

### Cách 2: Tự biên dịch từ mã nguồn (Nếu máy đã cài Go 1.25 trở lên)
```sh
go build -trimpath -o novel-translate-cli ./cmd/novel-translate-cli
```

---

## Hướng dẫn sử dụng từ A đến Z

### Bước 1: Chuẩn bị API Key (Chìa khóa để AI làm việc)

Bạn có thể dùng **Google Gemini** (khuyên dùng vì vừa thông minh, dịch nhanh lại có gói miễn phí/siêu rẻ) hoặc **OpenAI (ChatGPT)**:
- **Lấy key Gemini**: Vào [Google AI Studio](https://aistudio.google.com/), đăng nhập tài khoản Google và bấm **Get API Key**.
- Mở PowerShell và gõ lệnh sau để lưu key vào máy:
  ```powershell
  $env:NOVEL_TRANSLATE_API_KEY = "dán_mã_api_key_của_bạn_vào_đây"
  ```

### Bước 2: Tạo file cấu hình AI (`provider.jsonc`)

Tạo một file văn bản tên là `provider.jsonc` (ví dụ lưu tại `D:/provider.jsonc` hoặc ngay trong thư mục làm việc):

```jsonc
{
  "provider": "gemini-service",
  "model": "gemini-2.5-flash",
  "providers": {
    "gemini-service": {
      "type": "gemini",
      "api_key_env": "NOVEL_TRANSLATE_API_KEY",
      "models": [{"name": "gemini-2.5-flash", "context_window": 1000000}]
    }
  },
  "roles": {
    // Model chuyên dịch văn
    "translation": {"provider": "gemini-service", "model": "gemini-2.5-flash"},
    // Model chuyên soi lỗi và kiểm tra ngữ cảnh
    "translation_review": {"provider": "gemini-service", "model": "gemini-2.5-flash"},
    // Model tra cứu thông tin
    "translation_research": {"provider": "gemini-service", "model": "gemini-2.5-flash"}
  }
}
```

### Bước 3: Bắt đầu dịch cuốn sách đầu tiên!

Trong các ví dụ bên dưới, thay `<ten_bo>` bằng tên bộ truyện bạn muốn (ví dụ `rezero`, `sao`, `mushoku`...) và `<duong_dan_epub>` bằng đường dẫn tới file EPUB gốc tiếng Nhật của bạn.

#### 1. Khởi tạo dự án cho bộ truyện
```powershell
.\novel-translate-cli.exe init ./workspace/<ten_bo> --series <ten_bo> --title "Tên bộ truyện"
```
Lệnh này sẽ tự động tạo thư mục workspace với đầy đủ cấu trúc chuẩn (xem mục "Cấu trúc Workspace" bên dưới).

#### 2. Thêm tập sách tiếng Nhật vào dự án
```powershell
.\novel-translate-cli.exe volume add --workspace ./workspace/<ten_bo> --id vol01 --order 1 --source "<duong_dan_epub>"
```

#### 3. Khai báo file cấu hình AI
Mở file `workspace/<ten_bo>/workspace.yaml` bằng Notepad, thêm đường dẫn file cấu hình ở Bước 2:
```yaml
provider_config: D:/provider.jsonc
```

#### 4. (Rất khuyên dùng) Thiết lập bảng xưng hô & tên nhân vật
Mở file `workspace/<ten_bo>/Skills/SKILL.md` bằng Notepad và điền bảng tên nhân vật, cách xưng hô bạn muốn AI dịch. Xem hướng dẫn chi tiết ở mục "Thiết lập Quy ước Xưng hô & Tên nhân vật" bên dưới.

#### 5. Phân tích nội dung cuốn sách
```powershell
.\novel-translate-cli.exe import --workspace ./workspace/<ten_bo> --volume vol01
```

#### 6. Bấm dịch!
```powershell
.\novel-translate-cli.exe translate --workspace ./workspace/<ten_bo> --volume vol01
```
Hệ thống sẽ tự động bóc tách từng chương, gọi model dịch, chạy luồng review đối chiếu, lưu checkpoint và liên tục thông báo tiến độ trực tiếp trên màn hình.

---

## Cấu trúc Thư mục Workspace & Thiết lập Quy ước (Skills)

### Thư mục Workspace trông như thế nào?

Sau khi bạn chạy lệnh `init`, thư mục dự án sẽ có cấu trúc như sau:

```text
workspace/<ten_bo>/
├── workspace.yaml              # File cấu hình chính (khai báo AI, danh sách tập, font...)
├── Skills/                     # Nơi quản lý quy ước & phong cách dịch thuật
│   ├── SKILL.md                # [Quan trọng] Quy ước chính: tên nhân vật, cách xưng hô, thuật ngữ
│   ├── core/                   # Các quy tắc cốt lõi mặc định đã cài sẵn:
│   │   ├── jp-vn-translation/SKILL.md       # Quy tắc dịch Nhật sang Việt tự nhiên
│   │   ├── dialogue-and-honorifics/SKILL.md # Quy tắc xử lý đại từ, kính ngữ (-san, -kun, -chan...)
│   │   └── translation-review/SKILL.md      # Quy tắc dành cho AI Reviewer soi lỗi
│   ├── imported/               # Nơi lưu trữ tài liệu nếu bạn dùng lệnh skill import từ file ngoài
│   └── rules.json              # Danh sách quy tắc dịch thuật đã được chốt
├── vol01_translation/          # Bản thảo của tập sách (tự sinh ra trong quá trình dịch)
│   └── chapters/               # Các file Markdown từng chương (vd: vol01-c001.md)
├── Output/                     # Thư mục chứa file sách EPUB tiếng Việt sau khi xuất (export)
└── .translation_work/          # Dữ liệu nội bộ (checkpoint, cache, trạng thái review - không cần đụng vào)
```

### Thiết lập Quy ước Xưng hô & Tên nhân vật (`Skills/SKILL.md`)

File `Skills/SKILL.md` chính là **"bộ não ghi nhớ"** của bộ truyện. Khi dịch và khi soi lỗi, cả **AI Dịch** lẫn **AI Reviewer** đều bắt buộc phải tuân theo file này với độ ưu tiên cao nhất (`priority: 100`).

Mở file `workspace/<ten_bo>/Skills/SKILL.md` bằng Notepad và soạn thảo. Ví dụ cho một bộ truyện giả định:

```markdown
---
id: ten_bo_truyen
priority: 100
---
# Quy ước dịch bộ truyện

## 1. Tên nhân vật (Nhật -> Việt)
- 田中太郎 -> Tanaka Tarou
- 鈴木花子 -> Suzuki Hanako
- リナ -> Rina

## 2. Bảng xưng hô nhân vật (Rất quan trọng)
- Tarou gọi Hanako: xưng "tôi - cô ấy", khi thân thiết xưng "tớ - cậu"
- Hanako gọi Tarou: xưng "tôi - anh ấy"
- Rina: tự xưng "ta", gọi người khác "ngươi"

## 3. Thuật ngữ & Địa danh
- 王国 -> Vương quốc
- 魔法 -> Ma pháp
- 勇者 -> Dũng giả
```

**Lợi ích khi có file SKILL.md:**
- Cho dù bộ truyện dài hàng chục tập, AI vẫn nhớ chính xác ai gọi ai như thế nào, không bao giờ bị lẫn lộn xưng hô giữa các tập.
- Tên nhân vật và thuật ngữ luôn đồng nhất 100% từ đầu đến cuối sách.

### Nạp nhanh ghi chú có sẵn từ bên ngoài (`skill import`)

Nếu bạn đã có sẵn file ghi chú/glossary từ trước, không cần chép tay lại, chỉ cần gõ:

```powershell
.\novel-translate-cli.exe skill import --workspace ./workspace/<ten_bo> --source "D:/glossary.md"
```

Hệ thống sẽ tự động phân tích và nạp toàn bộ thuật ngữ vào thư mục `Skills/imported/` của dự án.

Nếu file ghi chú chỉ áp dụng cho một số tập nhất định, dùng cờ `--from-volume` và `--until-volume` (cận trên không bao gồm):
```sh
novel-translate-cli skill import --workspace W --source ./notes_vol26.md --from-volume 26 --until-volume 28
```

Xem danh sách tất cả các kỹ năng đang được áp dụng:
```powershell
.\novel-translate-cli.exe skill list --workspace ./workspace/<ten_bo>
```

---

## Đọc thử & Tự sửa tay theo ý thích (Tùy chọn)

Sau khi dịch xong (hoặc dịch được vài chương), bạn có thể mở các file bản thảo ra đọc:
- Đường dẫn: `workspace/<ten_bo>/vol01_translation/chapters/vol01-c001.md`
- Bạn có thể mở file này bằng bất kỳ ứng dụng nào (Notepad, VS Code, Word, Obsidian...).
- Sửa văn bản giữa `<!-- unit:ID -->` và `<!-- /unit -->` trong file; giữ nguyên ID, thứ tự và tokens.
- Sau khi lưu file, gõ lệnh sau để cập nhật vào hệ thống:
  ```powershell
  .\novel-translate-cli.exe sync --workspace ./workspace/<ten_bo> --volume vol01
  ```
Những câu bạn đã sửa tay sẽ được hệ thống đánh dấu bảo vệ vĩnh viễn, AI sẽ không bao giờ ghi đè lên nữa!

Nếu bạn muốn dịch lại một số đoạn sau khi sửa quy ước hoặc chỉnh scene:
```sh
novel-translate-cli retranslate plan --workspace W --volume vol01
novel-translate-cli retranslate apply --workspace W --plan PLAN_ID
```

Nếu bạn quyết định giữ một bản dịch sau khi xem vấn đề mà reviewer nêu:
```sh
novel-translate-cli review --workspace W --volume vol01 --keep BLOCK_ID --reason "Căn cứ quyết định"
```

---

## Xử lý tranh ảnh trong sách

Trước khi xuất file sách cuối cùng, hãy kiểm tra danh sách ảnh:
```powershell
.\novel-translate-cli.exe image list --workspace ./workspace/<ten_bo> --volume vol01
```

Mọi ảnh trong sách đều phải có quyết định hợp lệ trước khi chạy lệnh `export` cuối cùng. Hệ thống hỗ trợ 4 kiểu xử lý:

**1. Ảnh thuần minh họa (không có chữ):**
```powershell
.\novel-translate-cli.exe image assess --workspace W --volume vol01 --id ID_ANH --disposition no-text --reason "Tranh đen trắng, không có chữ"
```

**2. Ảnh có chữ nhưng giữ nguyên bản gốc** (ví dụ chữ ký tác giả, logo):
```powershell
.\novel-translate-cli.exe image assess --workspace W --volume vol01 --id ID_ANH --disposition keep-original --reason "Chữ ký tác giả giữ nguyên"
```

**3. Thay bằng ảnh đã edit tiếng Việt** (ví dụ bìa sách, bản đồ):
```powershell
.\novel-translate-cli.exe image assess --workspace W --volume vol01 --id ID_ANH --disposition replace --file covers/vol01_vi.jpg --reason "Ảnh bìa đã typeset tiếng Việt"
```

**4. Thêm bảng chú thích dịch bên dưới ảnh** (ví dụ trang giới thiệu nhân vật):
```powershell
.\novel-translate-cli.exe image assess --workspace W --volume vol01 --id ID_ANH --disposition transcript --transcript "Nhân vật A: Mô tả\nNhân vật B: Mô tả" --reason "Dịch trang giới thiệu nhân vật"
```
Hoặc truyền file transcript:
```powershell
.\novel-translate-cli.exe image assess --workspace W --volume vol01 --id ID_ANH --disposition transcript --transcript-file ./transcripts/char_intro.txt --reason "Dịch từ file transcript"
```
Bản dịch sẽ được hiển thị thành một khung chú thích trang nhã ngay dưới bức ảnh trong file sách EPUB.

---

## Xuất file EPUB tiếng Việt thành phẩm

Khi đã dịch xong các chương:
```powershell
.\novel-translate-cli.exe export --workspace ./workspace/<ten_bo> --volume vol01
```
File thành phẩm sẽ nằm tại: `workspace/<ten_bo>/Output/vol01_VI.epub`.

Bạn có thể xuất thử chỉ vài chương (xem trước khi dịch xong hết):
```sh
novel-translate-cli export --workspace W --volume vol01 --preview
novel-translate-cli export --workspace W --volume vol01 --from c004 --to c005
```

Bạn chỉ việc chép file EPUB vào điện thoại, máy tính bảng hoặc máy đọc sách (Kindle, Kobo) là có thể thưởng thức cuốn sách tiếng Việt hoàn hảo!

---

## Theo dõi tiến độ & chi phí token

Gõ lệnh sau để xem trạng thái dịch hiện tại:
```powershell
.\novel-translate-cli.exe status --workspace ./workspace/<ten_bo>
```

Kết quả hiển thị:
- Số khối đã dịch và đã đối chiếu xong cho mỗi tập.
- Số khối "stale" (cần dịch lại do thay đổi quy ước/nguồn).
- **Tổng số request đã gửi tới API**.
- **Chi phí USD đã biết** (tính từ token usage mà provider trả về).
- **Số request chưa rõ chi phí** (provider không trả thông tin giá).

Thêm cờ `--json` để lấy dữ liệu dạng JSON chi tiết (tiện cho script hoặc tích hợp):
```powershell
.\novel-translate-cli.exe status --workspace ./workspace/<ten_bo> --json
```

Bạn cũng có thể giới hạn số request tối đa trong `workspace.yaml`:
```yaml
run:
  max_requests: 500   # 0 = không giới hạn; số dương = giới hạn request thật
```

Hoặc giới hạn ngân sách USD trong `provider.jsonc`:
```jsonc
{
  "budget": {
    "book_usd": 0.50
  }
}
```
Khi bật giới hạn USD mà chi phí request trước chưa rõ, hệ thống sẽ dừng lại thay vì suy đoán giá.

---

## Thêm tập sách tiếp theo

Khi muốn dịch tập 2, tập 3..., chỉ cần đăng ký thêm volume:
```sh
novel-translate-cli volume add --workspace W --id vol02 --order 2 --previous vol01 --source SOURCE_VOL02.epub
novel-translate-cli import --workspace W --volume vol02
novel-translate-cli translate --workspace W --volume vol02
```

`previous` phải là tập đã đăng ký sớm hơn. Context chỉ lấy bản dịch đã chấp nhận trước vị trí hiện tại; không lấy sự kiện tương lai. Thiếu tập trước tạo cảnh báo, không bịa ký ức.

---

## Xử lý khi file EPUB gốc thay đổi

Nếu file EPUB gốc đã đăng ký bị thay đổi (ví dụ bạn tải lại bản mới), lệnh `translate`/`export` sẽ bị chặn. Xử lý như sau:
```sh
novel-translate-cli source plan --workspace W --volume vol01
novel-translate-cli source apply --workspace W --volume vol01 --plan PLAN_ID
novel-translate-cli resume --workspace W --volume vol01
```

Mapping bảo thủ theo document/DOM locator. Khối không map chắc chắn được archive; bản sửa tay map được giữ nhưng phải review lại.

---

## Hồi tưởng / Scene mapping (Nâng cao)

Khi một đoạn trong sách là hồi tưởng lại sự kiện ở tập trước, bạn có thể gán scene để AI dịch đúng ngữ cảnh:
```sh
novel-translate-cli scene set --workspace W --volume vol02 --id FIRST_BLOCK --end LAST_BLOCK --quote "Nguyên văn Nhật" --reason "Căn cứ xác định cảnh" --story-volume 1 --story-chapter 2 --story-block 3
```

Nếu chưa xác định được thời điểm cụ thể:
```sh
novel-translate-cli scene set --workspace W --volume vol02 --id FIRST_BLOCK --end LAST_BLOCK --quote "Nguyên văn Nhật" --reason "Chưa xác định được thời điểm" --uncertain
```

---

## Skills nâng cao & Research

### AI tự đề xuất quy ước
```sh
novel-translate-cli skill generate --workspace W --volume vol01
```
AI sẽ phân tích sách gốc và đề xuất tên nhân vật, thuật ngữ. Bạn xác nhận hoặc từ chối từng đề xuất:
```sh
novel-translate-cli skill resolve --workspace W --id RULE_ID --decision confirm --reason "Bằng chứng và quyết định"
novel-translate-cli skill resolve --workspace W --id RULE_ID --decision reject --reason "Không phù hợp nguồn/quy ước"
```

### Tra cứu thông tin từ web
```sh
novel-translate-cli skill research --workspace W --volume vol01 --term "Thuật ngữ đã có trong nguồn"
```

Cấu hình nguồn tra cứu trong `workspace.yaml`:
```yaml
research_sources:
  - id: wiki-source
    api: mediawiki
    base_url: https://example.fandom.com
    edition: fandom
    kind: community
```

Hỗ trợ 3 loại API: `mediawiki` (gọi `/api.php` để search), `fetch` (URL cấu hình), và `json` (GET với query `q`). Web chỉ là reference data; model không được tự tạo URL làm evidence.

### Cấu trúc file Skill (Cho ai muốn tùy chỉnh sâu)

Core skill là Markdown với optional YAML frontmatter:
```yaml
id: ten-skill            # ID duy nhất
priority: 50             # Số càng cao càng ưu tiên (SKILL.md mặc định = 100)
from:                    # Scope bắt đầu áp dụng (tùy chọn)
  volume: 1
until:                   # Scope kết thúc, không bao gồm (tùy chọn)
  volume: 5
references:              # Paths tương đối tới file tham chiếu (tùy chọn)
  - ./data/characters.md
```

Với thư viện chung, bật skill ID trong `enabled_skills` và thêm `skill_libraries` trong `workspace.yaml`; tránh hai file định nghĩa cùng ID.

---

## Cấu hình nâng cao trong workspace.yaml

Ví dụ cấu hình bổ sung (giữ các defaults mà `init` đã tạo):

```yaml
provider_config: /private/provider.jsonc
output:
  directory: Output
  template: sample-vi
  epubcheck: /tools/epubcheck/epubcheck.jar
  font_directory: /fonts/liberation
```

Thư mục font cần `LiberationSerif-Regular.ttf`, `LiberationSerif-Bold.ttf`, `LiberationSerif-Italic.ttf`, `LiberationSerif-BoldItalic.ttf`.

### Cài EPUBCheck (Kiểm tra chất lượng file EPUB)

Tải EPUBCheck distribution từ [W3C EPUBCheck](https://github.com/w3c/epubcheck/releases), giữ JAR và thư mục lib đi kèm, cài Java thích hợp và chạy:
```sh
java -jar /tools/epubcheck/epubcheck.jar --version
```
Khai báo đường dẫn trong workspace; thiếu checker không cho final export. Với Windows, CLI khởi chạy checker từ thư mục artifact để tránh lỗi đường dẫn workspace Unicode của Java.

---

## Sao lưu & Khôi phục

Dừng writer trước khi backup. Sao lưu toàn bộ workspace gồm:
- `workspace.yaml`
- `Skills/`
- Thư mục chapters
- `Output/`
- **Toàn bộ `.translation_work/`** (bao gồm source snapshots, generations, checksums, scene mappings, jobs, plans, rule history và usage)

Khi restore backup, giữ cấu trúc tương đối, đặt lại provider config/biến môi trường ở máy mới rồi chạy `status`. Credential config nằm riêng, không commit key vào Git.

---

## Bảng tra cứu lệnh nhanh

| Lệnh | Ý nghĩa |
| :--- | :--- |
| `init <thư mục> --series <tên>` | Khởi tạo thư mục dự án cho một bộ truyện mới |
| `volume add ...` | Đăng ký một tập sách tiếng Nhật mới vào bộ truyện |
| `import ...` | Bóc tách và kiểm kê cấu trúc file EPUB gốc |
| `translate ...` | Bắt đầu hoặc tiếp tục dịch tự động |
| `resume ...` | Tiếp tục dịch nếu bị mất mạng hoặc tạm dừng |
| `status [--json] ...` | Xem tiến độ dịch, số request, chi phí USD |
| `sync ...` | Cập nhật những câu bạn đã tự sửa tay trên file Markdown |
| `review ...` | Chạy lại review cho chương cụ thể, hoặc giữ bản dịch sau khi xem vấn đề |
| `retranslate plan/apply ...` | Lập kế hoạch và dịch lại các đoạn bị ảnh hưởng |
| `skill list ...` | Xem danh sách các kỹ năng / quy ước đang kích hoạt |
| `skill import ...` | Nạp file ghi chú / glossary thuật ngữ từ bên ngoài vào dự án |
| `skill generate ...` | AI tự đề xuất tên nhân vật, thuật ngữ từ sách gốc |
| `skill resolve ...` | Xác nhận hoặc từ chối đề xuất quy ước của AI |
| `skill research ...` | Tra cứu thông tin thuật ngữ từ web |
| `scene set/clear ...` | Gán hoặc xóa scene hồi tưởng cho một đoạn sách |
| `source plan/apply ...` | Xử lý khi file EPUB gốc bị thay đổi |
| `image list ...` | Xem danh sách tất cả tranh ảnh trong tập |
| `image assess ...` | Gắn nhãn hoặc thay thế ảnh có chữ |
| `export [--preview] ...` | Đóng gói sách thành file EPUB tiếng Việt hoàn chỉnh |
| `restore ...` | Khôi phục bản dịch về một generation cũ |

---

## Giới hạn hiện tại

- XML EPUB 2/3 theo namespace được hỗ trợ; file mã hóa, script, generated CSS text hoặc SVG text chưa được xử lý hoàn chỉnh.
- Chưa có OCR/dịch chữ trong ảnh tự động.
- Layout giữ XHTML của nguồn và hình xen trong spine; chưa có gộp tất cả trang hình thành `<figure>` theo mẫu.
- Exit code: `0` thành công; `1` lỗi cấu hình/IO/request; `2` còn vấn đề unresolved.

---

## Câu hỏi thường gặp (FAQ)

**1. Mất mạng hoặc máy tính tắt đột ngột thì có phải dịch lại từ đầu không?**
> **Không bao giờ!** Hệ thống lưu checkpoint liên tục sau từng câu. Khi mở máy lại, bạn chỉ cần gõ `resume` là nó tiếp tục ngay từ câu đang làm dở, không mất bất kỳ dữ liệu nào và không tốn phí dịch lại.

**2. Chi phí dịch một tập sách khoảng bao nhiêu tiền?**
> Nếu dùng các model tối ưu như **Gemini 2.5 Flash** hoặc **GPT-4o Mini**, chi phí dịch trọn vẹn cả một tập Light Novel (khoảng 80.000 - 100.000 chữ Nhật) chỉ tốn khoảng **vài nghìn đồng đến dưới 10.000 VND** (thậm chí hoàn toàn miễn phí nếu bạn dùng gói Free Tier của Google). Bạn có thể dùng lệnh `status` bất cứ lúc nào để xem chi phí thực tế đang bao nhiêu.

**3. File EPUB xuất ra có đọc được trên Kindle không?**
> **Hoàn toàn được!** File xuất ra đạt chuẩn quốc tế W3C EPUB 3. Bạn có thể gửi trực tiếp qua tính năng *Send-to-Kindle* của Amazon hoặc đọc trên app Apple Books, Kobo, Moon+ Reader mượt mà.

**4. Làm sao theo dõi xem đã tốn bao nhiêu token/tiền?**
> Gõ `status --workspace <path>` để xem tổng số request, chi phí USD đã biết và số request chưa rõ chi phí. Thêm `--json` để lấy dữ liệu chi tiết dạng JSON. Bạn cũng có thể đặt giới hạn ngân sách trong `provider.jsonc` (trường `budget.book_usd`) để tự động dừng khi vượt mức.

---

## Giấy phép

Mã nguồn được phân phối theo giấy phép [MIT License](LICENSE).
