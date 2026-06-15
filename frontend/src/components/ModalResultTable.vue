<template>
    <div class="modal-result-table">
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
import type { ModalResult, TheoryValue } from '../types'

defineProps<{
    modalResults: ModalResult[]
    theoryValues?: TheoryValue[]
    selectedMode: number
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
    const props = defineProps<{
        theoryValues?: TheoryValue[]
    }>()
    return props.theoryValues?.find(t => t.mode === index + 1)
}
</script>

<style scoped>
.modal-result-table {
    height: 100%;
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
