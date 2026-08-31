# معماری سیستم (Mosaic Architecture)

کتابخانه Mosaic بر اساس یک معماری لایه‌ای، شفاف و تفکیک‌شده طراحی شده است.

---

## جریان اجرایی (Runtime Flow)

```text
Job
 └─ encode.go (موتور هدایت و ارکستراسیون)
    ├─ prepareInputForEncoding (تصحیح چرخش اختیاری)
    ├─ probe.InputWithExecutor (استخراج متادیتا و بررسی صدا)
    ├─ ladder.Build (ساخت رندیشن‌های متناسب با ابعاد)
    ├─ optimize.Apply (اعمال سقف‌های بیت‌ریت و حذف رندیشن‌های اضافی)
    ├─ encryption.SetupKeyInfo (تولید کلید و آماده‌سازی AES-128)
    ├─ encoder.Encode{HLS|DASH}CMAFWithExecutor (پکیجینگ تک‌پاس FFmpeg)
    ├─ thumbnail.GenerateWithExecutor (تولید اسپرایت و WebVTT)
    ├─ subtitles.ProcessTracks (تبدیل و تزریق زیرنویس‌ها)
    └─ storage.UploadDirectory (آپلود موازی به فضای ابری S3)
```

---

## وظایف پکیج‌های داخلی

- **`probe`**: بازرسی متادیتا، چرخش و مشخصات استریم‌ها با FFprobe.
- **`ladder`**: ساخت رندیشن‌های متناسب با نسبت تصویر ورودی.
- **`optimize`**: بهینه‌سازی و سقف‌گذاری بیت‌ریت‌ها و تنظیم بافر VBV.
- **`encoder`**: تولید و اجرای دستورات FFmpeg برای استریم‌های HLS و DASH.
- **`thumbnail`**: تولید تصاویر اسپرایت شبکه‌ای و فایل راهنمای زمانی WebVTT.
- **`preview`**: سرور پیش‌نمایش لوکال به همراه وب‌پلیر با تم تاریک مدرن.
- **`subtitles`**: پردازش و تزریق زیرنویس‌ها به مانیفست‌های استریم.
- **`watermark`**: موقعیت‌دهی و شفاف‌سازی لوگوی واترمارک روی فریم‌ها.
- **`encryption`**: تولید کلیدهای ۱۶ بایتی امن و فایل‌های کانفیگ HLS AES-128.
- **`storage`**: آپلودر چندریسمانی با امضای خالص AWS SigV4 بدون وابستگی خارجی.
- **`internal/executor`**: لایه انتزاعی اجرای فرامین سیستم و جمع‌آوری تله‌متری منابع (CPU Time و RAM).
