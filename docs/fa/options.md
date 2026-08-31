# گزینه‌های پیکربندی (Functional Options)

کتابخانه Mosaic از الگوی طراحی **Functional Options** برای سفارشی‌سازی منعطف، نوع‌امن و ماژولار خط‌لوله پکیجینگ ویدیو استفاده می‌کند.

---

## ۱. کدک‌های نسل جدید و کنترل کیفیت (Codecs & CRF)

```go
func WithCodec(codec VideoCodec) Option
func WithHEVC() Option
func WithAV1() Option
func WithCRF(crf int) Option
```

- **`WithCodec(codec)`**: انتخاب انکودر ویدیویی هدف:
  - `mosaic.CodecH264`: کدک جهان‌شمول H.264 / AVC (`libx264`, `h264_nvenc`, `h264_vaapi`, `h264_videotoolbox`).
  - `mosaic.CodecHEVC`: کدک مدرن H.265 / HEVC (`libx265`, `hevc_nvenc`, `hevc_vaapi`, `hevc_videotoolbox`).
  - `mosaic.CodecAV1`: کدک نسل جدید و رایگان AV1 (`libsvtav1`, `av1_nvenc`, `av1_vaapi`).
- **`WithCRF(crf)`**: فعال‌سازی تکنیک **Capped-CRF**. در این روش FFmpeg کیفیت بصری را با مقدار ثابت حفظ می‌کند اما در صحنه‌های شلوغ و پرحرکت به سقف‌های `-maxrate` و `-bufsize` مقید می‌ماند تا ۳۰ تا ۵۰ درصد پهنای باند صرفه‌جویی شود.

```go
mosaic.WithAV1(),
mosaic.WithCRF(28),
```

---

## ۲. استوری‌بورد و تامبنیل اسکرابر (Storyboards)

```go
func WithThumbnails(cfg ...ThumbnailConfig) Option
```

- **عملکرد**: تولید یک تصویر شبکه‌ای فشرده JPEG (`thumbnails_0.jpg`) و فایل راهنمای زمانی استاندار WebVTT (`thumbnails.vtt`) با برچسب‌های مختصات `#xywh=x,y,w,h` جهت نمایش پیش‌نمایش ویدیو هنگام جابجایی روی نوار تایم‌لاین پلیر.

```go
mosaic.WithThumbnails(mosaic.ThumbnailConfig{
    IntervalSeconds: 2,   // عکس‌برداری هر ۲ ثانیه یکبار
    TileWidth:       160, // عرض هر فریم تامبنیل
    TileHeight:      90,  // ارتفاع هر فریم تامبنیل
    Columns:         5,   // ۵ ستون در هر فایل اسپرایت
    Rows:            5,
    Quality:         3,   // کیفیت عالی JPEG
})
```

---

## ۳. زیرنویس و نرمال‌سازی صدا (Subtitles & Loudnorm)

```go
func WithSubtitles(tracks ...SubtitleTrack) Option
func WithNormalizeAudio(enabled ...bool) Option
```

- **`WithSubtitles`**: دریافت یک یا چند فایل زیرنویس (`.srt` یا `.vtt`). فایل‌های SRT به طور خودکار به WebVTT تبدیل شده و با شناسه گروهی در مانیفست HLS (`#EXT-X-MEDIA:TYPE=SUBTITLES`) و DASH (`<AdaptationSet contentType="text">`) قرار می‌گیرند.
- **`WithNormalizeAudio`**: تنظیم و یکسان‌سازی بلندی صدا در تمام استریم بر اساس استاندارد جهانی **EBU R128 (`loudnorm=I=-16:TP=-1.5:LRA=11`)**.

```go
mosaic.WithNormalizeAudio(),
mosaic.WithSubtitles(
    mosaic.SubtitleTrack{
        Path:     "./subs/fa.srt",
        Language: "fa",
        Label:    "فارسی",
        Default:  true,
    },
    mosaic.SubtitleTrack{
        Path:     "./subs/en.srt",
        Language: "en",
        Label:    "English",
    },
),
```

---

## ۴. واترمارک و برندینگ داینامیک (Watermark)

```go
func WithWatermark(cfg WatermarkConfig) Option
```

- **عملکرد**: درج یکپارچه لوگو یا نشان برند (فرمت PNG یا WebP) روی تمام رندیشن‌های نردبان بدون تغییر نسبت ابعاد و بدون نیاز به پیش‌رندر ویدیویی.
- **موقعیت‌ها**: `PositionTopRight`، `PositionTopLeft`، `PositionBottomRight`، `PositionBottomLeft`، `PositionCenter`.

```go
mosaic.WithWatermark(mosaic.WatermarkConfig{
    Path:     "./assets/logo.png",
    Position: mosaic.PositionTopRight,
    OffsetX:  20,
    OffsetY:  20,
    Opacity:  0.80, // ۸۰ درصد شفافیت
})
```

---

## ۵. رمزنگاری امن سگمنت‌های HLS (AES-128)

```go
func WithAES128Encryption(cfg ...EncryptionConfig) Option
```

- **عملکرد**: تولید کلید رمزنگاری تصادفی ۱۶ بایتی امن (`enc.key`)، ساخت فایل راهنما (`enc.keyinfo`) و رمزنگاری تمامی سگمنت‌های مدیا با الگوریتم استاندارد AES-128 و برچسب‌گذاری `#EXT-X-KEY`.

---

## ۶. آپلود مستقیم ابری به S3 / MinIO / R2

```go
func WithS3Upload(cfg S3Config) Option
```

- **عملکرد**: آپلود بی‌درنگ و چندریسمانی تمامی فایل‌های استریم به باکت‌های سازگار با AWS S3 با استفاده از امضای خالص AWS SigV4 بدون وابستگی به پکیج‌های خارجی.

```go
mosaic.WithS3Upload(mosaic.S3Config{
    Endpoint:        "https://s3.us-east-1.amazonaws.com",
    Bucket:          "my-stream-bucket",
    Region:          "us-east-1",
    KeyPrefix:       "vod/movie-101",
    AccessKey:       "YOUR_ACCESS_KEY",
    SecretKey:       "YOUR_SECRET_KEY",
    ConcurrentFiles: 6,
})
```
