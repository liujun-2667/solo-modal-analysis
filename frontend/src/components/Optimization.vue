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
                <el-form-item label="目标函数容差">
                    <el-input-number v-model="optimizationParams.tolerance" :step="1e-8" />
                </el-form-item>
                <el-form-item label="约束违反容差">
                    <el-input-number v-model="optimizationParams.constraintTolerance" :step="1e-6" />
                </el-form-item>
                <el-form-item label="设计变量容差">
                    <el-input-number v-model="optimizationParams.designVarTolerance" :step="1e-4" />
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

            <div v-if="verificationResult" class="verification-section">
                <el-card title="优化结果验证报告" class="verification-card">
                    <div class="verification-summary">
                        <div class="verification-item">
                            <span class="label">质量减少百分比:</span>
                            <span class="value" :class="{ 'highlight': verificationResult.massReductionPercent > 0 }">
                                {{ verificationResult.massReductionPercent.toFixed(2) }}%
                            </span>
                        </div>
                    </div>

                    <div class="verification-table-section">
                        <h4>频率验证 (重新分析 vs 优化记录)</h4>
                        <el-table :data="frequencyVerificationData" border size="small">
                            <el-table-column label="模态阶数" prop="mode" />
                            <el-table-column label="优化记录值(Hz)" prop="optimized" />
                            <el-table-column label="重新分析值(Hz)" prop="reanalyzed" />
                            <el-table-column label="偏差(%)" prop="deviation">
                                <template #default="scope">
                                    <span :class="{ 'warning': scope.row.deviation >= 0.1 }">
                                        {{ scope.row.deviation.toFixed(4) }}%
                                    </span>
                                </template>
                            </el-table-column>
                        </el-table>
                    </div>

                    <div class="verification-table-section">
                        <h4>约束满足检查</h4>
                        <el-table :data="verificationResult.constraintChecks" border size="small">
                            <el-table-column label="模态阶数" prop="modeIndex" />
                            <el-table-column label="约束类型">
                                <template #default="scope">
                                    {{ scope.row.type === 'lower' ? '下限' : scope.row.type === 'upper' ? '上限' : '双向' }}
                                </template>
                            </el-table-column>
                            <el-table-column label="下限值(Hz)" prop="lowerBound">
                                <template #default="scope">
                                    {{ scope.row.lowerBound?.toFixed(4) || '-' }}
                                </template>
                            </el-table-column>
                            <el-table-column label="实际值(Hz)" prop="actualValue">
                                <template #default="scope">
                                    {{ scope.row.actualValue.toFixed(4) }}
                                </template>
                            </el-table-column>
                            <el-table-column label="上限值(Hz)" prop="upperBound">
                                <template #default="scope">
                                    {{ scope.row.upperBound?.toFixed(4) || '-' }}
                                </template>
                            </el-table-column>
                            <el-table-column label="满足">
                                <template #default="scope">
                                    <span :class="{ 'success': scope.row.satisfied, 'error': !scope.row.satisfied }">
                                        {{ scope.row.satisfied ? '✓' : '✗' }}
                                    </span>
                                </template>
                            </el-table-column>
                        </el-table>
                    </div>
                </el-card>
            </div>

            <div v-if="verifying" class="verifying-state">
                <el-card>
                    <div class="verifying-text">
                        <i class="el-icon-loading"></i>
                        <span>正在进行验证分析...</span>
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
const verificationResult = ref<VerificationResult | null>(null)
const verifying = ref(false)
const convergenceChartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null

interface VerificationResult {
    reanalyzedFrequencies: number[]
    frequencyDeviations: number[]
    constraintChecks: ConstraintCheck[]
    massReductionPercent: number
}

interface ConstraintCheck {
    modeIndex: number
    type: 'lower' | 'upper' | 'both'
    actualValue: number
    lowerBound?: number
    upperBound?: number
    satisfied: boolean
}

const optimizationParams = reactive({
    maxIterations: 30,
    tolerance: 1e-6,
    constraintTolerance: 1e-4,
    designVarTolerance: 0.001,
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

const frequencyVerificationData = computed(() => {
    if (!verificationResult.value || !optimizationResult.value) return []
    return verificationResult.value.reanalyzedFrequencies.map((freq, i) => ({
        mode: i + 1,
        optimized: optimizationResult.value?.finalFrequencies[i]?.toFixed(4) || '0',
        reanalyzed: freq.toFixed(4),
        deviation: verificationResult.value.frequencyDeviations[i]
    }))
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
            tolerance: optimizationParams.tolerance,
            constraintTolerance: optimizationParams.constraintTolerance,
            designVarTolerance: optimizationParams.designVarTolerance
        })

        if (response.success) {
            optimizationResult.value = response
            emit('optimizationComplete', response)
            await runVerification(response)
        } else {
            console.error(response.message)
        }
    } catch (error) {
        console.error('优化失败:', error)
    } finally {
        loading.value = false
    }
}

const runVerification = async (optimizationResult: OptimizationResponse) => {
    verifying.value = true
    try {
        const finalSections = [...props.sections]
        const enabledDVs = props.designVariables.filter(dv => dv.enabled)
        enabledDVs.forEach((dv, index) => {
            const sectionIndex = finalSections.findIndex(s => s.id === dv.sectionId)
            if (sectionIndex !== -1) {
                finalSections[sectionIndex][dv.property] = optimizationResult.finalDesignVariables[index]
            }
        })

        const response = await api.analyze({
            nodes: props.nodes,
            elements: props.elements,
            sections: finalSections,
            constraints: props.constraints,
            numModes: optimizationParams.numModes
        })

        if (response.success) {
            const reanalyzedFrequencies = response.modalResults.map(mr => mr.frequencyHz)
            const frequencyDeviations = reanalyzedFrequencies.map((freq, i) => {
                const optFreq = optimizationResult.finalFrequencies[i]
                return optFreq !== 0 ? Math.abs((freq - optFreq) / optFreq * 100) : 0
            })

            const constraintChecks: ConstraintCheck[] = frequencyConstraints.value.map(fc => {
                const actualValue = reanalyzedFrequencies[fc.modeIndex - 1] || 0
                let satisfied = true
                if (fc.type === 'lower' || fc.type === 'both') {
                    satisfied = satisfied && actualValue >= (fc.lowerBound || 0)
                }
                if (fc.type === 'upper' || fc.type === 'both') {
                    satisfied = satisfied && actualValue <= (fc.upperBound || Infinity)
                }
                return {
                    modeIndex: fc.modeIndex,
                    type: fc.type,
                    actualValue,
                    lowerBound: fc.lowerBound,
                    upperBound: fc.upperBound,
                    satisfied
                }
            })

            verificationResult.value = {
                reanalyzedFrequencies,
                frequencyDeviations,
                constraintChecks,
                massReductionPercent: massReduction.value
            }
        }
    } catch (error) {
        console.error('验证分析失败:', error)
    } finally {
        verifying.value = false
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

.verification-section {
    padding-top: 10px;
}

.verification-card {
    margin-top: 10px;
}

.verification-summary {
    display: flex;
    flex-wrap: wrap;
    gap: 20px;
    padding-bottom: 15px;
    border-bottom: 1px solid #e8e8e8;
    margin-bottom: 15px;
}

.verification-item {
    display: flex;
    align-items: center;
    gap: 8px;
}

.verification-item .label {
    font-weight: 500;
    color: #606266;
}

.verification-item .value {
    font-weight: 600;
    color: #303133;
}

.verification-table-section {
    margin-bottom: 15px;
}

.verification-table-section h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.verification-table-section :deep(.el-table) {
    font-size: 12px;
}

.verification-table-section .warning {
    color: #e6a23c;
    font-weight: 600;
}

.verification-table-section .success {
    color: #67c23a;
    font-weight: 600;
}

.verification-table-section .error {
    color: #f56c6c;
    font-weight: 600;
}

.verifying-state {
    padding-top: 10px;
}

.verifying-text {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 10px;
    padding: 20px;
    color: #606266;
}

.empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #909399;
}
</style>