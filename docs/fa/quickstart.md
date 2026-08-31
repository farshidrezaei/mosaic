# شروع سریع (Quick Start)

با کتابخانه و ابزار خط فرمان **Mosaic** می‌توانید در عرض چند دقیقه ویدیوهای خود را به استریم‌های استاندارد تطبیقی ABR تبدیل کنید.

---

## ۱. نصب و راه‌اندازی

### نصب کتابخانه در پروژه Go
```bash
go get github.com/farshidrezaei/mosaic
```

### نصب به عنوان ابزار خط فرمان (CLI)
```bash
# نصب مستقیم با Go
go install github.com/farshidrezaei/mosaic/cmd/mosaic@latest

# یا اجرا با داکر (به همراه FFmpeg از پیش نصب‌شده)
docker run --rm -v $(pwd):/workspace ghcr.io/farshidrezaei/mosaic -i input.mp4 -o ./output/hls --thumbnails
```

---

## ۲. استفاده از خط فرمان (CLI)

### پکیجینگ ساده HLS به همراه اسکرابر تامبنیل:
```bash
mosaic -i input.mp4 -o ./output/hls --thumbnails
```

### انکودینگ با کدک پیشرفته AV1، بهینه‌سازی Capped-CRF، واترمارک و نرمال‌سازی صدا:
```bash
mosaic -i input.mp4 -o ./output/hls_av1 \
  --codec av1 \
  --crf 28 \
  --watermark ./branding/logo.png \
  --normalize-audio \
  --thumbnails
```

### پکیجینگ امن با رمزنگاری HLS AES-128:
```bash
mosaic -i input.mp4 -o ./output/hls_encrypted --encrypt-aes128
```

### پکیجینگ و آپلود مستقیم ابری به S3 / MinIO:
```bash
mosaic -i input.mp4 -o ./output/hls_s3 \
  --s3-bucket my-stream-bucket \
  --s3-prefix vod/video-101 \
  --s3-region us-east-1
```

### اجرای پلیر تست لوکال (Web Preview Player):
```bash
mosaic preview ./output/hls
```
سپس آدرس **`http://localhost:8080`** را در مرورگر باز کنید تا استریم تولیدشده را همراه با تغییر کیفیت زنده و تامبنیل‌ها تست کنید!

---

## ۳. استفاده به عنوان کتابخانه در کد Go

### نمونه کامل فرآیند تولید استریم HLS در پروداکشن:

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/hls",
		Profile:   mosaic.ProfileVOD,
		ProgressHandler: func(info mosaic.ProgressInfo) {
			fmt.Printf("\r[%5.1f%%] time=%s bitrate=%s speed=%s",
				info.Percentage, info.CurrentTime, info.Bitrate, info.Speed)
		},
	}

	usage, err := mosaic.EncodeHls(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(), // چرخش خودکار ویدیوهای موبایل ۹۰ و ۲۷۰ درجه
		mosaic.WithNormalizeAudio(),       // استانداردسازی بلندی صدا طبق EBU R128
		mosaic.WithThumbnails(),           // تولید اسپرایت تامبنیل و فایل thumbnails.vtt
		mosaic.WithWatermark(mosaic.WatermarkConfig{
			Path:     "./branding/logo.png",
			Position: mosaic.PositionTopRight,
			Opacity:  0.85,
		}),
		mosaic.WithSubtitles(mosaic.SubtitleTrack{
			Path:     "./subtitles/fa.srt", // تبدیل خودکار به WebVTT
			Language: "fa",
			Label:    "فارسی",
			Default:  true,
		}),
		mosaic.WithThreads(4),
	)
	if err != nil {
		log.Fatalf("خطا در پکیجینگ HLS: %v", err)
	}

	fmt.Printf("\nپکیجینگ با موفقیت انجام شد! زمان مصرف پردازنده: %.2fs | اوج رم: %d KB\n",
		usage.UserTime, usage.MaxMemory)
}
```

---

## ۴. پکیجینگ استریم DASH CMAF

```go
package main

import (
	"context"
	"fmt"
	"log"

	"github.com/farshidrezaei/mosaic"
)

func main() {
	job := mosaic.Job{
		Input:     "input.mp4",
		OutputDir: "./output/dash",
		Profile:   mosaic.ProfileVOD,
	}

	_, err := mosaic.EncodeDash(
		context.Background(),
		job,
		mosaic.WithNormalizeOrientation(),
		mosaic.WithBFrames(2),
		mosaic.WithScaleBitrateWithFPS(),
	)
	if err != nil {
		log.Fatalf("خطا در پکیجینگ DASH: %v", err)
	}

	fmt.Println("استریم DASH با موفقیت در ./output/dash/manifest.mpd تولید شد.")
}
```
