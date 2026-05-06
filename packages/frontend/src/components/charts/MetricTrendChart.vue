<template>
  <div ref="chartRef" class="metric-trend-chart" />
</template>

<script setup lang="ts">
import { nextTick, onBeforeUnmount, ref, shallowRef, watch } from 'vue'
import * as echarts from 'echarts'

interface TrendSeries {
  name: string
  data: Array<{ time: string; value: number }>
  color?: string
}

interface EventPoint {
  time: string
  value: number
  label: string
  eventType: 'alert' | 'control'
}

const props = defineProps<{
  series: TrendSeries[]
  events?: EventPoint[]
  yAxisName?: string
}>()

const chartRef = ref<HTMLElement | null>(null)
const chart = shallowRef<echarts.ECharts | null>(null)

function render() {
  if (!chartRef.value) return
  if (!chart.value) chart.value = echarts.init(chartRef.value)

  const lineSeries: echarts.SeriesOption[] = props.series.map((item) => ({
    name: item.name,
    type: 'line',
    smooth: true,
    symbol: 'none',
    data: item.data.map((d) => [d.time, d.value]),
    lineStyle: item.color ? { color: item.color } : undefined
  }))

  const eventSeries: echarts.SeriesOption[] = [
    {
      name: '告警事件',
      type: 'scatter',
      data: (props.events || [])
        .filter((e) => e.eventType === 'alert')
        .map((e) => [e.time, e.value, e.label]),
      symbolSize: 10,
      itemStyle: { color: '#e6a23c' }
    },
    {
      name: '控制动作',
      type: 'scatter',
      data: (props.events || [])
        .filter((e) => e.eventType === 'control')
        .map((e) => [e.time, e.value, e.label]),
      symbolSize: 10,
      itemStyle: { color: '#409eff' }
    }
  ]

  chart.value.setOption(
    {
      tooltip: {
        trigger: 'axis',
        formatter: (params: unknown) => {
          const list = Array.isArray(params) ? params : [params]
          return list
            .map((p: any) => `${p.marker}${p.seriesName}: ${Array.isArray(p.data) ? p.data[1] : p.value}`)
            .join('<br/>')
        }
      },
      legend: { top: 8 },
      grid: { left: 48, right: 20, top: 36, bottom: 56 },
      dataZoom: [{ type: 'inside' }, { type: 'slider', height: 16 }],
      xAxis: { type: 'time' },
      yAxis: { type: 'value', name: props.yAxisName || '' },
      series: [...lineSeries, ...eventSeries]
    },
    { notMerge: true }
  )
}

watch(
  () => [props.series, props.events, props.yAxisName],
  async () => {
    await nextTick()
    render()
  },
  { deep: true, immediate: true }
)

onBeforeUnmount(() => {
  chart.value?.dispose()
})
</script>

<style scoped lang="scss">
.metric-trend-chart {
  height: 380px;
  width: 100%;
}
</style>
