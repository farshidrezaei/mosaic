import DefaultTheme from 'vitepress/theme'
import './style.css'
import AbrCalculator from './components/AbrCalculator.vue'
import StreamPlayer from './components/StreamPlayer.vue'

export default {
  extends: DefaultTheme,
  enhanceApp({ app }) {
    app.component('AbrCalculator', AbrCalculator)
    app.component('StreamPlayer', StreamPlayer)
  }
}
