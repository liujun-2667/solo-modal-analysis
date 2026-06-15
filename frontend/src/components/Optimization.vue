<template>
    <div class="optimization">
        <div class="toolbar">
            <el-button @click="runOptimization" type="primary" :loading="loading">
                运行优化
            </el-button>
            <el-form :model="optimizationParams" inline>
                <el-form-item label="最大迭代">
                    <el-input-number v-model="optimizationParams.maxIterations" :min="1" :max="100" />
                </el-form-item>
                <el-form-item label="容差">
                    <el-input-number v-model="optimizationParams.tolerance" :step="1e-8" />
                </el-form-item>
                <el-form-item label="模态阶数">
                    <el-input-number v-model="optimizationParams.numModes" :min="1" :max="50" />
                </el-form-item>
            </el-form>
        </div>

        <div class="constraint-section">
            <h4>频率约束</h4>
            <el-button @click="addConstraint" type="primary" size="small">添加约束</el-button>
            <el-table :data="frequencyConstraints" border>
                <el-table-column label="模态阶数">
                    <template #default="scope">
                        <el-input-number v-model="scope.row.modeIndex" :min="1" />
                    </template>
                </el-table-column>
                <el-table-column label="约束类型">
                    <template #default="scope">
                        <el-select v-model="scope.row.type">
                            <el-option label="下限" value="lower" />
                            <el-option label="上限" value="upper" />
                            <el-option label="双向" value="both" />
                        </el-select>
                    </template>
                </el-table-column>
                <el-table-column label="下限值(Hz)">
                    <template #default="scope">
                        <el-input-number 
                            v-model="scope.row.lowerBound" 
                            :disabled="scope.row.type === 'upper'"
                            :min="0" 
                        />
                    </template>
                </el-table-column>
                <el-table-column label="上限值(Hz)">
                    <template #default="scope">
                        <el-input-number 
                            v-model="scope.row.upperBound" 
                            :disabled="scope.row.type === 'lower'"
                            :min="0" 
                        />
                    </template>
                </el-table-column>
                <el-table-column>
                    <template #default="scope">
                        <el-button @click="removeConstraint(scope.$index)" type="danger" size="small">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
        </div>

        <div v-if="optimizationResult" class="results">
            <div class="convergence-section">
                <h4>收敛曲线</h4>
                <div ref="convergenceChartRef" class="chart"></div>
            </div>

            <div class="comparison-section">
                <div class="comparison-table">
                    <h4>设计变量对比</h4>
                    <el-table :data="designVarComparison" border>
                        <el-table-column label="设计变量" prop="name" />
                        <el-table-column label="初始值" prop="initial" />
                        <el-table-column label="优化后" prop="final" />
                        <el-table-column label="变化率" prop="change" />
                    </el-table>
                </div>
                <div class="comparison-table">
                    <h4>频率对比</h4>
                    <el-table :data="frequencyComparison" border>
                        <el-table-column label="模态阶数" prop="mode" />
                        <el-table-column label="初始频率(Hz)" prop="initial" />
                        <el-table-column label="优化后(Hz)" prop="final" />
                        <el-table-column label="变化率" prop="change" />
                    </el-table>
                </div>
            </div>

            <div class="summary-section">
                <el-card>
                    <div class="summary-item">
                        <span class="label">初始质量:</span>
                        <span class="value">{{ optimizationResult.initialMass.toExponential(3) }}</span>
                    </div>
                    <div class="summary-item">
                        <span class="label">优化后质量:</span>
                        <span class="value">{{ optimizationResult.finalMass.toExponential(3) }}</span>
                    </div>
                    <div class="summary-item">
                        <span class="label">质量减少:</span>
                        <span class="value" :class="{ 'highlight': massReduction > 0 }">
                            {{ massReduction.toFixed(2) }}%
                        </span>
                    </div>
                    <div class="summary-item">
                        <span class="label">收敛状态:</span>
                        <span class="value" :class="{ 'success': optimizationResult.converged, 'error': !optimizationResult.converged }">
                            {{ optimizationResult.converged ? '已收敛' : '未收敛' }}
                        </span>
                    </div>
                </el-card>
            </div>
        </div>

        <div v-else class="empty-state">
            <p>请设置设计变量和约束条件后运行优化</p>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive, computed, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import { api } from '../utils/api'
import type { Node, Element, Section, Constraint, DesignVariable, FrequencyConstraint, OptimizationResponse } from '../types'

const props = defineProps<{
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
    designVariables: DesignVariable[]
}>()

const emit = defineEmits<{
    (e: 'optimizationComplete', result: OptimizationResponse): void
}>()

const loading = ref(false)
const optimizationResult = ref<OptimizationResponse | null>(null)
const convergenceChartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

const optimizationParams = reactive({
    maxIterations: 30,
    tolerance: 1e-6,
    numModes: 10
})

const frequencyConstraints = ref<FrequencyConstraint[]>([
    { modeIndex: 1, type: 'lower', lowerBound: 50 }
])

const addConstraint = () => {
    frequencyConstraints.value.push({
        modeIndex: frequencyConstraints.value.length + 1,
        type: 'lower',
        lowerBound: 0
    })
}

const removeConstraint = (index: number) => {
    frequencyConstraints.value.splice(index, 1)
}

const designVarComparison = computed(() => {
    if (!optimizationResult.value) return []
    const enabledDVs = props.designVariables.filter(dv => dv.enabled)
    return enabledDVs.map((dv, i) => {
        const initial = optimizationResult.value?.initialDesignVariables[i] || 0
        const final = optimizationResult.value?.finalDesignVariables[i] || 0
        const change = initial !== 0 ? ((final - initial) / initial * 100) : 0
        return {
            name: `截面${dv.sectionId}.${dv.property}`,
            initial: initial.toExponential(3),
            final: final.toExponential(3),
            change: `${change >= 0 ? '+' : ''}${change.toFixed(2)}%`
        }
    })
})

const frequencyComparison = computed(() => {
    if (!optimizationResult.value) return []
    const initial = optimizationResult.value.initialFrequencies
    const final = optimizationResult.value.finalFrequencies
    return initial.map((freq, i) => {
        const finalFreq = final[i] || 0
        const change = freq !== 0 ? ((finalFreq - freq) / freq * 100) : 0
        return {
            mode: i + 1,
            initial: freq.toFixed(4),
            final: finalFreq.toFixed(4),
            change: `${change >= 0 ? '+' : ''}${change.toFixed(2)}%`
        }
    })
})

const massReduction = computed(() => {
    if (!optimizationResult.value) return 0
    const initial = optimizationResult.value.initialMass
    const final = optimizationResult.value.finalMass
    return initial !== 0 ? ((initial - final) / initial * 100) : 0
})

const runOptimization = async () => {
    loading.value = true
    try {
        const response = await api.solveOptimization({
            nodes: props.nodes,
            elements: props.elements,
            sections: props.sections,
            constraints: props.constraints,
            designVariables: props.designVariables,
            frequencyConstraints: frequencyConstraints.value,
            numModes: optimizationParams.numModes,
            maxIterations: optimizationParams.maxIterations,
            tolerance: optimizationParams.tolerance
        })

        if (response.success) {
            optimizationResult.value = response
            emit('optimizationComplete', response)
        } else {
            console.error(response.message)
        }
    } catch (error) {
        console.error('优化失败:', error)
    } finally {
        loading.value = false
    }
}

const updateConvergenceChart = () => {
    if (!chartInstance || !optimizationResult.value) return

    const iterations = optimizationResult.value.iterations
    const xData = iterations.map(i => i.iteration)
    const yData = iterations.map(i => i.objective)

    const option: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis'
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '3%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            data: xData,
            name: '迭代次数'
        },
        yAxis: {
            type: 'value',
            name: '总质量'
        },
        series: [{
            type: 'line',
            data: yData,
            smooth: true,
            symbol: 'circle',
            lineStyle: {
                width: 2,
                color: '#5470c6'
            },
            areaStyle: {
                color: new echarts.graphic.LinearGradient(0, 0, 0, 1, [
                    { offset: 0, color: 'rgba(84, 112, 198, 0.3)' },
                    { offset: 1, color: 'rgba(84, 112, 198, 0.05)' }
                ])
            }
        }]
    }

    chartInstance.setOption(option)
}

onMounted(() => {
    if (convergenceChartRef.value) {
        chartInstance = echarts.init(convergenceChartRef.value)
        updateConvergenceChart()
    }
})

onUnmounted(() => {
    if (chartInstance) {
        chartInstance.dispose()
    }
})

watch(optimizationResult, () => {
    updateConvergenceChart()
}, { deep: true })
</script>

<style scoped>
.optimization {
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

.constraint-section {
    padding: 15px 0;
    border-bottom: 1px solid #e8e8e8;
}

.constraint-section h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.results {
    flex: 1;
    display: flex;
    flex-direction: column;
    gap: 15px;
    overflow: auto;
    padding-top: 15px;
}

.convergence-section {
    flex: 1;
}

.convergence-section h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.convergence-section .chart {
    height: 200px;
}

.comparison-section {
    display: flex;
    gap: 15px;
}

.comparison-table {
    flex: 1;
}

.comparison-table h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.summary-section {
    padding-top: 10px;
}

.summary-section .el-card {
    display: flex;
    justify-content: space-around;
    flex-wrap: wrap;
}

.summary-item {
    display: flex;
    align-items: center;
    gap: 8px;
    padding: 10px 20px;
}

.summary-item .label {
    font-weight: 500;
    color: #606266;
}

.summary-item .value {
    font-weight: 600;
    color: #303133;
}

.summary-item .value.highlight {
    color: #67c23a;
}

.summary-item .value.success {
    color: #67c23a;
}

.summary-item .value.error {
    color: #f56c6c;
}

.empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #909399;
}
</style>