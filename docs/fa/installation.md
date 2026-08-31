# نصب و پیش‌نیازها (Installation)

## پیش‌نیازهای سیستمی
- **Go**: نسخه `1.25` یا بالاتر
- **FFmpeg**: نسخه `4.4` یا بالاتر (با پشتیبانی از `libx264`, `libx265`, `libsvtav1` و `aac`)
- **FFprobe**: ابزار بازرسی متادیتا (معمولاً همراه FFmpeg نصب می‌شود)

---

## ۱. نصب به عنوان کتابخانه در پروژه Go

```bash
go get github.com/farshidrezaei/mosaic
```

---

## ۲. نصب به عنوان ابزار خط فرمان (CLI)

```bash
go install github.com/farshidrezaei/mosaic/cmd/mosaic@latest
```

---

## ۳. اجرای مستقیم با داکر (Docker)

```bash
docker run --rm -v $(pwd):/workspace ghcr.io/farshidrezaei/mosaic -i input.mp4 -o ./output/hls --thumbnails
```
