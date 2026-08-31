# استراتژی تست و کنترل کیفیت (Testing Guide)

کتابخانه Mosaic از معماری ۱۰۰٪ مبتنی بر اینترفیس و تزریق وابستگی (`CommandExecutor`) بهره می‌برد که امکان اجرای سریع و بدون سربار تمامی تست‌ها را بدون نیاز به فراخوانی FFmpeg واقعی فراهم می‌کند.

---

## دستورات استاندارد اجرای تست‌ها

```bash
# اجرای تمامی تست‌های واحد به همراه تشخیص رقابت حافظه (Race Detector)
GOCACHE=/tmp/go-build go test -v -race ./...

# بازرسی استاتیک کدها
GOCACHE=/tmp/go-build go vet ./...

# اجرای لینتر پیشرفته (قانون صفر خطای لینتر)
golangci-lint run
```

---

## تست با استفاده از MockCommandExecutor

```go
mockExec := executor.NewMockCommandExecutor()
mockExec.AddResponse("ffprobe", []byte(`{ "streams": [...] }`), nil)

usage, err := mosaic.EncodeHlsWithExecutor(ctx, job, mockExec)
```
