1) hiện tại đang tag cả tag qc-.... và stg-... sau đó mới push 1 lần

-> đổi thành push độc lập (tag qc xong push, tag stg xong push), nhưng trước khi action push -> kiểm tra thử cả 2 tags đã tồn tại chưa nha

2) mỗi project có 1 rule riêng -> suy nghĩ brain storm cách tối ưu

Nhu cầu của tôi là gõ minhthetus-cli git tag-dev-stg, nó sẽ check project hiện tại và dựa vào rule hiện tại mà cook khác nhau

có một số kiểu tag để deploy, mà tôi chiêm nghiệm đc:
- tag prefix qc để deploy sandbox: qc-v1.1.0
- tag prefix stg để deploy sandbox: stg-v1.1.0
- tag thẳng số để deploy prod: v1.0.0
- tag prefix release để deploy prod: release-v1.10.0
- tag prefix qc-cicd-v1.0.0 để deploy sandbox

---

## 💡 Brainstorm Giải Pháp: Dynamic & User-Centric Git Tagging

### 1. Tại sao cần Push độc lập + Pre-flight Check? (Giải quyết vấn đề 1)
Khi chạy `git push origin tag1 tag2` trong cùng 1 lệnh, nhiều hệ thống CI/CD (GitLab CI / GitHub Actions / Jenkins) nhận 1 payload có nhiều refs cùng lúc sẽ **chỉ trigger 1 pipeline** hoặc **bỏ qua hoàn toàn** để chống spam.
Do đó, luồng chuẩn phải là:
1. **Sync & Pre-flight Check (Kiểm tra toàn bộ trước khi hành động):**
   - Chạy `git fetch --tags` để lấy danh sách tag mới nhất từ remote.
   - Với toàn bộ danh sách `targetTags` chuẩn bị tạo (VD: `["qc-v1.1.0", "stg-v1.1.0"]`), kiểm tra đồng thời ở cả **Local** và **Remote** xem có tag nào đã tồn tại chưa.
   - Nếu **bất kỳ** tag nào đã tồn tại ➡️ **Chặn ngay từ đầu** (không tạo bất kỳ tag nào để tránh trạng thái lửng lơ: tag 1 đã push mà tag 2 bị kẹt), đồng thời hiển thị gợi ý: *(1) Tự động cộng thêm 1 bản patch nữa, hoặc (2) Chỉ tạo & push tag còn thiếu*.
2. **Independent Tag & Push (Tạo xong push liền từng tag):**
   - Loop qua từng tag trong `targetTags`:
     - `git tag -a <tag_1> -m "<msg>"` ➡️ `git push origin <tag_1>` (Đợi hoàn tất để CI/CD bắt đúng webhook của tag 1)
     - `git tag -a <tag_2> -m "<msg>"` ➡️ `git push origin <tag_2>`

---

### 2. Kiến trúc Dynamic Tag Rule cho từng Project (Giải quyết vấn đề 2)

Để công cụ **vừa tiện cho user (không bắt cấu hình thủ công rườm rà), vừa dynamic (đáp ứng mọi biến thể prefix hiện tại và tương lai)**, giải pháp tốt nhất là kết hợp **3 cơ chế: Auto-Detect từ lịch sử Git + Smart Presets + Lưu cấu hình Local theo từng Repo**.

#### A. Lưu Rule của Project ở đâu là tối ưu nhất?
- **Dùng trực tiếp `.git/config` của repo đó** (tương tự cách `minhthetus-cli git account` đang lưu `user.name` và `user.email` vào `--local`):
  - Lưu bằng key riêng:
    - `git config --local minhthetus.tag.sandbox "qc-v,stg-v"` (hoặc `"qc-cicd-v"`)
    - `git config --local minhthetus.tag.prod "release-v"` (hoặc `"v"`)
  - **Ưu điểm vượt trội:**
    - Không sinh ra file rác trong source code của dự án (không sợ quên `.gitignore` hay commit nhầm lên repo công ty).
    - Đi theo repo vĩnh viễn dù đổi tên thư mục hay move folder đi chỗ khác.
- **Kết hợp file Global `~/.minhthetus-cli/git-tag-presets.json`:**
  - Lưu danh sách các **Tag Prefix Presets** thường dùng của user (`qc-v`, `stg-v`, `qc-cicd-v`, `release-v`, `v`). Khi user tự nhập 1 prefix mới hoàn toàn ở 1 dự án (VD: `uat-v`), CLI tự học và lưu thêm `uat-v` vào danh sách preset global để dự án sau có sẵn trong menu gợi ý!

#### B. Cơ chế Smart Auto-Discovery (Lần đầu chạy ở một Project chưa có Rule)
Khi user gõ `minhthetus-cli git tag-dev-stg` ở một repo chưa được set rule:
1. **Quét lịch sử Tag của chính Repo đó (`git tag -l`):**
   - Dùng Regex `^([a-zA-Z0-9_-]*?v?)\d+\.\d+\.\d+$` để tự động bóc tách các prefix đang thực sự được dùng trong repo này (VD: quét thấy repo này toàn dùng `qc-cicd-v1.0.0` và `release-v1.0.0`).
2. **Hiển thị Menu Multi-Select thông minh (Bubble Tea / Gum):**
   - Đưa các prefix **phát hiện được trong repo lên đầu và tick sẵn gợi ý**, bên dưới là các **Preset quen thuộc** + tùy chọn **Custom**:
     ```text
     🔍 Project này chưa có rule tag. Phát hiện các kiểu tag trong repo:
     > [x] qc-cicd-v*  (Phát hiện trong repo - mới nhất: qc-cicd-v1.0.0)
       [ ] qc-v*       (Preset Sandbox)
       [ ] stg-v*      (Preset Sandbox)
       [ ] release-v*  (Phát hiện trong repo - mới nhất: release-v1.10.0)
       [ ] v*          (Preset Prod - vd: v1.0.0)
       [ ] ➕ Nhập prefix mới (Custom)...
     ```
3. **Tự động nhớ cho các lần sau:**
   - Ngay khi user chọn xong (VD: chọn `qc-cicd-v*` hoặc chọn cặp `qc-v*` + `stg-v*`), CLI lưu thẳng vào `git config --local minhthetus.tag.sandbox` của repo đó.
   - Từ lần thứ 2 trở đi, khi gõ `minhthetus-cli git tag-dev-stg`, CLI đọc rule đã lưu và **chạy thẳng luôn** không hỏi lại!

---

### 3. Trải nghiệm người dùng (UX Flow) khi sử dụng hàng ngày

#### Luồng 1: Chạy bình thường (Đã có rule trong project)
Khi gõ `minhthetus-cli git tag-dev-stg`:
```text
ℹ Project Rule: [qc-v*, stg-v*] (dùng --reconfig hoặc -c để đổi rule)
🏷 Latest qc-v* : qc-v1.1.0 (1.1.0)
🏷 Latest stg-v*: stg-v1.1.0 (1.1.0)
ℹ Base version : 1.1.0
```
- **Bước chọn Version Bump:** `patch (1.1.1)` / `minor (1.2.0)` / `major (2.0.0)` / `custom (nhập số bất kỳ)`
- **Bước Pre-flight Check:**
  - `✔ Kiểm tra tag qc-v1.1.1 & stg-v1.1.1: Chưa tồn tại (Hợp lệ)`
- **Bước Tạo & Push độc lập:**
  - `✔ Created & Pushed: qc-v1.1.1`
  - `✔ Created & Pushed: stg-v1.1.1`

#### Luồng 2: Muốn bắn lẻ hoặc đổi Rule bất cứ lúc nào (Dynamic Flags)
Cung cấp các flag linh hoạt cho nhu cầu thực tế:
- `minhthetus-cli git tag-dev-stg --reconfig` (hoặc `-c`): Mở lại bảng Multi-Select để đổi rule của project hiện tại.
- `minhthetus-cli git tag-dev-stg --pick` (hoặc `-p`): Nếu rule của project là cả 2 tag `[qc-v*, stg-v*]`, nhưng hôm nay **chỉ muốn deploy mỗi QC thôi**, thêm `-p` sẽ cho phép tick chọn nhanh chạy 1 trong các prefix của rule mà không làm mất rule gốc.
- `minhthetus-cli git tag-dev-stg --prod` (hoặc tách/gộp menu môi trường `sandbox` vs `prod`):
  - Dùng rule của môi trường **Production** (`v1.0.0` hoặc `release-v1.10.0`).
  - Lần đầu chạy `--prod` cũng tự động auto-detect xem repo đang dùng `v*` hay `release-v*` để gợi ý và lưu riêng vào `minhthetus.tag.prod`.

---

### 4. Tóm tắt cấu trúc Rule Data để Code dễ mở rộng

Mỗi project có thể lưu cấu hình dưới dạng danh sách prefix theo môi trường (Environment Profile):

| Môi trường | Git Local Config Key | Ví dụ Project A | Ví dụ Project B | Ví dụ Project C |
| :--- | :--- | :--- | :--- | :--- |
| **Sandbox (Mặc định)** | `minhthetus.tag.sandbox` | `qc-v,stg-v` | `qc-cicd-v` | `qc-v` |
| **Production (`--prod`)** | `minhthetus.tag.prod` | `v` | `release-v` | `v` |

**Cách tính Base Version khi Rule có nhiều Prefix (VD: `qc-v,stg-v`):**
- Tự động loop qua tất cả prefix trong rule hiện tại: `for _, prefix := range rulePrefixes`
- Lấy latest tag của từng prefix (`getLatestTag(prefix + "*")`), extract số version (`1.1.0`), và tìm `maxVersion` giữa tất cả các prefix trong rule đó.
- Nhờ vậy, dù rule có 1 prefix (`qc-cicd-v`), 2 prefix (`qc-v`, `stg-v`), hay 3 prefix (`dev-v`, `qc-v`, `stg-v`), toàn bộ logic tính version, pre-flight check và sequential push đều chạy hoàn toàn động (100% dynamic loop), không còn bất kỳ dòng code nào hardcode chữ `qc` hay `stg` nữa!

