<template>
    <div class="mode-selector">
        <el-checkbox v-model="multiSelectMode" @change="handleMultiSelectChange">启用多模态叠加</el-checkbox>
        <div v-if="multiSelectMode" class="mode-list">
            <div v-for="(result, index) in modalResults" :key="index" class="mode-item">
                <el-checkbox v-model="getSelection(index).enabled">
                    第{{ index + 1 }}阶 ({{ result.frequencyHz.toFixed(2) }} Hz)
                </el-checkbox>
                <el-input-number
                    v-if="getSelection(index).enabled"
                    v-model="getSelection(index).scale"
                    :min="0.1"
                    :max="10"
                    :step="0.1"
                    size="small"
                />
            </div>
        </div>
    </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import type { ModalResult, ModeSelection } from '../types'

const props = defineProps<{
    modalResults: ModalResult[]
}>()

const emit = defineEmits<{
    (e: 'update', selections: ModeSelection[]): void
}>()

const multiSelectMode = ref(false)
const selections = ref<ModeSelection[]>([])

const getSelection = (index: number) => {
    while (selections.value.length <= index) {
        selections.value.push({ modeIndex: selections.value.length, enabled: false, scale: 1 })
    }
    return selections.value[index]
}

const handleMultiSelectChange = () => {
    if (!multiSelectMode.value) {
        selections.value.forEach(s => s.enabled = false)
    }
    emit('update', selections.value)
}

const watch = (() => {})
</script>

<style scoped>
.mode-selector {
    padding: 10px;
}

.mode-list {
    margin-top: 10px;
    max-height: 300px;
    overflow-y: auto;
}

.mode-item {
    display: flex;
    align-items: center;
    gap: 10px;
    padding: 5px 0;
    border-bottom: 1px solid #f0f0f0;
}
</style>
