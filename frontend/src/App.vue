<template>
    <div class="app-container">
        <header class="header">
            <h1>结构动力学模态分析与振型可视化工具</h1>
            <el-button @click="showPresetSelector = true" type="primary">加载预设算例</el-button>
        </header>

        <div class="main-content">
            <div class="left-panel">
                <el-card title="结构建模">
                    <StructureForm @update="handleModelUpdate" />
                </el-card>
                <el-card title="分析参数">
                    <el-form :model="analysisParams" label-width="100px">
                        <el-form-item label="模态阶数">
                            <el-input-number v-model="analysisParams.numModes" :min="1" :max="50" />
                        </el-form-item>
                    </el-form>
                    <el-button @click="runAnalysis" type="primary" style="width: 100%;">运行模态分析</el-button>
                </el-card>
            </div>

            <div class="center-panel">
                <el-card title="模态分析结果">
                    <ModalResultTable
                        v-if="modalResults.length > 0"
                        :modalResults="modalResults"
                        :theoryValues="theoryValues"
                        :selectedMode="selectedMode"
                        @select="handleModeSelect"
                    />
                    <div v-else class="empty-state">
                        <p>请先定义结构模型并运行分析</p>
                    </div>
                </el-card>
            </div>

            <div class="right-panel">
                <el-card title="3D可视化" class="viewer-card">
                    <Modal3DViewer
                        :nodes="nodes"
                        :elements="elements"
                        :modalResults="modalResults"
                        :selectedMode="selectedMode"
                        :modeSelections="modeSelections"
                    />
                </el-card>
                <el-card title="模态选择">
                    <ModeSelector :modalResults="modalResults" @update="handleModeSelectionsUpdate" />
                </el-card>
            </div>
        </div>

        <PresetSelector
            :visible="showPresetSelector"
            @close="showPresetSelector = false"
            @select="handlePresetSelect"
        />

        <el-message v-if="message" :type="messageType" :show-close="true">{{ message }}</el-message>
    </div>
</template>

<script setup lang="ts">
import { ref, reactive } from 'vue'
import StructureForm from './components/StructureForm.vue'
import ModalResultTable from './components/ModalResultTable.vue'
import Modal3DViewer from './components/Modal3DViewer.vue'
import ModeSelector from './components/ModeSelector.vue'
import PresetSelector from './components/PresetSelector.vue'
import { api } from './utils/api'
import type { Node, Element, Section, Constraint, ModalResult, TheoryValue, ModeSelection, Preset } from './types'

const nodes = ref<Node[]>([])
const elements = ref<Element[]>([])
const sections = ref<Section[]>([])
const constraints = ref<Constraint[]>([])
const modalResults = ref<ModalResult[]>([])
const theoryValues = ref<TheoryValue[]>([])
const selectedMode = ref(0)
const modeSelections = ref<ModeSelection[]>([])
const showPresetSelector = ref(false)
const message = ref('')
const messageType = ref<'success' | 'warning' | 'error' | 'info'>('success')

const analysisParams = reactive({
    numModes: 10
})

const handleModelUpdate = (data: { nodes: Node[], elements: Element[], sections: Section[], constraints: Constraint[] }) => {
    nodes.value = data.nodes
    elements.value = data.elements
    sections.value = data.sections
    constraints.value = data.constraints
}

const runAnalysis = async () => {
    if (nodes.value.length === 0) {
        showMessage('请先定义节点', 'warning')
        return
    }
    if (elements.value.length === 0) {
        showMessage('请先定义单元', 'warning')
        return
    }
    if (sections.value.length === 0) {
        showMessage('请先定义截面属性', 'warning')
        return
    }
    if (constraints.value.length === 0) {
        showMessage('请先定义约束条件', 'warning')
        return
    }

    try {
        const response = await api.analyze({
            nodes: nodes.value,
            elements: elements.value,
            sections: sections.value,
            constraints: constraints.value,
            numModes: analysisParams.numModes
        })

        if (response.success) {
            modalResults.value = response.modalResults
            theoryValues.value = response.theoryValues || []
            selectedMode.value = 0
            showMessage('分析成功', 'success')
        } else {
            showMessage(response.message, 'error')
        }
    } catch (error) {
        showMessage('分析失败，请检查网络连接', 'error')
        console.error(error)
    }
}

const handleModeSelect = (index: number) => {
    selectedMode.value = index
}

const handleModeSelectionsUpdate = (selections: ModeSelection[]) => {
    modeSelections.value = selections
}

const handlePresetSelect = async (preset: Preset) => {
    showPresetSelector.value = false
    try {
        const response = await api.loadPreset(preset.name)
        if (response.success) {
            nodes.value = response.model.nodes
            elements.value = response.model.elements
            sections.value = response.model.sections
            constraints.value = response.model.constraints
            modalResults.value = response.modalResults
            theoryValues.value = response.theoryValues || []
            selectedMode.value = 0
            showMessage(`已加载预设算例: ${preset.description}`, 'success')
        } else {
            showMessage(response.message, 'error')
        }
    } catch (error) {
        showMessage('加载预设算例失败', 'error')
        console.error(error)
    }
}

const showMessage = (msg: string, type: 'success' | 'warning' | 'error' | 'info') => {
    message.value = msg
    messageType.value = type
    setTimeout(() => {
        message.value = ''
    }, 3000)
}
</script>

<style>
* {
    margin: 0;
    padding: 0;
    box-sizing: border-box;
}

html, body, #app {
    width: 100%;
    height: 100%;
}

.app-container {
    width: 100%;
    height: 100%;
    display: flex;
    flex-direction: column;
    background: #f5f5f5;
}

.header {
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: white;
    padding: 15px 20px;
    display: flex;
    justify-content: space-between;
    align-items: center;
    box-shadow: 0 2px 10px rgba(0, 0, 0, 0.1);
}

.header h1 {
    font-size: 20px;
    font-weight: 600;
}

.main-content {
    flex: 1;
    display: flex;
    padding: 15px;
    gap: 15px;
    overflow: hidden;
}

.left-panel {
    width: 350px;
    display: flex;
    flex-direction: column;
    gap: 15px;
}

.center-panel {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.right-panel {
    width: 450px;
    display: flex;
    flex-direction: column;
    gap: 15px;
}

.el-card {
    flex: 1;
    display: flex;
    flex-direction: column;
    overflow: hidden;
}

.el-card__body {
    flex: 1;
    overflow: auto;
}

.viewer-card {
    flex: 2;
}

.empty-state {
    display: flex;
    align-items: center;
    justify-content: center;
    height: 100%;
    color: #909399;
}
</style>
