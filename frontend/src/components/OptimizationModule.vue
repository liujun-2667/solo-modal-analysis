<template>
    <div class="optimization-module">
        <el-tabs v-model="activeTab" type="border-card">
            <el-tab-pane label="设计变量" name="designVariables">
                <DesignVariables 
                    :sections="sections" 
                    @update="handleDesignVariablesUpdate" 
                />
            </el-tab-pane>
            <el-tab-pane label="灵敏度分析" name="sensitivity">
                <SensitivityAnalysis 
                    :nodes="nodes"
                    :elements="elements"
                    :sections="sections"
                    :constraints="constraints"
                    :designVariables="designVariables"
                    @recommendVariables="handleRecommendVariables"
                />
            </el-tab-pane>
            <el-tab-pane label="频率约束优化" name="optimization">
                <Optimization 
                    :nodes="nodes"
                    :elements="elements"
                    :sections="sections"
                    :constraints="constraints"
                    :designVariables="designVariables"
                    @optimizationComplete="handleOptimizationComplete"
                />
            </el-tab-pane>
            <el-tab-pane label="参数扫描" name="paramScan">
                <ParamScan 
                    :nodes="nodes"
                    :elements="elements"
                    :sections="sections"
                    :constraints="constraints"
                    :designVariables="designVariables"
                />
            </el-tab-pane>
            <el-tab-pane label="优化历史" name="history">
                <OptimizationHistory 
                    :designVarNames="enabledDesignVarNames"
                />
            </el-tab-pane>
        </el-tabs>
    </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import DesignVariables from './DesignVariables.vue'
import SensitivityAnalysis from './SensitivityAnalysis.vue'
import Optimization from './Optimization.vue'
import ParamScan from './ParamScan.vue'
import OptimizationHistory from './OptimizationHistory.vue'
import type { Node, Element, Section, Constraint, DesignVariable, OptimizationResponse } from '../types'

const props = defineProps<{
    nodes: Node[]
    elements: Element[]
    sections: Section[]
    constraints: Constraint[]
}>()

const activeTab = ref('designVariables')
const designVariables = ref<DesignVariable[]>([])

const enabledDesignVarNames = computed(() => {
    return designVariables.value
        .filter(dv => dv.enabled)
        .map(dv => `截面${dv.sectionId}.${dv.property}`)
})

const handleDesignVariablesUpdate = (variables: DesignVariable[]) => {
    designVariables.value = variables
}

const handleRecommendVariables = (variables: DesignVariable[]) => {
    designVariables.value = variables
}

const handleOptimizationComplete = (result: OptimizationResponse) => {
    const historyRecord = {
        initialDesignVariables: result.initialDesignVariables,
        finalDesignVariables: result.finalDesignVariables,
        initialFrequencies: result.initialFrequencies,
        finalFrequencies: result.finalFrequencies,
        initialMass: result.initialMass,
        finalMass: result.finalMass,
        iterations: result.iterations,
        converged: result.converged,
        designVarNames: enabledDesignVarNames.value
    }
    const historyComponent = document.querySelector('.optimization-history')
    if (historyComponent) {
        const event = new CustomEvent('optimizationComplete', { detail: historyRecord })
        historyComponent.dispatchEvent(event)
    }
}
</script>

<style scoped>
.optimization-module {
    height: 100%;
    display: flex;
    flex-direction: column;
}

.optimization-module :deep(.el-tabs__content) {
    flex: 1;
    overflow: auto;
}

.optimization-module :deep(.el-tab-pane) {
    height: 100%;
    display: flex;
    flex-direction: column;
}
</style>