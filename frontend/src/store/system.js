import { defineStore } from 'pinia'
import { ref } from 'vue'
import { getPublicSystemConfig } from '@/api/system'

export const useSystemStore = defineStore('system', () => {
    // State
    const siteTitle = ref(localStorage.getItem('siteTitle') || 'Gost Panel')
    const logoUrl = ref(localStorage.getItem('logoUrl') || 'https://gost.run/images/gost.png')
    const copyright = ref(localStorage.getItem('copyright') || '')

    // Actions
    async function fetchSystemConfig() {
        try {
            const res = await getPublicSystemConfig()
            if (res.data) {
                const config = res.data

                // Update state
                if (config.siteTitle) {
                    siteTitle.value = config.siteTitle
                    document.title = config.siteTitle
                    localStorage.setItem('siteTitle', config.siteTitle)
                }

                if (config.logoUrl) {
                    logoUrl.value = config.logoUrl
                    localStorage.setItem('logoUrl', config.logoUrl)
                } else {
                    logoUrl.value = ''
                    localStorage.removeItem('logoUrl')
                }

                if (config.copyright) {
                    copyright.value = config.copyright
                    localStorage.setItem('copyright', config.copyright)
                }
            }
        } catch (error) {
            console.error('获取系统配置失败:', error)
        }
    }

    // Initialize document title from storage immediately
    if (localStorage.getItem('siteTitle')) {
        document.title = localStorage.getItem('siteTitle')
    }

    return {
        siteTitle,
        logoUrl,
        copyright,
        fetchSystemConfig
    }
})
