<template>
    <div class="sensitivity-analysis">
        <div class="toolbar">
            <el-button @click="runSensitivity" type="primary" :loading="loading">
                运行灵敏度分析
            </el-button>
            <el-form :model="analysisParams" inline>
                <el-form-item label="模态阶数">
                    <el-input-number v-model="analysisParams.numModes" :min="1" :max="50" />
                </el-form-item>
            </el-form>
        </div>

        <div v-if="sensitivityResults.length > 0" class="results">
            <div class="table-section">
                <h4>灵敏度矩阵</h4>
                <el-table :data="sensitivityResults" border>
                    <el-table-column label="模态阶数" prop="modeIndex" />
                    <el-table-column label="频率(Hz)" prop="frequencyHz">
                        <template #default="scope">
                            {{ scope.row.frequencyHz.toFixed(4) }}
                        </template>
                    </el-table-column>
                    <el-table-column 
                        v-for="name in designVarNames" 
                        :key="name" 
                        :label="name"
                    >
                        <template #default="scope">
                            {{ getSensitivity(scope.row.sensitivities, name) }}
                        </template>
                    </el-table-column>
                </el-table>
            </div>

            <div class="chart-section">
                <h4>灵敏度柱状图</h4>
                <el-form :model="chartParams" inline>
                    <el-form-item label="选择模态">
                        <el-select v-model="chartParams.selectedMode" @change="updateChart">
                            <el-option 
                                v-for="result in sensitivityResults" 
                                :key="result.modeIndex"
                                :label="`第${result.modeIndex}阶 (${result.frequencyHz.toFixed(2)} Hz)`"
                                :value="result.modeIndex"
                            />
                        </el-select>
                    </el-form-item>
                </el-form>
                <div ref="chartRef" class="chart"></div>
            </div>
        </div>

        <div v-else class="empty-state">
            <p>请先运行灵敏度分析</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import type { SensitivityResult, SensitivityItem } from '../types'
import { api } from '../utils/api'
import type { Node, Element, Section, Constraint, DesignVariable } from '../types'

const props = defineProps<{
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    designVariables: DesignVariable[]
}>()

const loading = ref(false)
const sensitivityResults = ref<SensitivityResult[]>([])
const designVarNames = ref<string[]>([])
const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

const analysisParams = reactive({
    numModes: 10
})

const chartParams = reactive({
    selectedMode: 1
})

const getSensitivity = (sensitivities: SensitivityItem[], name: string): string => {
    const item = sensitivities.find(s => s.designVarId === name)
    if (item) {
        return item.sensitivity.toExponential(3)
    }
    return '-'
}

const runSensitivity = async () => {
    loading.value = true
    try {
        const response = await api.calculateSensitivity({
            nodes: props.nodes,
            elements: props.elements,
            sections: props.sections,
            constraints: props.constraints,
            designVariables: props.designVariables,
            frequencyConstraints: [],
            numModes: analysisParams.numModes,
            maxIterations: 50,
            tolerance: 1e-6
        })

        if (response.success) {
            sensitivityResults.value = response.results
            designVarNames.value = response.designVarNames
            if (response.results.length > 0) {
                chartParams.selectedMode = response.results[0].modeIndex
            }
        } else {
            console.error(response.message)
        }
    } catch (error) {
        console.error('灵敏度分析失败:', error)
    } finally {
        loading.value = false
    }
}

const updateChart = () => {
    if (!chartInstance || sensitivityResults.value.length === 0) return

    const result = sensitivityResults.value.find(r => r.modeIndex === chartParams.selectedMode)
    if (!result) return

    const data = result.sensitivities.map(s => ({
        name: s.designVarId,
        value: s.sensitivity
    }))

    const option: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis',
            axisPointer: {
                type: 'shadow'
            }
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            data: data.map(d => d.name),
            axisLabel: {
                rotate: 45,
                fontSize: 10
            }
        },
        yAxis: {
            type: 'value',
            name: '灵敏度 (df/dx)'
        },
        series: [{
            type: 'bar',
            data: data.map(d => ({
                value: d.value,
                itemStyle: {
                    color: d.value >= 0 ? '#5470c6' : '#91cc75'
                }
            })),
            label: {
                show: true,
                position: 'top',
                formatter: '{c:.2e}',
                fontSize: 8
            }
        }]
    }

    chartInstance.setOption(option)
}

onMounted(() => {
    if (chartRef.value) {
        chartInstance = echarts.init(chartRef.value)
        updateChart()
    }
})

onUnmounted(() => {
    if (chartInstance) {
        chartInstance.dispose()
    }
})

watch(sensitivityResults, () => {
    updateChart()
}, { deep: true })
</script>

<style scoped>
.sensitivity-analysis {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.toolbar {
    display: flex;
    align-items: center;
    gap: 15px;
    padding-bottom: 15px;
    border-bottom: 1px solid #e8e8e8;
}

.results {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 15px;
    overflow: auto;
    padding-top: 15px;
}

.table-section {
    overflow: auto;
}

.table-section h4,
.chart-section h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.chart-section {
    flex: 1;
    display: flex;
    flex-direction: column;
}

.chart {
    flex: 1;
    min-height: 200px;
}

.empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #909399;
}
</style>