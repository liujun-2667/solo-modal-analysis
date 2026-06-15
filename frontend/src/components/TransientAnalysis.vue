<template>
    <div class="transient-analysis">
        <el-card title="瞬态响应分析">
            <el-form :model="transientParams" label-width="100px">
                <el-form-item label="波形类型">
                    <el-select v-model="transientParams.waveformType" placeholder="请选择波形类型">
                        <el-option label="脉冲载荷" value="impulse" />
                        <el-option label="阶跃载荷" value="step" />
                        <el-option label="半正弦冲击" value="halfsine" />
                    </el-select>
                </el-form-item>
                <el-form-item label="激励幅值">
                    <el-input-number v-model="transientParams.amplitude" :min="0.001" :step="0.1" />
                </el-form-item>
                <el-form-item label="持续时间(s)">
                    <el-input-number v-model="transientParams.duration" :min="0.001" :step="0.001" />
                </el-form-item>
                <el-form-item label="激励节点">
                    <el-select v-model="transientParams.excitationNode" placeholder="请选择节点">
                        <el-option v-for="node in nodes" :key="node.id" :label="`节点${node.id}`" :value="node.id" />
                    </el-select>
                </el-form-item>
                <el-form-item label="激励方向">
                    <el-select v-model="transientParams.direction" placeholder="请选择方向">
                        <el-option label="X" value="X" />
                        <el-option label="Y" value="Y" />
                        <el-option label="Z" value="Z" />
                    </el-select>
                </el-form-item>
                <el-form-item label="观测节点">
                    <el-select v-model="transientParams.observationNode" placeholder="请选择节点">
                        <el-option v-for="node in nodes" :key="node.id" :label="`节点${node.id}`" :value="node.id" />
                    </el-select>
                </el-form-item>
                <el-form-item label="观测方向">
                    <el-select v-model="transientParams.observationDirection" placeholder="请选择方向">
                        <el-option label="X" value="X" />
                        <el-option label="Y" value="Y" />
                        <el-option label="Z" value="Z" />
                    </el-select>
                </el-form-item>
                <el-form-item label="时间步长(s)">
                    <el-input-number v-model="transientParams.timeStep" :min="0.0001" :step="0.0001" />
                </el-form-item>
                <el-form-item label="总计算时长(s)">
                    <el-input-number v-model="transientParams.totalTime" :min="0.01" :step="0.1" />
                </el-form-item>
                <el-form-item label="阻尼比">
                    <el-input-number v-model="transientParams.dampingRatio" :min="0" :max="1" :step="0.01" />
                </el-form-item>
            </el-form>
            <el-button @click="runTransient" type="primary" style="width: 100%; margin-top: 10px;">
                计算瞬态响应
            </el-button>
        </el-card>

        <el-card title="位移响应曲线" v-if="transientResult">
            <div ref="chartRef" class="chart-container"></div>
            <div class="animation-controls" v-if="transientResult">
                <el-button @click="toggleAnimation" :icon="isAnimating ? VideoPause : VideoPlay">
                    {{ isAnimating ? '暂停' : '播放' }}
                </el-button>
                <el-button @click="resetAnimation">重置</el-button>
                <el-slider v-model="animationSpeed" :min="0.1" :max="5" :step="0.1" style="width: 200px;" />
                <span>{{ animationSpeed.toFixed(1) }}x</span>
            </div>
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, watch } from 'vue'
import * as echarts from 'echarts'
import { VideoPlay, VideoPause } from '@element-plus/icons-vue'
import { api } from '../utils/api'
import type { Node, Element, Section, Constraint } from '../types'

const props = defineProps<{
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
}>()

const emit = defineEmits<{
    (e: 'animationUpdate', displacements: number[]): void
}>()

const transientParams = reactive({
    waveformType: 'impulse' as 'impulse' | 'step' | 'halfsine',
    amplitude: 1000,
    duration: 0.01,
    excitationNode: 1,
    direction: 'X',
    observationNode: 1,
    observationDirection: 'X',
    timeStep: 0.001,
    totalTime: 1,
    dampingRatio: 0.02
})

const transientResult = ref<{ timePoints: number[], displacements: number[], allDisplacements: number[][] } | null>(null)
const chartRef = ref<HTMLDivElement | null>(null)
let chartInstance: echarts.ECharts | null = null
const isAnimating = ref(false)
const animationSpeed = ref(1)
let animationId: number | null = null
let currentTimeIndex = ref(0)

const runTransient = async () => {
    try {
        const response = await api.calculateTransient({
            nodes: props.nodes,
            elements: props.elements,
            sections: props.sections,
            constraints: props.constraints,
            ...transientParams
        })

        if (response.success) {
            transientResult.value = {
                timePoints: response.timePoints,
                displacements: response.displacements,
                allDisplacements: response.allDisplacements
            }
            updateChart()
            currentTimeIndex.value = 0
            isAnimating.value = true
            startAnimation()
        } else {
            alert(response.message)
        }
    } catch (error) {
        alert('瞬态响应计算失败')
        console.error(error)
    }
}

const updateChart = () => {
    if (!chartRef.value || !transientResult.value) return

    if (!chartInstance) {
        chartInstance = echarts.init(chartRef.value)
    }

    const option: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis',
            formatter: (params: any) => {
                const data = params[0]
                return `时间: ${data.name.toFixed(4)}s<br/>位移: ${data.value.toFixed(6)}m`
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: {
            type: 'value',
            name: '时间 (s)',
            nameLocation: 'middle',
            nameGap: 30
        },
        yAxis: {
            type: 'value',
            name: '位移 (m)',
            nameLocation: 'middle',
            nameGap: 40
        },
        series: [{
            name: '位移响应',
            type: 'line',
            data: transientResult.value.timePoints.map((t, i) => ({
                value: [t, transientResult.value!.displacements[i]]
            })),
            smooth: true,
            lineStyle: {
                width: 2,
                color: '#667eea'
            },
            areaStyle: {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                    { offset: 0, color: 'rgba(102, 126, 234, 0.3)' },
                    { offset: 1, color: 'rgba(102, 126, 234, 0.05)' }
                ])
            }
        }]
    }

    chartInstance.setOption(option)
}

const startAnimation = () => {
    if (!transientResult.value || !isAnimating.value) return

    const animate = () => {
        if (!isAnimating.value) return

        currentTimeIndex.value = (currentTimeIndex.value + 1) % transientResult.value.allDisplacements.length
        
        if (transientResult.value.allDisplacements[currentTimeIndex.value]) {
            emit('animationUpdate', transientResult.value.allDisplacements[currentTimeIndex.value])
        }

        const delay = Math.max(10, 100 / animationSpeed.value)
        animationId = window.setTimeout(animate, delay)
    }

    animate()
}

const toggleAnimation = () => {
    isAnimating.value = !isAnimating.value
    if (isAnimating.value) {
        startAnimation()
    } else if (animationId) {
        clearTimeout(animationId)
    }
}

const resetAnimation = () => {
    isAnimating.value = false
    if (animationId) {
        clearTimeout(animationId)
    }
    currentTimeIndex.value = 0
    if (transientResult.value?.allDisplacements[0]) {
        emit('animationUpdate', transientResult.value.allDisplacements[0])
    }
}

watch(() => props.nodes, () => {
    if (props.nodes.length > 0) {
        transientParams.excitationNode = props.nodes[0].id
        transientParams.observationNode = props.nodes[0].id
    }
}, { immediate: true })

onMounted(() => {
    const handleResize = () => {
        chartInstance?.resize()
    }
    window.addEventListener('resize', handleResize)
})
</script>

<style scoped>
.transient-analysis {
    display: flex;
    flex-direction: column;
    gap: 15px;
}

.chart-container {
    height: 250px;
    width: 100%;
}

.animation-controls {
    display: flex;
    align-items: center;
    gap: 10px;
    margin-top: 10px;
    padding-top: 10px;
    border-top: 1px solid #eee;
}
</style>