<template>
    <div class="modal-result-table">
        <div class="export-buttons" v-if="modalResults.length > 0">
            <el-button @click="exportJSON" type="primary" size="small">
                <Download /> 导出JSON
            </el-button>
            <el-button @click="exportCSV" type="success" size="small">
                <Download /> 导出CSV
            </el-button>
        </div>
        <el-table :data="modalResults" @row-click="handleRowClick" :highlight-current-row="true">
            <el-table-column label="阶数" type="index" width="60" />
            <el-table-column label="频率(Hz)">
                <template #default="scope">
                    <span :class="{ 'rigid-body': scope.row.isRigidBody }">
                        {{ scope.row.isRigidBody ? '刚体模态' : scope.row.frequencyHz.toFixed(4) }}
                    </span>
                </template>
            </el-table-column>
            <el-table-column label="圆频率(rad/s)" prop="circularFreq" :formatter="formatNumber" />
            <el-table-column label="周期(s)" prop="period" :formatter="formatNumber" />
            <el-table-column label="质量参与系数">
                <template #default="scope">
                    <span>X: {{ (scope.row.massParticipation[0] * 100).toFixed(2) }}%</span>
                    <span class="ml-2">Y: {{ (scope.row.massParticipation[1] * 100).toFixed(2) }}%</span>
                    <span class="ml-2">Z: {{ (scope.row.massParticipation[2] * 100).toFixed(2) }}%</span>
                </template>
            </el-table-column>
            <el-table-column label="理论值">
                <template #default="scope">
                    <span v-if="getTheoryValue(scope.$index)">
                        {{ getTheoryValue(scope.$index)?.frequencyHz.toFixed(4) }}
                        <span class="text-danger ml-1">({{ getTheoryValue(scope.$index)?.error.toFixed(2) }}%)</span>
                    </span>
                </template>
            </el-table-column>
        </el-table>
    </div>
</template>

<script setup lang="ts">
import { Download } from '@element-plus/icons-vue'
import type { ModalResult, TheoryValue, Node, Element, Section, Constraint } from '../types'

const props = defineProps<{
    modalResults: ModalResult[]
    theoryValues?: TheoryValue[]
    selectedMode: number
    nodes?: Node[]
    elements?: Element[]
    sections?: Section[]
    constraints?: Constraint[]
}>()

const emit = defineEmits<{
    (e: 'select', modeIndex: number): void
}>()

const handleRowClick = (row: ModalResult, column: any, event: Event) => {
    const index = column.tableData.indexOf(row)
    emit('select', index)
}

const formatNumber = (row: ModalResult, column: any) => {
    return row[column.prop as keyof ModalResult].toFixed(4)
}

const getTheoryValue = (index: number) => {
    return props.theoryValues?.find(t => t.mode === index + 1)
}

const exportJSON = () => {
    const data = {
        model: {
            nodes: props.nodes || [],
            elements: props.elements || [],
            sections: props.sections || [],
            constraints: props.constraints || []
        },
        modalResults: props.modalResults.map(result => ({
            frequencyHz: result.frequencyHz,
            circularFreq: result.circularFreq,
            period: result.period,
            modeShape: result.modeShape,
            massParticipation: result.massParticipation,
            isRigidBody: result.isRigidBody
        })),
        theoryValues: props.theoryValues
    }

    downloadFile(JSON.stringify(data, null, 2), 'modal-analysis-results.json', 'application/json')
}

const exportCSV = () => {
    let csv = '阶数,频率(Hz),圆频率(rad/s),周期(s),质量参与系数X(%),质量参与系数Y(%),质量参与系数Z(%),是否刚体模态\n'
    
    props.modalResults.forEach((result, index) => {
        csv += `${index + 1},`
        csv += result.isRigidBody ? '刚体模态,' : `${result.frequencyHz.toFixed(6)},`
        csv += `${result.circularFreq.toFixed(6)},`
        csv += `${result.period.toFixed(6)},`
        csv += `${(result.massParticipation[0] * 100).toFixed(4)},`
        csv += `${(result.massParticipation[1] * 100).toFixed(4)},`
        csv += `${(result.massParticipation[2] * 100).toFixed(4)},`
        csv += `${result.isRigidBody ? '是' : '否'}\n`
    })

    downloadFile(csv, 'modal-analysis-results.csv', 'text/csv')
}

const downloadFile = (content: string, filename: string, type: string) => {
    const blob = new Blob([content], { type })
    const url = URL.createObjectURL(blob)
    const link = document.createElement('a')
    link.href = url
    link.download = filename
    document.body.appendChild(link)
    link.click()
    document.body.removeChild(link)
    URL.revokeObjectURL(url)
}
</script>

<style scoped>
.modal-result-table {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.export-buttons {
    display: flex;
    gap: 10px;
    margin-bottom: 15px;
}

.rigid-body {
    color: #909399;
    font-style: italic;
}

.ml-2 {
    margin-left: 8px;
}

.text-danger {
    color: #f56c6c;
}
</style>
