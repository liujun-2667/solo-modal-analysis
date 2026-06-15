<template>
    <div class="param-scan">
        <div class="toolbar">
            <el-button @click="runParamScan" type="primary" :loading="loading">
                运行参数扫描
            </el-button>
            <el-radio-group v-model="scanType" class="scan-type">
                <el-radio label="1D">一维扫描</el-radio>
                <el-radio label="2D">二维扫描</el-radio>
            </el-radio-group>
        </div>

        <div class="scan-params">
            <el-form :model="scanParams" label-width="100px">
                <el-form-item label="设计变量">
                    <el-select v-model="scanParams.designVarId">
                        <el-option 
                            v-for="dv in availableDesignVars" 
                            :key="`${dv.sectionId}-${dv.property}`"
                            :label="`截面${dv.sectionId}.${dv.property}`"
                            :value="`${dv.sectionId}-${dv.property}`"
                        />
                    </el-select>
                </el-form-item>
                <el-form-item label="扫描起始值">
                    <el-input-number v-model="scanParams.scanStart" :step="0.001" />
                </el-form-item>
                <el-form-item label="扫描结束值">
                    <el-input-number v-model="scanParams.scanEnd" :step="0.001" />
                </el-form-item>
                <el-form-item label="扫描步数">
                    <el-input-number v-model="scanParams.numSteps" :min="2" :max="50" />
                </el-form-item>
                <el-form-item label="模态阶数">
                    <el-input-number v-model="scanParams.numModes" :min="1" :max="20" />
                </el-form-item>

                <div v-if="scanType === '2D'" class="second-variable">
                    <el-form-item label="第二设计变量">
                        <el-select v-model="scanParams.secondVarId">
                            <el-option 
                                v-for="dv in availableDesignVars" 
                                :key="`${dv.sectionId}-${dv.property}`"
                                :label="`截面${dv.sectionId}.${dv.property}`"
                                :value="`${dv.sectionId}-${dv.property}`"
                            />
                        </el-select>
                    </el-form-item>
                    <el-form-item label="第二变量起始">
                        <el-input-number v-model="scanParams.secondScanStart" :step="0.001" />
                    </el-form-item>
                    <el-form-item label="第二变量结束">
                        <el-input-number v-model="scanParams.secondScanEnd" :step="0.001" />
                    </el-form-item>
                    <el-form-item label="第二变量步数">
                        <el-input-number v-model="scanParams.secondNumSteps" :min="2" :max="30" />
                    </el-form-item>
                    <el-form-item label="显示模态">
                        <el-input-number v-model="scanParams.targetMode" :min="1" :max="20" />
                    </el-form-item>
                </div>
            </el-form>
        </div>

        <div v-if="scanResult" class="results">
            <div v-if="scanType === '1D'" class="chart-section">
                <h4>参数-频率曲线</h4>
                <div ref="chartRef" class="chart"></div>
            </div>

            <div v-else class="chart-section">
                <h4>参数等高线图 (第{{ scanParams.targetMode }}阶频率)</h4>
                <div ref="chartRef" class="chart"></div>
            </div>
        </div>

        <div v-else class="empty-state">
            <p>请设置扫描参数后运行参数扫描</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import { api } from '../utils/api'
import type { Node, Element, Section, Constraint, DesignVariable, ParamScanResponse } from '../types'

const props = defineProps<{
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    designVariables: DesignVariable[]
}>()

const loading = ref(false)
const scanType = ref<'1D' | '2D'>('1D')
const scanResult = ref<ParamScanResponse | null>(null)
const chartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

const scanParams = reactive({
    designVarId: '',
    scanStart: 0.001,
    scanEnd: 0.01,
    numSteps: 20,
    numModes: 5,
    secondVarId: '',
    secondScanStart: 0.001,
    secondScanEnd: 0.01,
    secondNumSteps: 10,
    targetMode: 1
})

const availableDesignVars = computed(() => {
    return props.designVariables.filter(dv => dv.enabled)
})

const getDesignVarFromString = (id: string): DesignVariable | undefined => {
    const [sectionId, property] = id.split('-')
    return props.designVariables.find(
        dv => dv.sectionId === parseInt(sectionId) && dv.property === property
    )
}

const runParamScan = async () => {
    if (!scanParams.designVarId) {
        console.error('请选择设计变量')
        return
    }

    loading.value = true
    try {
        const dv = getDesignVarFromString(scanParams.designVarId)
        if (!dv) {
            console.error('设计变量不存在')
            loading.value = false
            return
        }

        const request: any = {
            nodes: props.nodes,
            elements: props.elements,
            sections: props.sections,
            constraints: props.constraints,
            designVariable: {
                sectionId: dv.sectionId,
                property: dv.property,
                enabled: true,
                lowerBound: dv.lowerBound,
                upperBound: dv.upperBound,
                initialValue: dv.initialValue
            },
            scanStart: scanParams.scanStart,
            scanEnd: scanParams.scanEnd,
            numSteps: scanParams.numSteps,
            numModes: scanParams.numModes
        }

        if (scanType.value === '2D') {
            const secondDv = getDesignVarFromString(scanParams.secondVarId)
            if (secondDv) {
                request.secondVariable = {
                    sectionId: secondDv.sectionId,
                    property: secondDv.property,
                    enabled: true,
                    lowerBound: secondDv.lowerBound,
                    upperBound: secondDv.upperBound,
                    initialValue: secondDv.initialValue
                }
                request.secondScanStart = scanParams.secondScanStart
                request.secondScanEnd = scanParams.secondScanEnd
                request.secondNumSteps = scanParams.secondNumSteps
            }
        }

        const response = await api.performParamScan(request)

        if (response.success) {
            scanResult.value = response
        } else {
            console.error(response.message)
        }
    } catch (error) {
        console.error('参数扫描失败:', error)
    } finally {
        loading.value = false
    }
}

const updateChart = () => {
    if (!chartInstance || !scanResult.value) return

    if (scanType.value === '1D') {
        update1DChart()
    } else {
        update2DChart()
    }
}

const update1DChart = () => {
    if (!scanResult.value) return

    const scanValues = scanResult.value.scanValues
    const frequencies = scanResult.value.frequencies

    const series = frequencies[0]?.map((_, modeIndex) => ({
        name: `第${modeIndex + 1}阶`,
        type: 'line' as const,
        data: frequencies.map(freqs => freqs[modeIndex] || null),
        smooth: true,
        symbol: 'circle',
        symbolSize: 4
    }))

    const option: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis'
        },
        legend: {
            data: Array.from({ length: scanParams.numModes }, (_, i) => `第${i + 1}阶`),
            bottom: 0
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '15%',
            containLabel: true
        },
        xAxis: {
            type: 'value',
            name: '参数值'
        },
        yAxis: {
            type: 'value',
            name: '频率(Hz)'
        },
        series
    }

    chartInstance!.setOption(option)
}

const update2DChart = () => {
    if (!scanResult.value || !scanResult.value.secondScanValues || !scanResult.value.frequencies2D) return

    const xData = scanResult.value.scanValues
    const yData = scanResult.value.secondScanValues
    const zData = scanResult.value.frequencies2D

    const option: echarts.EChartsOption = {
        tooltip: {
            position: 'top',
            formatter: (params: any) => {
                return `X: ${params.data[0].toFixed(6)}<br/>Y: ${params.data[1].toFixed(6)}<br/>频率: ${params.data[2].toFixed(2)} Hz`
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
            name: '变量1',
            data: xData
        },
        yAxis: {
            type: 'value',
            name: '变量2',
            data: yData
        },
        visualMap: {
            min: Math.min(...zData.flat()),
            max: Math.max(...zData.flat()),
            calculable: true,
            orient: 'vertical',
            right: '5%',
            top: 'center',
            text: ['高', '低'],
            inRange: {
                color: ['#313695', '#4575b4', '#74add1', '#abd9e9', '#e0f3f8', '#ffffbf', '#fee090', '#fdae61', '#f46d43', '#d73027', '#a50026']
            }
        },
        series: [{
            name: '频率',
            type: 'heatmap',
            data: zData.flatMap((row, i) => 
                row.map((val, j) => [xData[i], yData[j], val])
            ),
            label: {
                show: false
            },
            emphasis: {
                itemStyle: {
                    shadowBlur: 10,
                    shadowColor: 'rgba(0, 0, 0, 0.5)'
                }
            }
        }]
    }

    chartInstance!.setOption(option)
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

watch([scanResult, scanType], () => {
    updateChart()
}, { deep: true })
</script>

<style scoped>
.param-scan {
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

.scan-type {
    margin-left: auto;
}

.scan-params {
    padding: 15px 0;
    border-bottom: 1px solid #e8e8e8;
}

.scan-params .el-form {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 15px;
}

.second-variable {
    grid-column: 1 / -1;
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: 15px;
    padding-top: 15px;
    border-top: 1px dashed #e8e8e8;
    margin-top: 15px;
}

.results {
    flex: 1;
    overflow: auto;
    padding-top: 15px;
}

.chart-section {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.chart-section h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.chart {
    flex: 1;
    min-height: 300px;
}

.empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #909399;
}
</style>