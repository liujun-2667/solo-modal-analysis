<template>
    <div class="mac-matrix">
        <el-card title="模态MAC矩阵">
            <div class="load-buttons">
                <el-button @click="loadFirstSet" type="primary">加载第一组结果</el-button>
                <el-button @click="loadSecondSet" type="success">加载第二组结果</el-button>
            </div>
            
            <div class="result-info" v-if="firstSetLoaded">
                <span>第一组: {{ firstSetCount }} 阶模态</span>
            </div>
            <div class="result-info" v-if="secondSetLoaded">
                <span>第二组: {{ secondSetCount }} 阶模态</span>
            </div>

            <div class="heatmap-container" v-if="showHeatmap">
                <div class="heatmap-wrapper">
                    <div class="heatmap">
                        <div
                            v-for="(row, i) in macMatrix"
                            :key="i"
                            class="heatmap-row"
                        >
                            <div
                                v-for="(value, j) in row"
                                :key="j"
                                class="heatmap-cell"
                                :style="{ backgroundColor: getColor(value) }"
                                @mouseenter="showTooltip(value, i, j, $event)"
                                @mouseleave="hideTooltip"
                            >
                                {{ value.toFixed(3) }}
                            </div>
                        </div>
                    </div>
                    <div class="y-axis">
                        <span v-for="i in firstSetCount" :key="i">{{ i }}</span>
                    </div>
                </div>
                <div class="x-axis">
                    <span v-for="i in secondSetCount" :key="i">{{ i }}</span>
                </div>
                <div class="color-bar">
                    <span>0</span>
                    <div class="color-gradient"></div>
                    <span>1</span>
                </div>
            </div>

            <div class="empty-state" v-if="!showHeatmap">
                <p>请加载两组分析结果以计算MAC矩阵</p>
            </div>
        </el-card>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import type { ModalResult } from '../types'

const props = defineProps<{
    currentModalResults: ModalResult[]
}>()

const emit = defineEmits<{
    (e: 'clearSets'): void
}>()

const firstSet = ref<ModalResult[]>([])
const secondSet = ref<ModalResult[]>([])

const firstSetLoaded = computed(() => firstSet.value.length > 0)
const secondSetLoaded = computed(() => secondSet.value.length > 0)
const firstSetCount = computed(() => firstSet.value.length)
const secondSetCount = computed(() => secondSet.value.length)

const showHeatmap = computed(() => firstSetLoaded.value && secondSetLoaded.value)

const macMatrix = computed(() => {
    if (!firstSetLoaded.value || !secondSetLoaded.value) return []
    
    const matrix: number[][] = []
    const modes1 = firstSet.value.filter(m => !m.isRigidBody)
    const modes2 = secondSet.value.filter(m => !m.isRigidBody)
    
    for (let i = 0; i < modes1.length; i++) {
        const row: number[] = []
        for (let j = 0; j < modes2.length; j++) {
            row.push(calculateMAC(modes1[i].modeShape, modes2[j].modeShape))
        }
        matrix.push(row)
    }
    
    return matrix
})

const calculateMAC = (phi1: number[], phi2: number[]): number => {
    let dotProduct = 0
    let norm1 = 0
    let norm2 = 0
    
    const minLen = Math.min(phi1.length, phi2.length)
    
    for (let i = 0; i < minLen; i++) {
        dotProduct += phi1[i] * phi2[i]
        norm1 += phi1[i] * phi1[i]
        norm2 += phi2[i] * phi2[i]
    }
    
    if (norm1 === 0 || norm2 === 0) return 0
    
    return (dotProduct * dotProduct) / (norm1 * norm2)
}

const getColor = (value: number): string => {
    const r = Math.floor(255 * (1 - value))
    const g = Math.floor(255 * value)
    const b = Math.floor(200 * (1 - value))
    return `rgb(${r}, ${g}, ${b})`
}

const loadFirstSet = () => {
    if (props.currentModalResults.length > 0) {
        firstSet.value = JSON.parse(JSON.stringify(props.currentModalResults))
    }
}

const loadSecondSet = () => {
    if (props.currentModalResults.length > 0) {
        secondSet.value = JSON.parse(JSON.stringify(props.currentModalResults))
    }
}

const showTooltip = (value: number, i: number, j: number, event: MouseEvent) => {
    hideTooltip()
    const tooltip = document.createElement('div')
    tooltip.className = 'mac-tooltip'
    tooltip.textContent = `MAC(${i + 1}, ${j + 1}) = ${value.toFixed(4)}`
    tooltip.style.left = `${event.clientX + 10}px`
    tooltip.style.top = `${event.clientY + 10}px`
    document.body.appendChild(tooltip)
}

const hideTooltip = () => {
    const tooltips = document.querySelectorAll('.mac-tooltip')
    tooltips.forEach(tooltip => tooltip.remove())
}
</script>

<style scoped>
.mac-matrix {
    height: 100%;
}

.load-buttons {
    display: flex;
    gap: 10px;
    margin-bottom: 15px;
}

.result-info {
    padding: 5px 10px;
    background: #f5f5f5;
    border-radius: 4px;
    margin-bottom: 10px;
    font-size: 12px;
    color: #666;
}

.heatmap-container {
    margin-top: 15px;
}

.heatmap-wrapper {
    display: flex;
}

.y-axis {
    display: flex;
    flex-direction: column;
    justify-content: space-around;
    padding-right: 5px;
    text-align: right;
    font-size: 10px;
    color: #666;
    min-width: 20px;
}

.heatmap {
    display: flex;
    flex-direction: column;
    gap: 2px;
}

.heatmap-row {
    display: flex;
    gap: 2px;
}

.heatmap-cell {
    width: 30px;
    height: 30px;
    display: flex;
    align-items: center;
    justify-content: center;
    font-size: 10px;
    color: #333;
    border-radius: 2px;
    cursor: pointer;
    transition: transform 0.1s;
}

.heatmap-cell:hover {
    transform: scale(1.1);
    z-index: 10;
}

.x-axis {
    display: flex;
    justify-content: space-around;
    padding-left: 25px;
    margin-top: 5px;
    font-size: 10px;
    color: #666;
}

.color-bar {
    display: flex;
    align-items: center;
    gap: 5px;
    margin-top: 15px;
    font-size: 12px;
    color: #666;
}

.color-gradient {
    flex: 1;
    height: 20px;
    border-radius: 4px;
    background: linear-gradient(to right, rgb(255, 0, 200), rgb(0, 255, 0));
}

.empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 200px;
    color: #909399;
}

.mac-tooltip {
    position: fixed;
    background: rgba(0, 0, 0, 0.8);
    color: white;
    padding: 8px 12px;
    border-radius: 4px;
    font-size: 12px;
    z-index: 1000;
    pointer-events: none;
}
</style>