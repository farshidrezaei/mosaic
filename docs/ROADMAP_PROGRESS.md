# Mosaic Development & Roadmap Progress

این سند وضعیت پیشرفت و نقشه راه توسعه قابلیت‌های جدید و پیشرفته در پروژه **Mosaic** را نشان می‌دهد.

---

## 📌 وضعیت کلی فازها (Overall Phase Status)

| فاز | عنوان | وضعیت | پیشرفت |
|:---:|---|:---:|:---:|
| **۱** | پیش‌نمایش، Sprite تامبنیل، Trick-Play و DevTools | 🟢 تکمیل شد | ۱۰۰٪ |
| **۲** | زیرنویس، چندصدایی و نرمال‌سازی صدا | 🟢 تکمیل شد | ۱۰۰٪ |
| **۳** | واترمارک داینامیک، برندینگ و تیزرها | 🟢 تکمیل شد | ۱۰۰٪ |
| **۴** | کدک‌های نسل جدید (AV1/HEVC) و کیفیت هوشمند (CRF/VMAF) | 🟢 تکمیل شد | ۱۰۰٪ |
| **۵** | امنیت و رمزنگاری (HLS AES-128 & DRM Ready) | 🟢 تکمیل شد | ۱۰۰٪ |
| **۶** | آپلود مستقیم ابری (S3/MinIO) و رویدادها (Webhooks) | 🟢 تکمیل شد | ۱۰۰٪ |
| **۷** | پرتال مستندات هوشمند VitePress (i18n + SEO + ABR Simulator + Player) | 🟢 تکمیل شد | ۱۰۰٪ |

---

## 🔍 چک‌لیست و جزئیات پیشرفت تسک‌ها

### 🚀 فاز ۱: پیش‌نمایش و تامبنیل (Visual Preview & DevTools) — [تکمیل شد ✅]
- [x] **P1.1: پکیج `thumbnail` و تولید خودکار Sprite Sheet & WebVTT**
  - [x] ساخت پکیج `thumbnail` با تنظیمات فاصله زمانی، ابعاد کاشی و شبکه
  - [x] پیاده‌سازی لاجیک FFmpeg برای تولید Sprite Sheet
  - [x] تولید خودکار فایل استاندارد `thumbnails.vtt` با مختصات `#xywh=x,y,w,h`
  - [x] افزودن Optionهای تابعی `WithThumbnails(...)` به روت `mosaic`
  - [x] تست‌های جامع واحد با Mock Executor
- [x] **P1.2: سرور پیش‌نمایش و دستور CLI `mosaic preview`**
  - [x] ساخت پکیج `preview` و راه‌اندازی سرور HTTP لوکال با قابلیت CORS
  - [x] تعبیه وب‌پلیر مدرن و دارک مود (HLS.js / Dash.js) با پشتیبانی از تامبنیل و سوئیچ رندیشن
  - [x] افزودن زیرفرمان `mosaic preview` به `cmd/mosaic`
  - [x] تست‌های واحد سرور پیش‌نمایش
- [x] **P1.3: پلی‌لیست Trick-Play (I-Frames Only)**
  - [x] تولید خودکار پلی‌لیست‌های `#EXT-X-I-FRAMES-ONLY` برای HLS با گزینه `WithIFrames()`

---

### 💬 فاز ۲: زیرنویس و مدیریت صدای پیشرفته (Subtitles & Multi-Audio) — [تکمیل شد ✅]
- [x] **P2.1: مدیریت ترک‌های زیرنویس (WebVTT / SRT)**
  - [x] پکیج `subtitles` برای تبدیل خودکار SRT به WebVTT
  - [x] تزریق `#EXT-X-MEDIA:TYPE=SUBTITLES` در پلی‌لیست‌های HLS
  - [x] تزریق AdaptationSet زیرنویس در DASH Manifest (`manifest.mpd`)
  - [x] افزودن گزینه `WithSubtitles(...)` به API کتابخانه
- [x] **P2.2: تصحیح و هماهنگ‌سازی نسبت ابعاد (Aspect Ratio Fix)**
  - [x] حل خطای تطابق DAR در Adaptationsetهای DASH و رندیشن‌های پرتره
- [x] **P2.3: استانداردسازی بلندی صدا (Audio Normalization - EBU R128)**
  - [x] فعال‌سازی فیلتر خودکار `loudnorm=I=-16:TP=-1.5:LRA=11` با گزینه `WithNormalizeAudio()` و فلگ `--normalize-audio`

---

### 🖼️ فاز ۳: برندینگ و واترمارک (Branding & Overlays) — [تکمیل شد ✅]
- [x] **P3.1: واترمارک داینامیک (Watermark / Logo Overlay)**
  - [x] پکیج `watermark` با موقعیت‌های استاندارد (Top-Right, Top-Left, Bottom-Right, Bottom-Left, Center)
  - [x] تنظیم خودکار مقیاس بر اساس عرض هر رندیشن و اعمال شفافیت (Opacity)
  - [x] افزودن گزینه `WithWatermark(cfg WatermarkConfig)` و فلگ CLI `--watermark`

---

### 🧬 فاز ۴: کدک‌های نسل جدید و انکودینگ هوشمند (Next-Gen Codecs & Smart Encoding) — [تکمیل شد ✅]
- [x] **P4.1: پشتیبانی از کدک‌های AV1 و HEVC/H.265**
  - [x] پشتیبانی از انکودرهای نرم‌افزاری (`libsvtav1`, `libx265`) و سخت‌افزاری GPU (`nvenc`, `vaapi`, `videotoolbox`)
  - [x] گزینه‌های تابعی `WithCodec()`, `WithHEVC()`, `WithAV1()` و فلگ CLI `--codec`
- [x] **P4.2: حالت بهینه‌سازی بر پایه محتوا (CRF / Content-Aware Bitrate)**
  - [x] انکود بر پایه کیفیت هدف با حفظ محدودیت سقف پهنای باند (`-maxrate` و `-bufsize`)
  - [x] گزینه تابعی `WithCRF(int)` و فلگ CLI `--crf`

---

### 🔒 فاز ۵: امنیت، رمزنگاری و مدیریت دسترسی (Security & DRM) — [تکمیل شد ✅]
- [x] **P5.1: رمزنگاری HLS AES-128 خودکار**
  - [x] پکیج `encryption` برای تولید خودکار کلید امن ۱۶ بایتی و فایل `enc.keyinfo`
  - [x] رمزنگاری سگمنت‌های HLS و درج تگ `#EXT-X-KEY:METHOD=AES-128` در پلی‌لیست‌ها
  - [x] افزودن گزینه `WithAES128Encryption(...)` و فلگ CLI `--encrypt-aes128`

---

### ☁️ فاز ۶: سرویس‌های ابری و آپلود مستقیم (Cloud & Storage) — [تکمیل شد ✅]
- [x] **P6.1: آپلودر مستقیم به Cloud Storage (S3 / MinIO / Cloudflare R2)**
  - [x] پکیج سبک `storage` بدون هیچ وابستگی خارجی با پیاده‌سازی خالص AWS Signature Version 4
  - [x] آپلود همزمان (Concurrent Worker Pool) فایل‌های استریم
  - [x] تنظیم خودکار هدرهای بهینه `Content-Type` و `Cache-Control` (مانند `immutable` برای سگمنت‌ها و `no-cache` برای مانیفست‌ها)
  - [x] گزینه `WithS3Upload(cfg S3Config)` و فلگ‌های CLI مربوط به S3

---

### 🌐 فاز ۷: پرتال مستندات هوشمند و نسل جدید (Next-Gen Docs Portal) — [تکمیل شد ✅]
- [x] **P7.1: مهاجرت کامل به VitePress (Vue 3) و سئوی ۱۰۰٪**
  - [x] تولید استاتیک SSG تمام صفحات با لود آنی زیر ۵۰ میلی‌ثانیه و Sitemap
  - [x] سیستم دوزبانه جامع (English LTR و فارسی راست‌چین RTL با فونت Vazirmatn)
  - [x] منوی ورژنینگ و انتخاب نسخه در هدر
  - [x] کامپوننت هوشمند ماشین‌حساب نردبان کیفیت ABR (`<AbrCalculator />`)
  - [x] کامپوننت پلیر استریم آنلاین و لوکال با انتخاب فایل `master.m3u8` (`<StreamPlayer />`)
  - [x] طراحی لوکس شیشه‌ای Liquid Glass با پس‌زمینه نوری Aurora

---

*تمامی فازهای نقشه راه به صورت ۱۰۰٪ کامل پیاده‌سازی و تست شدند.*

