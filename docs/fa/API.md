# مرجع کامل رابط برنامه‌نویسی (Public API Reference)

این سند راهنمای کامل انواع داده‌ها، ساختارها، توابع و گزینه‌های پیکربندی پکیج `github.com/farshidrezaei/mosaic` به زبان فارسی است.

---

## ۱. ایمپورت پکیج

```go
import "github.com/farshidrezaei/mosaic"
```

---

## ۲. ساختارهای اصلی (Core Structs)

### ساختار Job
```go
type Job struct {
	Input           string          // آدرس فایل ورودی محلی یا URL راه دور (HTTP/HTTPS)
	OutputDir       string          // مسیر پوشه ذخیره‌سازی مانیفست‌ها و سگمنت‌ها
	Profile         Profile         // پروفایل ویدیو: ProfileVOD یا ProfileLive
	ProgressHandler ProgressHandler // تابع کالبک دریافت پیشرفت بلادرنگ
}
```

### ساختار ProgressInfo
```go
type ProgressInfo struct {
	CurrentTime string  // زمان سپری شده (مثال: "00:01:23.456000")
	Bitrate     string  // بیت‌ریت لحظه‌ای انکود (مثال: "2450.3kbits/s")
	Speed       string  // ضریب سرعت انکودینگ (مثال: "2.15x")
	Percentage  float64 // درصد دقیق پیشرفت از 0.0 تا 100.0
}
```

### نوع داده کدک‌های ویدیویی (VideoCodec)
```go
type VideoCodec = config.VideoCodec

const (
	CodecH264 VideoCodec = "h264" // کدک استاندارد H.264 / AVC
	CodecHEVC VideoCodec = "hevc" // کدک با راندمان بالا H.265 / HEVC
	CodecAV1  VideoCodec = "av1"  // کدک باز و نسل جدید AV1
)
```

### ساختار تنظیمات تامبنیل (ThumbnailConfig)
```go
type ThumbnailConfig struct {
	SpriteFilename  string // الگوی نام فایل اسپرایت (پیش‌فرض: "thumbnails_%d.jpg")
	VTTFilename     string // نام فایل راهنمای WebVTT (پیش‌فرض: "thumbnails.vtt")
	IntervalSeconds int    // فاصله زمانی ثبت تامبنیل‌ها به ثانیه (پیش‌فرض: ۲)
	TileWidth       int    // عرض هر فریم تامبنیل به پیکسل (پیش‌فرض: ۱۶۰)
	TileHeight      int    // ارتفاع هر فریم تامبنیل به پیکسل (پیش‌فرض: ۹۰)
	Columns         int    // تعداد ستون‌های شبکه اسپرایت (پیش‌فرض: ۵)
	Rows            int    // تعداد سطرهای شبکه اسپرایت (پیش‌فرض: ۵)
	Quality         int    // کیفیت تصویر JPEG از ۱ تا ۳۱ (پیش‌فرض: ۳)
}
```

### ساختار واترمارک و موقعیت‌ها (WatermarkConfig)
```go
type WatermarkConfig struct {
	Path     string            // مسیر تصویر لوگو (PNG یا WebP)
	Position WatermarkPosition // موقعیت قرارگیری روی فریم
	OffsetX  int               // فاصله افقی از لبه تصویر به پیکسل (پیش‌فرض: ۲۰)
	OffsetY  int               // فاصله عمودی از لبه تصویر به پیکسل (پیش‌فرض: ۲۰)
	Opacity  float64           // میزان شفافیت آلفا از 0.0 تا 1.0 (پیش‌فرض: 1.0)
}

const (
	PositionTopRight    WatermarkPosition = "top-right"    // بالا-راست
	PositionTopLeft     WatermarkPosition = "top-left"     // بالا-چپ
	PositionBottomRight WatermarkPosition = "bottom-right" // پایین-راست
	PositionBottomLeft  WatermarkPosition = "bottom-left"  // پایین-چپ
	PositionCenter      WatermarkPosition = "center"       // وسط تصویر
)
```

### ساختار تراک زیرنویس (SubtitleTrack)
```go
type SubtitleTrack struct {
	Path     string // مسیر فایل زیرنویس (.srt یا .vtt)
	Language string // کد زبان استاندارد ایزو (مثال: "fa", "en")
	Label    string // عنوان نمایشی در پلیر (مثال: "فارسی", "English")
	Default  bool   // آیا این تراک به صورت پیش‌فرض فعال باشد
	Forced   bool   // آیا این زیرنویس از نوع اجباری (Forced) است
}
```

### ساختار رمزنگاری HLS (EncryptionConfig)
```go
type EncryptionConfig struct {
	KeyURI string // آدرس URI کلید در پلی‌لیست HLS (پیش‌فرض: "enc.key")
	IV     string // بردار اولیه هگزادسیمال اختیاری ۳۲ کاراکتری
	Key    []byte // کلید ۱۶ بایتی امن AES (در صورت nil بودن خودکار تولید می‌شود)
}
```

### ساختار اتصال ابری S3 (S3Config)
```go
type S3Config struct {
	Endpoint        string // آدرس سرور S3 یا MinIO یا R2
	Bucket          string // نام باکت مقصد
	Region          string // ریجن AWS (پیش‌فرض: "us-east-1")
	AccessKey       string // شناسه دسترسی Access Key
	SecretKey       string // کلید امنیتی Secret Key
	KeyPrefix       string // پیشوند مسیر کلید اشیاء (مثال: "vod/streams/movie1")
	ConcurrentFiles int    // تعداد ورکر‌های آپلود همزمان (پیش‌فرض: ۵)
	UseSSL          bool   // استفاده از پروتکل امن HTTPS (پیش‌فرض: true)
}
```

---

## ۳. توابع اصلی پکیجینگ

### تابع EncodeHls
```go
func EncodeHls(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
```
ویدیو را به استریم چندکیفیته HLS به همراه مانیفست اصلی `master.m3u8` و سگمنت‌های فریم‌بندی‌شده `fMP4` یا `TS` تبدیل می‌کند.

### تابع EncodeDash
```go
func EncodeDash(ctx context.Context, job Job, opts ...Option) (*executor.Usage, error)
```
ویدیو را به استریم چندکیفیته MPEG-DASH بر اساس ساختار CMAF به همراه فایل مانیفست `manifest.mpd` تبدیل می‌کند.

---

## ۴. جدول گزینه‌های پیکربندی (Functional Options)

| گزینه پیکربندی | عملکرد و کاربرد |
|---|---|
| `mosaic.WithCodec(codec)` | تعیین فرمت فشرده‌سازی ویدیو (`CodecH264`, `CodecHEVC`, `CodecAV1`). |
| `mosaic.WithHEVC()` | میانبر فعال‌سازی کدک H.265 / HEVC. |
| `mosaic.WithAV1()` | میانبر فعال‌سازی کدک نسل جدید AV1. |
| `mosaic.WithCRF(crf)` | فعال‌سازی بهینه‌سازی Capped-CRF (کیفیت ثابت با سقف بیت‌ریت). |
| `mosaic.WithThumbnails(cfg...)` | تولید اسپرایت تامبنیل و فایل نشانه‌گذاری زمانی `thumbnails.vtt`. |
| `mosaic.WithSubtitles(tracks...)` | تبدیل خودکار SRT به WebVTT و تزریق به پلی‌لیست‌های HLS و DASH. |
| `mosaic.WithNormalizeAudio()` | یکسان‌سازی خودکار بلندی صدا با استاندارد برودکست EBU R128 (`loudnorm`). |
| `mosaic.WithWatermark(cfg)` | درج لوگو/واترمارک داینامیک با موقعیت‌دهی و شفافیت اختصاصی. |
| `mosaic.WithAES128Encryption(cfg...)` | رمزنگاری امن سگمنت‌های HLS با کلید ۱۶ بایتی AES-128. |
| `mosaic.WithS3Upload(cfg)` | آپلود مستقیم و همزمان خروجی به فضای ابری S3 / MinIO / Cloudflare R2. |
| `mosaic.WithIFrames()` | تولید پلی‌لیست‌های تریک‌پلی فریم‌های I (`#EXT-X-I-FRAMES-ONLY`). |
| `mosaic.WithNormalizeOrientation()` | بازرسی چرخش و تصحیح فیزیکی ویدیوهای ضبط‌شده با موبایل. |
| `mosaic.WithThreads(n)` | تعیین تعداد ریسمان‌های پردازنده برای انکودینگ FFmpeg. |
| `mosaic.WithBFrames(n)` | تنظیم تعداد فریم‌های B متوالی برای رندیشن‌ها. |
| `mosaic.WithScaleBitrateWithFPS()` | افزایش هوشمند سقف بیت‌ریت برای ویدیوهای با نرخ فریم بالا (>30 FPS). |
| `mosaic.WithNVENC()` | استفاده از شتاب‌دهنده سخت‌افزاری کارت‌های گرافیک NVIDIA. |
| `mosaic.WithVAAPI()` | استفاده از شتاب‌دهنده سخت‌افزاری Intel و AMD. |
| `mosaic.WithVideoToolbox()` | استفاده از شتاب‌دهنده سخت‌افزاری اپل در سیستم‌عامل macOS. |
