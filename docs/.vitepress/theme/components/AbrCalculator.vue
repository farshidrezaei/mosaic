<template>
  <div class="abr-calculator glass-card">
    <div class="calc-header">
      <div class="calc-title">
        <span class="calc-icon">⚡</span>
        <h3>{{ isFa ? 'ماشین‌حساب هوشمند نردبان کیفیت (ABR Calculator)' : 'Interactive ABR Ladder & Bitrate Calculator' }}</h3>
      </div>
      <span class="pill-live">{{ isFa ? 'محاسبه بلادرنگ' : 'Live Simulation' }}</span>
    </div>

    <p class="calc-desc">
      {{ isFa 
        ? 'رزولوشن، نسبت تصویر و تنظیمات ویدیوی خود را وارد کنید تا ساختار Renditionها و بهینه‌سازی‌های Mosaic را به صورت زنده مشاهده کنید.' 
        : 'Select your source video parameters to simulate how Mosaic computes aspect-preserving ladders, bitrate caps, and VBV buffer sizes.' 
      }}
    </p>

    <div class="calc-grid">
      <!-- Input Resolution -->
      <div class="input-group">
        <label>{{ isFa ? 'پیکربندی ابعاد منبع (Resolution)' : 'Source Resolution Preset' }}</label>
        <select v-model="selectedPreset" @change="applyPreset">
          <option value="1080p_landscape">1080p Landscape (1920x1080 - 16:9)</option>
          <option value="4k_landscape">4K Ultra HD (3840x2160 - 16:9)</option>
          <option value="720p_landscape">720p HD (1280x720 - 16:9)</option>
          <option value="1080p_portrait">Mobile Portrait (1080x1920 - 9:16)</option>
          <option value="square">Square Video (1080x1080 - 1:1)</option>
          <option value="custom">{{ isFa ? 'ابعاد سفارشی (Custom)' : 'Custom Dimensions' }}</option>
        </select>
      </div>

      <!-- Custom Inputs -->
      <div v-if="selectedPreset === 'custom'" class="input-row">
        <div class="input-group">
          <label>{{ isFa ? 'عرض (Width)' : 'Width (px)' }}</label>
          <input type="number" v-model.number="sourceWidth" min="100" max="7680" step="2" />
        </div>
        <div class="input-group">
          <label>{{ isFa ? 'ارتفاع (Height)' : 'Height (px)' }}</label>
          <input type="number" v-model.number="sourceHeight" min="100" max="4320" step="2" />
        </div>
      </div>

      <!-- Framerate & Codec -->
      <div class="input-row">
        <div class="input-group">
          <label>{{ isFa ? 'نرخ فریم (Framerate)' : 'Framerate (FPS)' }}</label>
          <select v-model.number="sourceFps">
            <option :value="24">24 FPS (Film)</option>
            <option :value="30">30 FPS (Standard)</option>
            <option :value="60">60 FPS (High-Framerate)</option>
          </select>
        </div>

        <div class="input-group">
          <label>{{ isFa ? 'کدک ویدیو (Codec)' : 'Target Video Codec' }}</label>
          <select v-model="selectedCodec">
            <option value="h264">H.264 / AVC (Universal)</option>
            <option value="hevc">HEVC / H.265 (High Efficiency)</option>
            <option value="av1">AV1 (Next-Gen Open)</option>
          </select>
        </div>
      </div>

      <!-- Options Toggles -->
      <div class="toggles-row">
        <label class="checkbox-label">
          <input type="checkbox" v-model="fpsScaling" />
          <span>{{ isFa ? 'مقیاس بیت‌ریت برای نرخ فریم بالا (>30 FPS)' : 'Scale Bitrate with High FPS (>30)' }}</span>
        </label>
        <label class="checkbox-label">
          <input type="checkbox" v-model="useCRF" />
          <span>{{ isFa ? 'بهینه‌سازی بر پایه محتوا (Capped-CRF)' : 'Content-Aware Capped-CRF' }}</span>
        </label>
      </div>
    </div>

    <!-- Calculated Output Table -->
    <div class="results-container">
      <div class="results-header">
        <h4>{{ isFa ? 'نردبان رندیشن‌های تولیدشده (Generated ABR Ladder)' : 'Computed ABR Renditions' }}</h4>
        <span class="badge-ratio">
          {{ isFa ? 'نسبت تصویر:' : 'Aspect Ratio:' }} {{ computedRatio }}
        </span>
      </div>

      <div class="table-responsive">
        <table class="ladder-table">
          <thead>
            <tr>
              <th>{{ isFa ? 'سطح' : 'Tier' }}</th>
              <th>{{ isFa ? 'ابعاد خروجی' : 'Resolution' }}</th>
              <th>{{ isFa ? 'بیت‌ریت سقف (MaxRate)' : 'Max Bitrate' }}</th>
              <th>{{ isFa ? 'بافر VBV (BufSize)' : 'VBV Buffer' }}</th>
              <th>{{ isFa ? 'پروفایل / لول' : 'Profile & Level' }}</th>
              <th>{{ isFa ? 'صرفه‌جویی' : 'Bandwidth Savings' }}</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(rung, idx) in computedLadder" :key="idx">
              <td>
                <span class="rung-badge" :class="'rung-' + idx">#{{ idx + 1 }}</span>
              </td>
              <td class="font-mono font-bold">{{ rung.width }}x{{ rung.height }}</td>
              <td class="font-mono text-cyan">{{ rung.maxRate }} kbps</td>
              <td class="font-mono">{{ rung.bufSize }} kbps</td>
              <td><span class="profile-tag">{{ rung.profile }} ({{ rung.level }})</span></td>
              <td class="text-green font-bold">{{ rung.savings }}</td>
            </tr>
          </tbody>
        </table>
      </div>

      <div class="summary-box">
        <div class="summary-item">
          <span class="summary-label">{{ isFa ? 'تعداد استریم‌های تولیدی:' : 'Total Renditions:' }}</span>
          <span class="summary-value">{{ computedLadder.length }} Variants</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">{{ isFa ? 'کل پهنای باند ماکسیمم:' : 'Peak Stream Bandwidth:' }}</span>
          <span class="summary-value text-cyan">{{ totalBandwidth }} kbps</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">{{ isFa ? 'راندمان فشرده‌سازی کدک:' : 'Codec Efficiency Gain:' }}</span>
          <span class="summary-value text-green">{{ codecEfficiency }}</span>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vitepress'

const route = useRoute()
const isFa = computed(() => route.path.startsWith('/fa/') || route.path.startsWith('/mosaic/fa/'))

const selectedPreset = ref('1080p_landscape')
const sourceWidth = ref(1920)
const sourceHeight = ref(1080)
const sourceFps = ref(30)
const selectedCodec = ref('h264')
const fpsScaling = ref(true)
const useCRF = ref(true)

function applyPreset() {
  switch (selectedPreset.value) {
    case '1080p_landscape':
      sourceWidth.value = 1920
      sourceHeight.value = 1080
      break
    case '4k_landscape':
      sourceWidth.value = 3840
      sourceHeight.value = 2160
      break
    case '720p_landscape':
      sourceWidth.value = 1280
      sourceHeight.value = 720
      break
    case '1080p_portrait':
      sourceWidth.value = 1080
      sourceHeight.value = 1920
      break
    case 'square':
      sourceWidth.value = 1080
      sourceHeight.value = 1080
      break
  }
}

const computedRatio = computed(() => {
  const gcd = (a: number, b: number): number => b === 0 ? a : gcd(b, a % b)
  const d = gcd(sourceWidth.value, sourceHeight.value)
  return `${sourceWidth.value / d}:${sourceHeight.value / d}`
})

const codecMultiplier = computed(() => {
  switch (selectedCodec.value) {
    case 'av1': return 0.60
    case 'hevc': return 0.70
    default: return 1.0
  }
})

const codecEfficiency = computed(() => {
  switch (selectedCodec.value) {
    case 'av1': return '40% - 50% vs H.264'
    case 'hevc': return '30% - 35% vs H.264'
    default: return 'Standard Baseline'
  }
})

const computedLadder = computed(() => {
  const w = sourceWidth.value
  const h = sourceHeight.value
  const fps = sourceFps.value

  const rungs: Array<{ height: number; defaultRate: number; profile: string; level: string }> = []

  if (h >= 1080) {
    rungs.push({ height: 1080, defaultRate: 5000, profile: 'main', level: '4.0' })
    rungs.push({ height: 720, defaultRate: 3000, profile: 'main', level: '3.1' })
    rungs.push({ height: 360, defaultRate: 1000, profile: 'baseline', level: '3.0' })
  } else if (h >= 720) {
    rungs.push({ height: 720, defaultRate: 3000, profile: 'main', level: '3.1' })
    rungs.push({ height: 360, defaultRate: 1000, profile: 'baseline', level: '3.0' })
  } else if (h >= 360) {
    rungs.push({ height: 360, defaultRate: 1000, profile: 'baseline', level: '3.0' })
  } else {
    rungs.push({ height: h, defaultRate: 1000, profile: 'baseline', level: '3.0' })
  }

  let fpsFactor = 1.0
  if (fpsScaling.value && fps > 30) {
    fpsFactor = Math.min(1.5, fps / 30.0)
  }

  return rungs.map(r => {
    let targetW = Math.round((r.height * w) / h)
    if (targetW % 2 !== 0) targetW++
    let targetH = r.height
    if (targetH % 2 !== 0) targetH++

    let rate = r.defaultRate
    if (r.height >= 1080) rate = Math.min(rate, 5000)
    else if (r.height >= 720) rate = Math.min(rate, 3000)
    else rate = Math.min(rate, 1000)

    rate = Math.round(rate * fpsFactor * codecMultiplier.value)
    if (useCRF.value) {
      rate = Math.round(rate * 0.88)
    }

    const bufSize = rate * 2
    const savings = selectedCodec.value === 'h264' && !useCRF.value 
      ? '0%' 
      : `${Math.round((1 - (rate / (r.defaultRate * fpsFactor))) * 100)}%`

    return {
      width: targetW,
      height: targetH,
      maxRate: rate,
      bufSize,
      profile: selectedCodec.value === 'h264' ? r.profile : selectedCodec.value.toUpperCase(),
      level: r.level,
      savings
    }
  })
})

const totalBandwidth = computed(() => {
  return computedLadder.value.reduce((acc, r) => acc + r.maxRate, 0)
})
</script>

<style scoped>
.glass-card {
  background: rgba(13, 20, 36, 0.75);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 20px;
  padding: 1.75rem;
  margin: 2rem 0;
  box-shadow: 0 20px 50px -10px rgba(0, 0, 0, 0.5), inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.calc-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 0.75rem;
  flex-wrap: wrap;
  gap: 10px;
}

.calc-title {
  display: flex;
  align-items: center;
  gap: 10px;
}

.calc-title h3 {
  margin: 0 !important;
  font-size: 1.35rem !important;
  font-weight: 800 !important;
  color: #ffffff !important;
}

.calc-icon {
  font-size: 1.5rem;
}

.pill-live {
  background: rgba(2, 132, 199, 0.2);
  border: 1px solid rgba(56, 189, 248, 0.4);
  color: #38bdf8;
  font-size: 0.78rem;
  font-weight: 700;
  padding: 4px 10px;
  border-radius: 9999px;
}

.calc-desc {
  color: #94a3b8;
  font-size: 0.92rem;
  margin-bottom: 1.5rem;
  line-height: 1.6;
}

.calc-grid {
  display: flex;
  flex-direction: column;
  gap: 16px;
  margin-bottom: 2rem;
  background: rgba(10, 15, 29, 0.6);
  padding: 1.25rem;
  border-radius: 16px;
  border: 1px solid rgba(255, 255, 255, 0.06);
}

.input-row {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 14px;
}

.input-group {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.input-group label {
  font-size: 0.85rem;
  font-weight: 600;
  color: #cbd5e1;
}

.input-group select,
.input-group input {
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 10px;
  color: #f8fafc;
  padding: 8px 12px;
  font-size: 0.9rem;
  font-family: inherit;
  outline: none;
  transition: border-color 0.2s;
}

.input-group select:focus,
.input-group input:focus {
  border-color: #38bdf8;
}

.toggles-row {
  display: flex;
  gap: 20px;
  flex-wrap: wrap;
  margin-top: 6px;
}

.checkbox-label {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 0.88rem;
  color: #cbd5e1;
  cursor: pointer;
}

.results-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
  flex-wrap: wrap;
  gap: 10px;
}

.results-header h4 {
  margin: 0 !important;
  font-size: 1.15rem;
  font-weight: 700;
  color: #f8fafc;
}

.badge-ratio {
  background: rgba(99, 102, 241, 0.2);
  border: 1px solid rgba(129, 140, 248, 0.4);
  color: #818cf8;
  font-size: 0.82rem;
  padding: 4px 10px;
  border-radius: 8px;
  font-family: var(--vp-font-family-mono);
}

.table-responsive {
  overflow-x: auto;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.ladder-table {
  width: 100%;
  border-collapse: collapse;
  margin: 0 !important;
  font-size: 0.9rem;
}

.ladder-table th {
  background: rgba(30, 41, 59, 0.8) !important;
  color: #f8fafc !important;
  padding: 10px 14px !important;
  font-weight: 700;
  text-align: left;
}

[dir="rtl"] .ladder-table th {
  text-align: right;
}

.ladder-table td {
  padding: 10px 14px !important;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.rung-badge {
  background: rgba(2, 132, 199, 0.2);
  color: #38bdf8;
  font-size: 0.75rem;
  font-weight: 800;
  padding: 3px 8px;
  border-radius: 6px;
}

.profile-tag {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  padding: 2px 8px;
  border-radius: 6px;
  font-size: 0.8em;
  color: #94a3b8;
}

.text-cyan { color: #38bdf8; }
.text-green { color: #34d399; }
.font-mono { font-family: var(--vp-font-family-mono); }
.font-bold { font-weight: 700; }

.summary-box {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
  gap: 12px;
  margin-top: 1.25rem;
  background: rgba(15, 23, 42, 0.6);
  padding: 1rem;
  border-radius: 12px;
  border: 1px solid rgba(255, 255, 255, 0.08);
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.summary-label {
  font-size: 0.8rem;
  color: #94a3b8;
}

.summary-value {
  font-size: 1.1rem;
  font-weight: 800;
  font-family: var(--vp-font-family-mono);
  color: #f8fafc;
}
</style>
