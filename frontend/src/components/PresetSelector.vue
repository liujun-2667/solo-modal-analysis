<template>
    <el-dialog title="选择预设算例" :visible="visible" @close="handleClose">
        <el-table :data="presets" @row-click="handleSelect">
            <el-table-column label="名称" prop="description" />
            <el-table-column>
                <template #default="scope">
                    <el-button @click="handleSelect(scope.row)" type="primary" size="small">加载</el-button>
                </template>
            </el-table-column>
        </el-table>
    </el-dialog>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { api } from '../utils/api'
import type { Preset } from '../types'

const props = defineProps<{
    visible: boolean
}>()

const emit = defineEmits<{
    (e: 'close'): void
    (e: 'select', preset: Preset): void
}>()

const presets = ref<Preset[]>([])

const loadPresets = async () => {
    try {
        presets.value = await api.getPresets()
    } catch (error) {
        console.error('Failed to load presets:', error)
    }
}

const handleSelect = (preset: Preset) => {
    emit('select', preset)
}

const handleClose = () => {
    emit('close')
}

onMounted(() => {
    loadPresets()
})
</script>
