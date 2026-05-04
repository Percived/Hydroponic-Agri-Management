import { defineStore } from 'pinia'
import { ref } from 'vue'
import { greenhouseApi } from '@/api'
import type { Greenhouse } from '@/types'

export const useGreenhouseStore = defineStore('greenhouse', () => {
  const greenhouses = ref<Greenhouse[]>([])
  const loading = ref(false)

  // 获取温室列表
  async function fetchGreenhouses() {
    loading.value = true
    try {
      const data = await greenhouseApi.getGreenhouses({ page: 1, page_size: 100 })
      greenhouses.value = data.items
    } catch {
      // 错误已处理
    } finally {
      loading.value = false
    }
  }

  return {
    greenhouses,
    loading,
    fetchGreenhouses
  }
})
