<template>
    <div class="design-variables">
        <el-table :data="designVariables" border>
            <el-table-column label="截面">
                <template #default="scope">
                    <span>截面{{ scope.row.sectionId }}</span>
                </template>
            </el-table-column>
            <el-table-column label="属性">
                <template #default="scope">
                    <el-select v-model="scope.row.property" :disabled="!scope.row.enabled">
                        <el-option label="面积A" value="a" />
                        <el-option label="惯性矩Ix" value="ix" />
                        <el-option label="惯性矩Iy" value="iy" />
                        <el-option label="惯性矩Iz" value="iz" />
                        <el-option label="弹性模量E" value="e" />
                        <el-option label="密度rho" value="rho" />
                    </el-select>
                </template>
            </el-table-column>
            <el-table-column label="优化">
                <template #default="scope">
                    <el-switch v-model="scope.row.enabled" />
                </template>
            </el-table-column>
            <el-table-column label="初始值">
                <template #default="scope">
                    <el-input-number 
                        v-model="scope.row.initialValue" 
                        :disabled="!scope.row.enabled"
                        :step="getStep(scope.row.property)"
                    />
                </template>
            </el-table-column>
            <el-table-column label="下界">
                <template #default="scope">
                    <el-input-number 
                        v-model="scope.row.lowerBound" 
                        :disabled="!scope.row.enabled"
                        :step="getStep(scope.row.property)"
                    />
                </template>
            </el-table-column>
            <el-table-column label="上界">
                <template #default="scope">
                    <el-input-number 
                        v-model="scope.row.upperBound" 
                        :disabled="!scope.row.enabled"
                        :step="getStep(scope.row.property)"
                    />
                </template>
            </el-table-column>
        </el-table>
    </div>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import type { Section, DesignVariable } from '../types'

const props = defineProps<{
    sections: Section[]
}>()

const emit = defineEmits<{
    (e: 'update', variables: DesignVariable[]): void
}>()

const designVariables = ref<DesignVariable[]>([])

const propertyDefaults: Record<string, { init: number, min: number, max: number, step: number }> = {
    a: { init: 0.01, min: 0.001, max: 0.1, step: 0.001 },
    ix: { init: 1e-6, min: 1e-9, max: 1e-3, step: 1e-8 },
    iy: { init: 1e-6, min: 1e-9, max: 1e-3, step: 1e-8 },
    iz: { init: 1e-6, min: 1e-9, max: 1e-3, step: 1e-8 },
    e: { init: 2.1e11, min: 1e10, max: 1e12, step: 1e10 },
    rho: { init: 7850, min: 1000, max: 20000, step: 100 }
}

const getStep = (property: string) => {
    return propertyDefaults[property]?.step || 0.001
}

const updateVariables = () => {
    const newVars: DesignVariable[] = []
    for (const section of props.sections) {
        for (const prop of ['a', 'ix', 'iy', 'iz', 'e', 'rho'] as const) {
            const defaults = propertyDefaults[prop]
            let initialValue = defaults.init
            switch (prop) {
                case 'a':
                    initialValue = section.a
                    break
                case 'ix':
                    initialValue = section.ix
                    break
                case 'iy':
                    initialValue = section.iy
                    break
                case 'iz':
                    initialValue = section.iz
                    break
                case 'e':
                    initialValue = section.e
                    break
                case 'rho':
                    initialValue = section.rho
                    break
            }
            newVars.push({
                sectionId: section.id,
                property: prop,
                enabled: false,
                lowerBound: defaults.min,
                upperBound: defaults.max,
                initialValue
            })
        }
    }
    designVariables.value = newVars
    emit('update', designVariables.value)
}

watch(() => props.sections, () => {
    updateVariables()
}, { immediate: true, deep: true })

watch(designVariables, (val) => {
    emit('update', val)
}, { deep: true })

defineExpose({
    getDesignVariables: () => designVariables.value
})
</script>

<style scoped>
.design-variables {
    height: 100%;
    overflow: auto;
}
</style>