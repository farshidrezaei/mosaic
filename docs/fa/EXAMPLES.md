# کاتالوگ نمونه‌کدها (Examples)

تمامی نمونه‌کدهای زیر به صورت کامل و آماده اجرا در پوشه [`examples/`](https://github.com/farshidrezaei/mosaic/tree/main/examples) مخزن گیت‌هاب قرار دارند. برای هر مورد، هم شیوه اجرا در کد Go و هم دستور معادل خط فرمان (CLI) آورده شده است.

---

## ۱. پکیجینگ ساده HLS (Simple HLS)
**مسیر**: `examples/simple_hls/main.go`

پکیجینگ ویدیوی VOD به همراه گزارش لحظه‌ای درصد پیشرفت و نرمال‌سازی خودکار چرخش ویدیو.

### اجرا با Go:
```bash
cd examples/simple_hls && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
mosaic -i input.mp4 -o ./output/hls --normalize
```

---

## ۲. پکیجینگ پیشرفته DASH CMAF (Advanced DASH)
**مسیر**: `examples/advanced_dash/main.go`

استریم DASH CMAF با تنظیم فریم‌های B (`WithBFrames(2)`)، مقیاس بیت‌ریت بر پایه نرخ فریم (`WithScaleBitrateWithFPS()`) و تعداد ریسمان‌های سفارشی پردازنده.

### اجرا با Go:
```bash
cd examples/advanced_dash && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
mosaic -i input.mp4 -o ./output/dash --format=dash --bframes=2 --threads=4
```

---

## ۳. اسکرابر تامبنیل و پلیر پیش‌نمایش (Thumbnails & Web Preview)
**مسیر**: `examples/thumbnails_and_preview/main.go`

تولید خودکار اسپرایت‌های تایم‌لاین (`thumbnails_0.jpg`)، فایل نشانه‌گذاری زمانی (`thumbnails.vtt`) و اجرای وب‌پلیر لوکال با تم تاریک مدرن.

### اجرا با Go:
```bash
cd examples/thumbnails_and_preview && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
# ۱. پکیجینگ ویدیو به همراه تامبنیل‌ها
mosaic -i input.mp4 -o ./output/hls --thumbnails

# ۲. اجرای پلیر پیش‌نمایش تحت وب روی پورت 8080
mosaic preview ./output/hls
```

---

## ۴. واترمارک داینامیک، زیرنویس و نرمال‌سازی صدا
**مسیر**: `examples/watermark_and_subtitles/main.go`

درج لوگوی شفاف با موقعیت‌دهی اختصاصی، استانداردسازی بلندی صدا بر اساس EBU R128 و تبدیل خودکار زیرنویس SRT به WebVTT.

### اجرا با Go:
```bash
cd examples/watermark_and_subtitles && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
mosaic -i input.mp4 -o ./output/hls_branded \
  --watermark ./branding/logo.png \
  --subtitles ./subtitles/fa.srt,./subtitles/en.srt \
  --normalize-audio
```

---

## ۵. رمزنگاری امن HLS AES-128
**مسیر**: `examples/encryption_aes128/main.go`

تولید خودکار کلیدهای امن ۱۶ بایتی (`enc.key`)، ساخت فایل‌های کانفیگ و رمزنگاری سگمنت‌های HLS با برچسب استاندارد `#EXT-X-KEY`.

### اجرا با Go:
```bash
cd examples/encryption_aes128 && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
# پکیجینگ با رمزنگاری سگمنت‌ها
mosaic -i input.mp4 -o ./output/hls_encrypted --encrypt-aes128

# پخش استریم رمزنگاری‌شده در پلیر پیش‌نمایش
mosaic preview ./output/hls_encrypted
```

---

## ۶. کدک‌های نسل جدید AV1 و HEVC (بهینه‌سازی Capped-CRF)
**مسیر**: `examples/nextgen_av1_hevc/main.go`

انکودینگ ویدیو با راندمان بالا با استفاده از کدک‌های **AV1** (`libsvtav1`) و **HEVC** (`libx265`) به همراه بهینه‌سازی کیفیت بر پایه محتوا برای کاهش ۳۰ تا ۵۰ درصدی حجم ویدیو.

### اجرا با Go:
```bash
cd examples/nextgen_av1_hevc && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
# انکود با کدک نسل جدید AV1 و CRF 28
mosaic -i input.mp4 -o ./output/hls_av1 --codec=av1 --crf=28

# انکود با کدک HEVC (H.265) و CRF 23
mosaic -i input.mp4 -o ./output/hls_hevc --codec=hevc --crf=23
```

---

## ۷. آپلود مستقیم ابری به S3 و MinIO
**مسیر**: `examples/s3_cloud_upload/main.go`

آپلود موازی و همزمان تمامی فایل‌های استریم تولیدشده (`.m3u8`, `.mpd`, `.m4s`, `.vtt`, `.jpg`) به فضاهای ابری S3 / MinIO / Cloudflare R2 با استفاده از امضای خالص AWS SigV4 بدون وابستگی خارجی.

### اجرا با Go:
```bash
cd examples/s3_cloud_upload && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
mosaic -i input.mp4 -o ./output/hls \
  --s3-bucket my-stream-bucket \
  --s3-prefix vod/movie-101 \
  --s3-region us-east-1 \
  --s3-endpoint https://s3.us-east-1.amazonaws.com
```

---

## ۸. استریم پخش زنده کم‌تاخیر (Live Low-Latency)
**مسیر**: `examples/live_streaming/main.go`

تولید استریم‌های زنده با استفاده از `ProfileLive`، سگمنت‌های ۲ ثانیه‌ای و فلگ‌های Low-Latency CMAF برای HLS و DASH.

### اجرا با Go:
```bash
cd examples/live_streaming && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
mosaic -i input.mp4 -o ./output/hls_live --profile=live
```

---

## ۹. نرمال‌سازی چرخش ویدیوهای موبایل
**مسیر**: `examples/orientation_normalization/main.go`

بازرسی متادیتای ماتریس چرخش، چرخش فیزیکی فریم‌های ویدیو و پاک‌سازی تگ چرخش در خروجی جهت جلوگیری از پخش دفرمه در پلیرها.

### اجرا با Go:
```bash
cd examples/orientation_normalization && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
mosaic -i mobile_video.mp4 -o ./output/hls_rotated --normalize
```

---

## ۱۰. نوار پیشرفت گرافیکی در ترمینال
**مسیر**: `examples/progress_monitoring/main.go`

رندر نوار پیشرفت یونیکد زیبا در خط فرمان (`[██████░░░░] 60.0%`) با نمایش سرعت، بیت‌ریت، زمان سپری‌شده و مصرف رم (RSS).

### اجرا با Go:
```bash
cd examples/progress_monitoring && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
mosaic -i input.mp4 -o ./output/hls --log-level=info
```

---

## ۱۱. شتاب‌دهنده‌های سخت‌افزاری کارت گرافیک (Multi-GPU)
**مسیر**: `examples/multi_gpu/main.go`

انکود پرسرعت با استفاده از کارت‌های گرافیک NVIDIA NVENC، اینتل و AMD (VAAPI) و اپل (VideoToolbox).

### اجرا با Go:
```bash
cd examples/multi_gpu && go run main.go
```

### دستور معادل خط فرمان (CLI):
```bash
# کارت گرافیک NVIDIA
mosaic -i input.mp4 -o ./output/hls --gpu=nvenc

# کارت گرافیک Intel / AMD (VAAPI)
mosaic -i input.mp4 -o ./output/hls --gpu=vaapi

# شتاب‌دهنده اپل در سیستم‌عامل macOS
mosaic -i input.mp4 -o ./output/hls --gpu=videotoolbox
```
