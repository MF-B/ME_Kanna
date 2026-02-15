import { ref } from 'vue'
import en from '../locales/en'
import zh from '../locales/zh'

const currentLocale = ref('zh') // Default to Chinese
const messages = { en, zh }

export function useI18n() {
    const t = (path) => {
        const keys = path.split('.')
        let value = messages[currentLocale.value]
        for (const key of keys) {
            if (value === undefined || value === null) break
            value = value[key]
        }
        return value !== undefined ? value : path
    }

    const toggleLocale = () => {
        currentLocale.value = currentLocale.value === 'en' ? 'zh' : 'en'
    }

    return {
        locale: currentLocale,
        t,
        toggleLocale
    }
}
