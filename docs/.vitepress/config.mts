import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'Mosaic',
  description: 'Predictable, Production-Ready Adaptive Bitrate (ABR) Video Packaging for Go (HLS & DASH CMAF)',
  base: '/mosaic/',
  cleanUrls: true,
  lastUpdated: true,

  sitemap: {
    hostname: 'https://farshidrezaei.github.io/mosaic/'
  },

  head: [
    ['link', { rel: 'icon', type: 'image/jpeg', href: '/mosaic/logo.jpg' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.googleapis.com' }],
    ['link', { rel: 'preconnect', href: 'https://fonts.gstatic.com', crossorigin: '' }],
    ['link', { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@300;400;500;600;700;800&family=Vazirmatn:wght@300;400;500;600;700;800&family=JetBrains+Mono:wght@400;500;600&display=swap' }],
    ['meta', { name: 'theme-color', content: '#0284c7' }],
    ['meta', { property: 'og:type', content: 'website' }],
    ['meta', { property: 'og:locale', content: 'en_US' }],
    ['meta', { property: 'og:site_name', content: 'Mosaic Documentation' }],
    ['meta', { property: 'og:image', content: 'https://farshidrezaei.github.io/mosaic/logo.jpg' }],
    ['meta', { name: 'twitter:card', content: 'summary_large_image' }],
    ['meta', { name: 'twitter:image', content: 'https://farshidrezaei.github.io/mosaic/logo.jpg' }]
  ],

  locales: {
    root: {
      label: 'English',
      lang: 'en',
      title: 'Mosaic',
      description: 'Predictable, Production-Ready Adaptive Bitrate (ABR) Video Packaging for Go',
      themeConfig: {
        nav: [
          { text: 'Guide', link: '/quickstart' },
          { text: 'API Reference', link: '/API' },
          { text: 'Options', link: '/options' },
          { text: 'Architecture', link: '/ARCHITECTURE' },
          { text: 'Examples', link: '/EXAMPLES' },
          {
            text: 'v1.8.0',
            items: [
              { text: 'Changelog', link: '/CHANGELOG' },
              { text: 'GitHub Releases', link: 'https://github.com/farshidrezaei/mosaic/releases' }
            ]
          }
        ],

        sidebar: [
          {
            text: 'Getting Started',
            collapsed: false,
            items: [
              { text: 'Overview', link: '/README' },
              { text: 'Quick Start', link: '/quickstart' },
              { text: 'Installation', link: '/installation' }
            ]
          },
          {
            text: 'Architecture & Engine',
            collapsed: false,
            items: [
              { text: 'System Architecture', link: '/ARCHITECTURE' },
              { text: 'Encoding & CMAF Pipeline', link: '/ENCODING' },
              { text: 'Orientation Normalization', link: '/orientation' }
            ]
          },
          {
            text: 'API & Configuration',
            collapsed: false,
            items: [
              { text: 'Public API', link: '/API' },
              { text: 'Functional Options', link: '/options' }
            ]
          },
          {
            text: 'Examples & Community',
            collapsed: false,
            items: [
              { text: 'Examples Catalog', link: '/EXAMPLES' },
              { text: 'Development Roadmap', link: '/ROADMAP_PROGRESS' },
              { text: 'Testing Strategy', link: '/TESTING' },
              { text: 'Troubleshooting & FAQ', link: '/TROUBLESHOOTING' },
              { text: 'Changelog', link: '/CHANGELOG' }
            ]
          }
        ],

        docFooter: {
          prev: 'Previous Page',
          next: 'Next Page'
        },

        editLink: {
          pattern: 'https://github.com/farshidrezaei/mosaic/edit/main/docs/:path',
          text: 'Edit this page on GitHub'
        },

        footer: {
          message: 'Released under the MIT License.',
          copyright: 'Copyright © 2026 Farshid Rezaei. Built with Vue 3 & VitePress.'
        }
      }
    },

    fa: {
      label: 'فارسی',
      lang: 'fa',
      dir: 'rtl',
      title: 'موزاییک (Mosaic)',
      description: 'کتابخانه قدرتمند و آماده پروداکشن برای پکیجینگ ویدیویی بیت‌ریت تطبیقی (ABR) به HLS و DASH در Go',
      themeConfig: {
        nav: [
          { text: 'شروع سریع', link: '/fa/quickstart' },
          { text: 'مرجع API', link: '/fa/API' },
          { text: 'تنظیمات و فلگ‌ها', link: '/fa/options' },
          { text: 'معماری سیستم', link: '/fa/ARCHITECTURE' },
          { text: 'نمونه‌کدها', link: '/fa/EXAMPLES' },
          {
            text: 'نسخه ۱.۸.۰',
            items: [
              { text: 'تغییرات نسخه‌ها', link: '/CHANGELOG' },
              { text: 'ریلیزهای گیت‌هاب', link: 'https://github.com/farshidrezaei/mosaic/releases' }
            ]
          }
        ],

        sidebar: [
          {
            text: 'شروع کار',
            collapsed: false,
            items: [
              { text: 'معرفی کلی', link: '/fa/' },
              { text: 'شروع سریع', link: '/fa/quickstart' },
              { text: 'نصب و پیش‌نیازها', link: '/fa/installation' }
            ]
          },
          {
            text: 'معماری و خط لوله انکودینگ',
            collapsed: false,
            items: [
              { text: 'معماری سیستم', link: '/fa/ARCHITECTURE' },
              { text: 'خط‌لوله انکودینگ و CMAF', link: '/fa/ENCODING' },
              { text: 'نرمال‌سازی چرخش ویدیو', link: '/fa/orientation' }
            ]
          },
          {
            text: 'توابع و گزینه‌ها',
            collapsed: false,
            items: [
              { text: 'مرجع API و توابع', link: '/fa/API' },
              { text: 'گزینه‌های پیکربندی (Options)', link: '/fa/options' }
            ]
          },
          {
            text: 'نمونه‌ها و راهنماها',
            collapsed: false,
            items: [
              { text: 'کاتالوگ نمونه‌کدها', link: '/fa/EXAMPLES' },
              { text: 'نقشه راه و پیشرفت', link: '/ROADMAP_PROGRESS' },
              { text: 'استراتژی تست', link: '/fa/TESTING' },
              { text: 'عیب‌یابی و پرسش‌های متداول', link: '/fa/TROUBLESHOOTING' },
              { text: 'تغییرات نسخه‌ها', link: '/CHANGELOG' }
            ]
          }
        ],

        docFooter: {
          prev: 'صفحه قبلی',
          next: 'صفحه بعدی'
        },

        outline: {
          label: 'در این صفحه'
        },

        editLink: {
          pattern: 'https://github.com/farshidrezaei/mosaic/edit/main/docs/:path',
          text: 'ویرایش این صفحه در گیت‌هاب'
        },

        footer: {
          message: 'منتشر شده تحت لایسنس MIT.',
          copyright: 'تمامی حقوق محفوظ است © ۲۰۲۶ فرشید رضایی. توسعه داده شده با Vue 3 و VitePress.'
        },

        darkModeSwitchLabel: 'تم تاریک/روشن',
        sidebarMenuLabel: 'منو',
        returnToTopLabel: 'بازگشت به بالا'
      }
    }
  },

  themeConfig: {
    logo: '/logo.jpg',

    socialLinks: [
      { icon: 'github', link: 'https://github.com/farshidrezaei/mosaic' }
    ],

    search: {
      provider: 'local',
      options: {
        locales: {
          root: {
            translations: {
              button: {
                buttonText: 'Search docs',
                buttonAriaLabel: 'Search docs'
              },
              modal: {
                noResultsText: 'No results found',
                resetButtonTitle: 'Reset search',
                footer: {
                  selectText: 'to select',
                  navigateText: 'to navigate',
                  closeText: 'to close'
                }
              }
            }
          },
          fa: {
            translations: {
              button: {
                buttonText: 'جستجو در مستندات...',
                buttonAriaLabel: 'جستجو در مستندات'
              },
              modal: {
                noResultsText: 'نتیجه‌ای یافت نشد',
                resetButtonTitle: 'پاک کردن جستجو',
                footer: {
                  selectText: 'انتخاب',
                  navigateText: 'پیمایش',
                  closeText: 'بستن'
                }
              }
            }
          }
        }
      }
    }
  }
})
