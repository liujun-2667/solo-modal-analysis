<template>
    <div class="optimization-history">
        <div class="toolbar">
            <el-button @click="clearHistory" type="danger" size="small" :disabled="historyRecords.length === 0">
                清空历史
            </el-button>
            <el-button 
                @click="showCompareDialog" 
                type="primary" 
                size="small" 
                :disabled="selectedRecords.length !== 2"
            >
                对比选中记录
            </el-button>
        </div>

        <div v-if="historyRecords.length > 0" class="records-list">
            <el-table :data="historyRecords" border :selection-change="handleSelectionChange">
                <el-table-column type="selection" :selectable="() => selectedRecords.length < 2"></el-table-column>
                <el-table-column label="名称">
                    <template #default="scope">
                        <span>{{ scope.row.name }}</span>
                    </template>
                </el-table-column>
                <el-table-column label="时间">
                    <template #default="scope">
                        <span>{{ formatTime(scope.row.timestamp) }}</span>
                    </template>
                </el-table-column>
                <el-table-column label="初始质量">
                    <template #default="scope">
                        <span>{{ scope.row.initialMass.toExponential(3) }}</span>
                    </template>
                </el-table-column>
                <el-table-column label="最终质量">
                    <template #default="scope">
                        <span>{{ scope.row.finalMass.toExponential(3) }}</span>
                    </template>
                </el-table-column>
                <el-table-column label="收敛">
                    <template #default="scope">
                        <el-tag :type="scope.row.converged ? 'success' : 'warning'">
                            {{ scope.row.converged ? '是' : '否' }}
                        </el-tag>
                    </template>
                </el-table-column>
                <el-table-column label="操作">
                    <template #default="scope">
                        <el-button @click="viewRecord(scope.row)" type="primary" size="small">查看</el-button>
                        <el-button @click="deleteRecord(scope.row.id)" type="danger" size="small">删除</el-button>
                    </template>
                </el-table-column>
            </el-table>
        </div>

        <div v-else class="empty-state">
            <p>暂无优化历史记录</p>
        </div>

        <el-dialog v-if="selectedRecord" title="优化结果详情" :visible="!!selectedRecord" @close="selectedRecord = null" width="80%">
            <div class="detail-content">
                <div class="summary-row">
                    <span class="label">名称:</span>
                    <span class="value">{{ selectedRecord.name }}</span>
                </div>
                <div class="summary-row">
                    <span class="label">时间:</span>
                    <span class="value">{{ formatTime(selectedRecord.timestamp) }}</span>
                </div>
                <div class="summary-row">
                    <span class="label">初始质量:</span>
                    <span class="value">{{ selectedRecord.initialMass.toExponential(3) }}</span>
                </div>
                <div class="summary-row">
                    <span class="label">最终质量:</span>
                    <span class="value">{{ selectedRecord.finalMass.toExponential(3) }}</span>
                </div>
                <div class="summary-row">
                    <span class="label">质量减少:</span>
                    <span class="value" :class="{ 'highlight': massReduction > 0 }">
                        {{ massReduction.toFixed(2) }}%
                    </span>
                </div>
                <div class="summary-row">
                    <span class="label">收敛状态:</span>
                    <span class="value" :class="{ 'success': selectedRecord.converged, 'error': !selectedRecord.converged }">
                        {{ selectedRecord.converged ? '已收敛' : '未收敛' }}
                    </span>
                </div>

                <div class="chart-row">
                    <h4>收敛曲线</h4>
                    <div ref="detailChartRef" class="chart"></div>
                </div>

                <div class="tables-row">
                    <div class="detail-table">
                        <h4>设计变量对比</h4>
                        <el-table :data="designVarComparison" border size="small">
                            <el-table-column label="变量" prop="name" />
                            <el-table-column label="初始值" prop="initial" />
                            <el-table-column label="最终值" prop="final" />
                            <el-table-column label="变化率" prop="change" />
                        </el-table>
                    </div>
                    <div class="detail-table">
                        <h4>频率对比</h4>
                        <el-table :data="frequencyComparison" border size="small">
                            <el-table-column label="模态" prop="mode" />
                            <el-table-column label="初始频率" prop="initial" />
                            <el-table-column label="最终频率" prop="final" />
                            <el-table-column label="变化率" prop="change" />
                        </el-table>
                    </div>
                </div>
            </div>
        </el-dialog>

        <el-dialog title="优化记录对比" :visible="compareDialogVisible" @close="compareDialogVisible = false" width="90%">
            <div class="compare-content">
                <div class="compare-chart-row">
                    <h4>收敛曲线对比</h4>
                    <div ref="compareChartRef" class="chart"></div>
                </div>

                <div class="compare-tables-row">
                    <div class="compare-table">
                        <h4>设计变量终值对比</h4>
                        <el-table :data="designVarCompareData" border size="small">
                            <el-table-column label="变量" prop="name" />
                            <el-table-column :label="selectedRecords[0]?.name" prop="record1" />
                            <el-table-column :label="selectedRecords[1]?.name" prop="record2" />
                            <el-table-column label="差异(%)" prop="diff">
                                <template #default="scope">
                                    <span :class="{ 'significant': scope.row.significant }">
                                        {{ scope.row.diff }}%
                                    </span>
                                </template>
                            </el-table-column>
                        </el-table>
                    </div>
                    <div class="compare-table">
                        <h4>频率结果对比</h4>
                        <el-table :data="frequencyCompareData" border size="small">
                            <el-table-column label="模态" prop="mode" />
                            <el-table-column :label="selectedRecords[0]?.name" prop="record1" />
                            <el-table-column :label="selectedRecords[1]?.name" prop="record2" />
                            <el-table-column label="差异(%)" prop="diff">
                                <template #default="scope">
                                    <span :class="{ 'significant': scope.row.significant }">
                                        {{ scope.row.diff }}%
                                    </span>
                                </template>
                            </el-table-column>
                        </el-table>
                    </div>
                </div>
            </div>
        </el-dialog>
    </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'
import type { OptimizationHistoryRecord } from '../types'

const props = defineProps<{
    designVarNames: string[]
}>()

const emit = defineEmits<{
    (e: 'restore', record: OptimizationHistoryRecord): void
}>()

const historyRecords = ref<OptimizationHistoryRecord[]>([])
const selectedRecord = ref<OptimizationHistoryRecord | null>(null)
const selectedRecords = ref<OptimizationHistoryRecord[]>([])
const compareDialogVisible = ref(false)
const detailChartRef = ref<HTMLElement | null>(null)
const compareChartRef = ref<HTMLElement | null>(null)
let chartInstance: echarts.ECharts | null = null
let compareChartInstance: echarts.ECharts | null = null

const massReduction = computed(() => {
    if (!selectedRecord.value) return 0
    const initial = selectedRecord.value.initialMass
    const final = selectedRecord.value.finalMass
    return initial !== 0 ? ((initial - final) / initial * 100) : 0
})

const designVarComparison = computed(() => {
    if (!selectedRecord.value) return []
    return selectedRecord.value.designVarNames.map((name, i) => {
        const initial = selectedRecord.value?.initialDesignVariables[i] || 0
        const final = selectedRecord.value?.finalDesignVariables[i] || 0
        const change = initial !== 0 ? ((final - initial) / initial * 100) : 0
        return {
            name,
            initial: initial.toExponential(3),
            final: final.toExponential(3),
            change: `${change >= 0 ? '+' : ''}${change.toFixed(2)}%`
        }
    })
})

const frequencyComparison = computed(() => {
    if (!selectedRecord.value) return []
    const initial = selectedRecord.value.initialFrequencies
    const final = selectedRecord.value.finalFrequencies
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

const formatTime = (timestamp: string): string => {
    const date = new Date(timestamp)
    return date.toLocaleString('zh-CN')
}

const handleSelectionChange = (val: OptimizationHistoryRecord[]) => {
    selectedRecords.value = val
}

const showCompareDialog = () => {
    if (selectedRecords.value.length === 2) {
        compareDialogVisible.value = true
    }
}

const designVarCompareData = computed(() => {
    if (selectedRecords.value.length !== 2) return []
    const record1 = selectedRecords.value[0]
    const record2 = selectedRecords.value[1]
    const allNames = [...new Set([...record1.designVarNames, ...record2.designVarNames])]
    
    return allNames.map(name => {
        const idx1 = record1.designVarNames.indexOf(name)
        const idx2 = record2.designVarNames.indexOf(name)
        const final1 = idx1 !== -1 ? record1.finalDesignVariables[idx1] : 0
        const final2 = idx2 !== -1 ? record2.finalDesignVariables[idx2] : 0
        const diff = final1 !== 0 ? Math.abs((final2 - final1) / final1 * 100) : 0
        return {
            name,
            record1: final1.toExponential(3),
            record2: final2.toExponential(3),
            diff: diff.toFixed(2),
            significant: diff > 10
        }
    })
})

const frequencyCompareData = computed(() => {
    if (selectedRecords.value.length !== 2) return []
    const record1 = selectedRecords.value[0]
    const record2 = selectedRecords.value[1]
    const maxLen = Math.max(record1.finalFrequencies.length, record2.finalFrequencies.length)
    
    return Array.from({ length: maxLen }, (_, i) => {
        const freq1 = record1.finalFrequencies[i] || 0
        const freq2 = record2.finalFrequencies[i] || 0
        const diff = freq1 !== 0 ? Math.abs((freq2 - freq1) / freq1 * 100) : 0
        return {
            mode: i + 1,
            record1: freq1.toFixed(4),
            record2: freq2.toFixed(4),
            diff: diff.toFixed(2),
            significant: diff > 10
        }
    })
})

const updateCompareChart = () => {
    if (!compareChartInstance || selectedRecords.value.length !== 2) return

    const record1 = selectedRecords.value[0]
    const record2 = selectedRecords.value[1]
    
    const xData1 = record1.iterations.map(i => i.iteration)
    const yData1 = record1.iterations.map(i => i.objective)
    const xData2 = record2.iterations.map(i => i.iteration)
    const yData2 = record2.iterations.map(i => i.objective)

    const option: echarts.EChartsOption = {
        tooltip: {
            trigger: 'axis'
        },
        legend: {
            data: [record1.name, record2.name],
            bottom: 0
        },
        grid: {
            left: '3%',
            right: '4%',
            bottom: '15%',
            containLabel: true
        },
        xAxis: {
            type: 'category',
            data: [...new Set([...xData1, ...xData2])].sort((a, b) => a - b),
            name: '迭代次数'
        },
        yAxis: {
            type: 'value',
            name: '总质量'
        },
        series: [
            {
                name: record1.name,
                type: 'line',
                data: xData1.map((_, i) => yData1[i]),
                smooth: true,
                symbol: 'circle',
                lineStyle: {
                    width: 2,
                    color: '#5470c6'
                }
            },
            {
                name: record2.name,
                type: 'line',
                data: xData2.map((_, i) => yData2[i]),
                smooth: true,
                symbol: 'circle',
                lineStyle: {
                    width: 2,
                    color: '#91cc75'
                }
            }
        ]
    }

    compareChartInstance.setOption(option)
}

const viewRecord = (record: OptimizationHistoryRecord) => {
    selectedRecord.value = record
}

const deleteRecord = (id: string) => {
    const index = historyRecords.value.findIndex(r => r.id === id)
    if (index !== -1) {
        historyRecords.value.splice(index, 1)
        saveHistory()
    }
}

const clearHistory = () => {
    historyRecords.value = []
    saveHistory()
}

const addRecord = (record: Omit<OptimizationHistoryRecord, 'id' | 'timestamp' | 'name'>) => {
    const newRecord: OptimizationHistoryRecord = {
        ...record,
        id: Date.now().toString(),
        timestamp: new Date().toISOString(),
        name: `优化 ${historyRecords.value.length + 1}`
    }
    historyRecords.value.unshift(newRecord)
    saveHistory()
}

const saveHistory = () => {
    localStorage.setItem('optimizationHistory', JSON.stringify(historyRecords.value))
}

const loadHistory = () => {
    const saved = localStorage.getItem('optimizationHistory')
    if (saved) {
        try {
            historyRecords.value = JSON.parse(saved)
        } catch (e) {
            console.error('加载历史记录失败:', e)
        }
    }
}

const updateDetailChart = () => {
    if (!chartInstance || !selectedRecord.value) return

    const iterations = selectedRecord.value.iterations
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
            }
        }]
    }

    chartInstance.setOption(option)
}

onMounted(() => {
    loadHistory()
    if (detailChartRef.value) {
        chartInstance = echarts.init(detailChartRef.value)
        updateDetailChart()
    }
})

onUnmounted(() => {
    if (chartInstance) {
        chartInstance.dispose()
    }
    if (compareChartInstance) {
        compareChartInstance.dispose()
    }
})

watch(selectedRecord, () => {
    updateDetailChart()
}, { deep: true })

watch([selectedRecords, compareDialogVisible], () => {
    if (compareDialogVisible.value && compareChartRef.value) {
        if (!compareChartInstance) {
            compareChartInstance = echarts.init(compareChartRef.value)
        }
        updateCompareChart()
    }
}, { deep: true })

defineExpose({
    addRecord
})
</script>

<style scoped>
.optimization-history {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.toolbar {
    padding-bottom: 15px;
    border-bottom: 1px solid #e8e8e8;
}

.records-list {
    flex: 1;
    overflow: auto;
    padding-top: 15px;
}

.empty-state {
    flex: 1;
    display: flex;
    align-items: center;
    justify-content: center;
    color: #909399;
}

.detail-content {
    padding: 10px;
}

.summary-row {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 8px 0;
    border-bottom: 1px dashed #e8e8e8;
}

.summary-row .label {
    font-weight: 500;
    color: #606266;
    width: 100px;
}

.summary-row .value {
    font-weight: 500;
}

.summary-row .value.highlight {
    color: #67c23a;
}

.summary-row .value.success {
    color: #67c23a;
}

.summary-row .value.error {
    color: #f56c6c;
}

.chart-row {
    margin-top: 20px;
}

.chart-row h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.chart-row .chart {
    height: 200px;
}

.tables-row {
    display: flex;
    gap: 20px;
    margin-top: 20px;
}

.detail-table {
    flex: 1;
}

.detail-table h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.compare-content {
    padding: 10px;
}

.compare-chart-row {
    margin-bottom: 20px;
}

.compare-chart-row h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.compare-chart-row .chart {
    height: 250px;
}

.compare-tables-row {
    display: flex;
    gap: 20px;
}

.compare-table {
    flex: 1;
}

.compare-table h4 {
    margin-bottom: 10px;
    font-size: 14px;
    font-weight: 600;
}

.compare-table :deep(.el-table) {
    font-size: 12px;
}

.significant {
    color: #f56c6c;
    font-weight: 600;
}
</style>