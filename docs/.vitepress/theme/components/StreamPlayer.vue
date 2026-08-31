<template>
  <div class="stream-player-card glass-card">
    <div class="player-header">
      <div class="player-title">
        <span class="player-icon">📺</span>
        <div>
          <h3>{{ isFa ? 'پلیر و آنالایزر استریم‌های ویدیویی (Stream Inspector)' : 'Interactive Stream Player & Inspector' }}</h3>
          <p class="player-subtitle">
            {{ isFa 
              ? 'پوشه خروجی استریم سیستم خود را انتخاب کنید یا آدرس فایل master.m3u8 را وارد نمایید.' 
              : 'Inspect your local package output folder directly or enter a remote/localhost master.m3u8 URL.' 
            }}
          </p>
        </div>
      </div>
      <span class="pill-engine">Powered by Hls.js & Dash</span>
    </div>

    <!-- Mode Selector Tabs -->
    <div class="source-tabs">
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'file' }" 
        @click="switchTab('file')"
      >
        <span class="tab-icon">📄</span>
        {{ isFa ? 'انتخاب فایل master.m3u8' : 'Select master.m3u8 File' }}
      </button>
      <button 
        class="tab-btn" 
        :class="{ active: activeTab === 'url' }" 
        @click="switchTab('url')"
      >
        <span class="tab-icon">🔗</span>
        {{ isFa ? 'آدرس URL یا استریم‌های دمو' : 'Stream URL & Demos' }}
      </button>
    </div>

    <!-- Tab 1: Local File Input -->
    <div v-if="activeTab === 'file'" class="tab-content">
      <div 
        class="drop-zone"
        :class="{ dragging: isDragging, 'has-files': localFilesCount > 0 }"
        @dragover.prevent="isDragging = true"
        @dragleave.prevent="isDragging = false"
        @drop.prevent="handleFileDrop"
        @click="triggerFileSelect"
      >
        <input 
          type="file" 
          ref="fileInput" 
          accept=".m3u8,.mpd,.ts,.m4s,.mp4,.vtt,.key" 
          multiple 
          class="hidden-input" 
          @change="handleFileChange" 
        />
        <div class="drop-icon">🎬</div>
        <div class="drop-text">
          <h4 v-if="localFilesCount === 0">
            {{ isFa ? 'فایل master.m3u8 (یا فایل‌های استریم) را انتخاب کنید یا اینجا بکشید' : 'Select master.m3u8 (or stream files) or drag & drop here' }}
          </h4>
          <h4 v-else class="text-cyan">
            {{ isFa ? `✅ فایل ${selectedMasterName} لود شد (${localFilesCount} فایل در حافظه)` : `✅ Loaded ${selectedMasterName} (${localFilesCount} files in memory)` }}
          </h4>
          <p>
            {{ isFa 
              ? 'پخش آنی و مستقیم بدون آپلود در سرور؛ برای استریم‌های لوکال می‌توانید تمام فایل‌های پوشه خروجی را همزمان انتخاب فرمایید.' 
              : 'Direct in-memory playback. For local streams, select the master.m3u8 along with variant files.' 
            }}
          </p>
        </div>
      </div>
    </div>

    <!-- Tab 2: URL / Demos Input -->
    <div v-if="activeTab === 'url'" class="tab-content">
      <div class="url-input-bar">
        <input 
          type="text" 
          v-model="streamUrl" 
          :placeholder="isFa ? 'آدرس استریم (مثال: http://localhost:8080/master.m3u8)...' : 'Enter stream URL (e.g. http://localhost:8080/master.m3u8)...'" 
          @keyup.enter="loadUrlStream"
        />
        <button class="btn-play-load" @click="loadUrlStream">
          {{ isFa ? 'پخش استریم ▶' : 'Load & Play ▶' }}
        </button>
      </div>

      <div class="demo-chips">
        <span class="chips-label">{{ isFa ? 'نمونه‌های آماده تست:' : 'Quick Demo Streams:' }}</span>
        <button class="chip-btn" @click="loadDemo('https://test-streams.mux.dev/x36xhzz/x36xhzz.m3u8')">
          ⚡ Multi-Bitrate HLS
        </button>
        <button class="chip-btn" @click="loadDemo('https://dash.akamaized.net/akamai/bbb_30fps/bbb_30fps.mpd')">
          🎬 Big Buck Bunny DASH
        </button>
        <button class="chip-btn" @click="loadDemo('http://localhost:8080/master.m3u8')">
          💻 Local Mosaic Preview (:8080)
        </button>
      </div>
    </div>

    <!-- Video Player Screen -->
    <div class="video-section">
      <div class="video-wrapper">
        <video ref="videoElement" controls playsinline></video>
        <div v-if="loadingStream" class="loader-overlay">
          <div class="spinner"></div>
          <span>{{ isFa ? 'در حال بارگذاری و تحلیل مانیفست استریم...' : 'Parsing stream manifest & buffer...' }}</span>
        </div>
      </div>

      <!-- Live Stream Diagnostics / Inspector -->
      <div class="diagnostics-panel">
        <div class="diag-header">
          <span class="diag-title">{{ isFa ? '📊 تله‌متری و مشخصات زنده استریم:' : '📊 Live Stream Telemetry & Rendition Controls:' }}</span>
          <span class="status-pill" :class="streamStatus.type">{{ streamStatus.text }}</span>
        </div>

        <div class="diag-grid">
          <!-- Quality Selector -->
          <div class="diag-item">
            <label>{{ isFa ? 'کیفیت فعال (ABR Quality):' : 'ABR Rendition:' }}</label>
            <select v-model="selectedLevelIndex" @change="changeQualityLevel">
              <option :value="-1">Auto (Dynamic Adaptive)</option>
              <option v-for="(lvl, i) in qualityLevels" :key="i" :value="i">
                {{ lvl.height ? `${lvl.height}p (${Math.round(lvl.bitrate / 1000)} kbps)` : `Level ${i}` }}
              </option>
            </select>
          </div>

          <!-- Current Resolution -->
          <div class="diag-item">
            <label>{{ isFa ? 'رزولوشن فعلی ویدیو:' : 'Current Video Dimensions:' }}</label>
            <span class="diag-val font-mono">{{ currentResolution || 'Auto / Probing...' }}</span>
          </div>

          <!-- Current Bitrate -->
          <div class="diag-item">
            <label>{{ isFa ? 'بیت‌ریت لحظه‌ای:' : 'Current Bitrate:' }}</label>
            <span class="diag-val font-mono text-cyan">{{ currentBitrate || 'Dynamic' }}</span>
          </div>

          <!-- Buffer Length -->
          <div class="diag-item">
            <label>{{ isFa ? 'بافر دانلود شده:' : 'Forward Buffer:' }}</label>
            <span class="diag-val font-mono">{{ bufferLength.toFixed(1) }}s</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useRoute } from 'vitepress'

const route = useRoute()
const isFa = computed(() => route.path.startsWith('/fa/') || route.path.startsWith('/mosaic/fa/'))

const activeTab = ref<'file' | 'url'>('file')
const isDragging = ref(false)
const fileInput = ref<HTMLInputElement | null>(null)
const videoElement = ref<HTMLVideoElement | null>(null)
const loadingStream = ref(false)

const streamUrl = ref('')
const localFilesCount = ref(0)
const selectedMasterName = ref('')
const fileMap = new Map<string, File>()

const qualityLevels = ref<Array<{ height: number; bitrate: number }>>([])
const selectedLevelIndex = ref(-1)
const currentResolution = ref('')
const currentBitrate = ref('')
const bufferLength = ref(0)
const streamStatus = ref({ type: 'idle', text: 'Ready' })

let hlsInstance: any = null
let bufferInterval: any = null

onMounted(async () => {
  bufferInterval = setInterval(updateBufferStats, 1000)
})

onUnmounted(() => {
  if (hlsInstance) {
    hlsInstance.destroy()
    hlsInstance = null
  }
  if (bufferInterval) {
    clearInterval(bufferInterval)
  }
})

function switchTab(tab: 'file' | 'url') {
  activeTab.value = tab
}

function triggerFileSelect() {
  if (fileInput.value) {
    fileInput.value.click()
  }
}

function handleFileChange(e: Event) {
  const target = e.target as HTMLInputElement
  if (target.files && target.files.length > 0) {
    processLocalFiles(Array.from(target.files))
  }
}

function handleFileDrop(e: DragEvent) {
  isDragging.value = false
  if (e.dataTransfer && e.dataTransfer.files) {
    processLocalFiles(Array.from(e.dataTransfer.files))
  }
}

function processLocalFiles(files: File[]) {
  fileMap.clear()
  localFilesCount.value = files.length

  let masterFile: File | null = null

  for (const f of files) {
    const name = f.name.toLowerCase()
    fileMap.set(f.name, f)
    fileMap.set(name, f)

    if (name === 'master.m3u8') {
      masterFile = f
    } else if (!masterFile && name.endsWith('.m3u8')) {
      masterFile = f
    }
  }

  if (masterFile) {
    selectedMasterName.value = masterFile.name
    playLocalMaster(masterFile)
  } else {
    streamStatus.value = {
      type: 'error',
      text: isFa.value ? 'فایل master.m3u8 در بین فایل‌ها یافت نشد!' : 'No master.m3u8 found among selected files!'
    }
  }
}

async function playLocalMaster(masterFile: File) {
  loadingStream.value = true
  streamStatus.value = { type: 'active', text: isFa.value ? 'پخش محلی ABR' : 'Local In-Memory Stream' }

  try {
    const HlsModule = await import('hls.js')
    const Hls = HlsModule.default

    if (hlsInstance) {
      hlsInstance.destroy()
    }

    // Custom Hls.js Loader reading local File objects in memory
    class CustomLocalLoader extends Hls.DefaultConfig.loader {
      load(context: any, config: any, callbacks: any) {
        const urlStr = context.url
        let filename = urlStr.substring(urlStr.lastIndexOf('/') + 1)
        if (filename.includes('?')) {
          filename = filename.substring(0, filename.indexOf('?'))
        }

        const localFile = fileMap.get(filename) || fileMap.get(filename.toLowerCase())

        if (localFile) {
          const reader = new FileReader()
          if (context.responseType === 'arraybuffer') {
            reader.readAsArrayBuffer(localFile)
          } else {
            reader.readAsText(localFile)
          }

          reader.onload = () => {
            callbacks.onSuccess(
              { data: reader.result, url: context.url },
              { stats: { trequest: 0, tfirst: 0, tload: 0, loaded: localFile.size, total: localFile.size } },
              context
            )
          }
          reader.onerror = () => {
            callbacks.onError({ code: 404, text: 'File read error' }, context)
          }
        } else {
          // Fallback to default loader
          super.load(context, config, callbacks)
        }
      }
    }

    if (Hls.isSupported() && videoElement.value) {
      const hls = new Hls({
        pLoader: CustomLocalLoader as any,
        fLoader: CustomLocalLoader as any,
        enableWorker: true,
      })

      hlsInstance = hls
      const masterUrl = URL.createObjectURL(masterFile)
      hls.loadSource(masterUrl)
      hls.attachMedia(videoElement.value)

      setupHlsEvents(hls)
      videoElement.value.play().catch(() => {})
    } else if (videoElement.value && videoElement.value.canPlayType('application/vnd.apple.mpegurl')) {
      // Native Safari
      videoElement.value.src = URL.createObjectURL(masterFile)
      videoElement.value.play().catch(() => {})
    }
  } catch (err) {
    console.error('Failed to load Hls.js:', err)
  } finally {
    loadingStream.value = false
  }
}

async function loadUrlStream() {
  if (!streamUrl.value) return
  loadingStream.value = true
  streamStatus.value = { type: 'active', text: isFa.value ? 'استریم آنلاین' : 'Online Stream' }

  try {
    const HlsModule = await import('hls.js')
    const Hls = HlsModule.default

    if (hlsInstance) {
      hlsInstance.destroy()
    }

    if (Hls.isSupported() && videoElement.value) {
      const hls = new Hls({ enableWorker: true })
      hlsInstance = hls
      hls.loadSource(streamUrl.value)
      hls.attachMedia(videoElement.value)

      setupHlsEvents(hls)
      videoElement.value.play().catch(() => {})
    } else if (videoElement.value) {
      videoElement.value.src = streamUrl.value
      videoElement.value.play().catch(() => {})
    }
  } catch (err) {
    console.error('Failed to play stream URL:', err)
  } finally {
    loadingStream.value = false
  }
}

function loadDemo(url: string) {
  streamUrl.value = url
  loadUrlStream()
}

function setupHlsEvents(hls: any) {
  hls.on(hls.constructor.Events.MANIFEST_PARSED, (_: any, data: any) => {
    qualityLevels.value = data.levels.map((lvl: any) => ({
      height: lvl.height,
      bitrate: lvl.bitrate
    }))
    loadingStream.value = false
  })

  hls.on(hls.constructor.Events.LEVEL_SWITCHED, (_: any, data: any) => {
    const level = hls.levels[data.level]
    if (level) {
      currentResolution.value = `${level.width}x${level.height}`
      currentBitrate.value = `${Math.round(level.bitrate / 1000)} kbps`
    }
  })

  hls.on(hls.constructor.Events.ERROR, (_: any, data: any) => {
    if (data.fatal) {
      streamStatus.value = {
        type: 'error',
        text: isFa.value ? 'خطا در بارگذاری استریم' : `Playback Error: ${data.details}`
      }
    }
  })
}

function changeQualityLevel() {
  if (hlsInstance) {
    hlsInstance.currentLevel = selectedLevelIndex.value
  }
}

function updateBufferStats() {
  if (videoElement.value && videoElement.value.buffered.length > 0) {
    const ct = videoElement.value.currentTime
    const bufEnd = videoElement.value.buffered.end(videoElement.value.buffered.length - 1)
    bufferLength.value = Math.max(0, bufEnd - ct)

    if (videoElement.value.videoWidth > 0 && !currentResolution.value) {
      currentResolution.value = `${videoElement.value.videoWidth}x${videoElement.value.videoHeight}`
    }
  }
}
</script>

<style scoped>
.stream-player-card {
  background: rgba(13, 20, 36, 0.75);
  backdrop-filter: blur(20px) saturate(180%);
  -webkit-backdrop-filter: blur(20px) saturate(180%);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 20px;
  padding: 1.75rem;
  margin: 2rem 0;
  box-shadow: 0 20px 50px -10px rgba(0, 0, 0, 0.5), inset 0 1px 0 rgba(255, 255, 255, 0.1);
}

.player-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1.25rem;
  flex-wrap: wrap;
  gap: 12px;
}

.player-title {
  display: flex;
  align-items: flex-start;
  gap: 12px;
}

.player-title h3 {
  margin: 0 !important;
  font-size: 1.35rem !important;
  font-weight: 800 !important;
  color: #ffffff !important;
}

.player-subtitle {
  margin: 4px 0 0 0 !important;
  font-size: 0.88rem !important;
  color: #94a3b8 !important;
}

.player-icon {
  font-size: 1.8rem;
  margin-top: 2px;
}

.pill-engine {
  background: rgba(2, 132, 199, 0.2);
  border: 1px solid rgba(56, 189, 248, 0.4);
  color: #38bdf8;
  font-size: 0.78rem;
  font-weight: 700;
  padding: 4px 12px;
  border-radius: 9999px;
}

.source-tabs {
  display: flex;
  gap: 10px;
  margin-bottom: 1.25rem;
}

.tab-btn {
  display: flex;
  align-items: center;
  gap: 8px;
  background: rgba(15, 23, 42, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #94a3b8;
  padding: 8px 18px;
  border-radius: 12px;
  font-size: 0.9rem;
  font-weight: 700;
  cursor: pointer;
  transition: all 0.2s ease;
}

.tab-btn:hover {
  background: rgba(30, 41, 59, 0.8);
  color: #f8fafc;
}

.tab-btn.active {
  background: #0284c7;
  color: #ffffff;
  border-color: #38bdf8;
  box-shadow: 0 0 20px rgba(56, 189, 248, 0.35);
}

.tab-content {
  margin-bottom: 1.5rem;
}

.drop-zone {
  border: 2px dashed rgba(255, 255, 255, 0.18);
  background: rgba(10, 15, 29, 0.5);
  border-radius: 16px;
  padding: 2rem;
  text-align: center;
  cursor: pointer;
  transition: all 0.25s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 8px;
}

.drop-zone:hover,
.drop-zone.dragging {
  border-color: #38bdf8;
  background: rgba(2, 132, 199, 0.1);
  transform: scale(1.01);
}

.drop-zone.has-files {
  border-color: #34d399;
  background: rgba(16, 185, 129, 0.08);
}

.hidden-input {
  display: none;
}

.drop-icon {
  font-size: 2.4rem;
}

.drop-text h4 {
  margin: 0 0 6px 0 !important;
  font-size: 1.1rem;
  font-weight: 700;
  color: #f8fafc;
}

.drop-text p {
  margin: 0 !important;
  font-size: 0.85rem;
  color: #94a3b8;
  max-width: 600px;
}

.url-input-bar {
  display: flex;
  gap: 10px;
  margin-bottom: 12px;
}

.url-input-bar input {
  flex: 1;
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  color: #f8fafc;
  padding: 10px 16px;
  font-size: 0.92rem;
  outline: none;
  transition: border-color 0.2s;
}

.url-input-bar input:focus {
  border-color: #38bdf8;
}

.btn-play-load {
  background: linear-gradient(135deg, #0284c7 0%, #6366f1 100%);
  color: #ffffff;
  border: none;
  padding: 0 20px;
  border-radius: 12px;
  font-weight: 700;
  font-size: 0.9rem;
  cursor: pointer;
  transition: transform 0.2s;
}

.btn-play-load:hover {
  transform: translateY(-2px);
}

.demo-chips {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.chips-label {
  font-size: 0.82rem;
  color: #94a3b8;
  font-weight: 600;
}

.chip-btn {
  background: rgba(15, 23, 42, 0.8);
  border: 1px solid rgba(255, 255, 255, 0.1);
  color: #cbd5e1;
  font-size: 0.8rem;
  font-weight: 600;
  padding: 4px 12px;
  border-radius: 9999px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.chip-btn:hover {
  background: rgba(30, 41, 59, 0.9);
  border-color: #38bdf8;
  color: #38bdf8;
}

.video-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.video-wrapper {
  position: relative;
  background: #000;
  border-radius: 16px;
  overflow: hidden;
  border: 1px solid rgba(255, 255, 255, 0.1);
  box-shadow: 0 20px 40px -10px rgba(0, 0, 0, 0.6);
  max-height: 500px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.video-wrapper video {
  width: 100%;
  max-height: 500px;
  display: block;
}

.loader-overlay {
  position: absolute;
  inset: 0;
  background: rgba(7, 11, 20, 0.85);
  backdrop-filter: blur(8px);
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 12px;
  color: #38bdf8;
  font-weight: 600;
  font-size: 0.92rem;
}

.spinner {
  width: 40px;
  height: 40px;
  border: 3px solid rgba(56, 189, 248, 0.2);
  border-top-color: #38bdf8;
  border-radius: 50%;
  animation: spin 0.8s linear infinite;
}

@keyframes spin {
  to { transform: rotate(360deg); }
}

.diagnostics-panel {
  background: rgba(10, 15, 29, 0.7);
  border: 1px solid rgba(255, 255, 255, 0.08);
  border-radius: 16px;
  padding: 1.25rem;
}

.diag-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 1rem;
}

.diag-title {
  font-size: 0.92rem;
  font-weight: 700;
  color: #f8fafc;
}

.status-pill {
  font-size: 0.75rem;
  font-weight: 700;
  padding: 3px 10px;
  border-radius: 9999px;
  background: rgba(100, 116, 139, 0.2);
  color: #94a3b8;
}

.status-pill.active {
  background: rgba(16, 185, 129, 0.2);
  color: #34d399;
  border: 1px solid rgba(52, 211, 153, 0.4);
}

.status-pill.error {
  background: rgba(239, 68, 68, 0.2);
  color: #f87171;
  border: 1px solid rgba(248, 113, 113, 0.4);
}

.diag-grid {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 12px;
}

.diag-item {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.diag-item label {
  font-size: 0.78rem;
  font-weight: 600;
  color: #94a3b8;
}

.diag-item select {
  background: rgba(15, 23, 42, 0.9);
  border: 1px solid rgba(255, 255, 255, 0.1);
  border-radius: 8px;
  color: #f8fafc;
  padding: 6px 10px;
  font-size: 0.85rem;
  outline: none;
}

.diag-val {
  font-size: 0.95rem;
  font-weight: 700;
  color: #f8fafc;
}

.text-cyan { color: #38bdf8; }
.font-mono { font-family: var(--vp-font-family-mono); }
</style>
