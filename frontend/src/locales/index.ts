// @ts-check

import { createI18n } from 'vue-i18n'
import zhCN from './zh-CN.json'
import zhTW from './zh-TW.json'
import enUS from './en-US.json'

const messages = {
  'zh-CN': zhCN,
  'zh-TW': zhTW,
  'en-US': enUS,
}

type MessageSchema = typeof enUS

const i18n = createI18n<[MessageSchema], 'zh-CN' | 'zh-TW' | 'en-US'>({
  legacy: false,
  locale: 'zh-CN',
  fallbackLocale: 'zh-CN',
  messages,
})

export default i18n