# عیب‌یابی و پرسش‌های متداول (Troubleshooting & FAQ)

---

### ۱. پخش نشدن فایل‌های HLS رمزنگاری‌شده در پلیرهای محلی (ffplay / mpv)
**علت**: به دلیل سیاست‌های امنیتی سندباکس FFmpeg، باز کردن فایل کلید با پسوند `enc.key` از مسیر نسبی در فایل سیستم محلی مسدود می‌شود (`Filename extension of 'enc.key' is not a common multimedia extension`).
**راه‌حل**:
- هنگام تست با ffplay پارامتر `-allowed_extensions ALL` را پاس دهید:
  ```bash
  ffplay -allowed_extensions ALL ./output/hls/master.m3u8
  ```
- یا استریم را روی سرور وب لوکال با دستور `mosaic preview` اجرا و تست کنید:
  ```bash
  mosaic preview ./output/hls
  ```

---

### ۲. خطای عدم تطابق Aspect Ratio در DASH (`conflicting stream aspect ratios`)
**علت**: استاندارد DASH ایزو برابری کامل نسبت تصویر در تمام رندیشن‌ها را الزامی می‌داند.
**راه‌حل**: در نسخه‌های ۱.۸.۰ به بعد Mosaic، نسبت DAR با پارامتر `setdar=W/H` در فیلترگراف به صورت خودکار محاسبه و تثبیت شده است.

---

### ۳. کندی انکودینگ روی پردازنده‌های ضعیف
**راه‌حل**:
- از شتاب‌دهنده‌های سخت‌افزاری کارت گرافیک (`WithNVENC()`, `WithVAAPI()`, `WithVideoToolbox()`) یا فلگ `--gpu nvenc` استفاده کنید.
- تعداد ریسمان‌های پردازنده را با `WithThreads(4)` به صورت دستی مشخص کنید.
